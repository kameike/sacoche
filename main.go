package main

import (
	"flag"
	"fmt"
	"log"
	"os/exec"
	"strings"

	"github.com/BurntSushi/toml"
)

func main() {
	var ppm PackageManager = &phonyPackageManager{}
	var bpm PackageManager = &brewPackageManager{}
	var bcpm PackageManager = &brewcaskPackageManager{}
	var gopm PackageManager = &goPackageManager{}
	bpm.Initialize()
	bcpm.Initialize()
	gopm.Initialize()

	flag.Parse()
	filename := flag.Arg(0)

	if filename == "" {
		fmt.Println("need filename")
		return
	}

	list := map[string]LoadedPackageData{}
	_, e := toml.DecodeFile(filename, &list)
	if e != nil {
		panic(e.Error())
	}

	cmds := []ExecutablePackageCmd{}

	for k, v := range list {
		v.Name = PackageName(k)

		for _, mgr := range v.PackageManager {

			if mgr == "brew" {
				res, _ := bpm.Inspect(v)
				cmds = append(cmds, res)
			}

			if mgr == "phony" {
				res, _ := ppm.Inspect(v)
				cmds = append(cmds, res)
			}

			if mgr == "brewcask" {
				res, _ := bcpm.Inspect(v)
				cmds = append(cmds, res)
			}

			if mgr == "go" {
				res, _ := gopm.Inspect(v)
				cmds = append(cmds, res)
			}
		}
	}

	cmds = resolveDependency(cmds)

	subcmd := flag.Arg(1)
	if subcmd == "" || subcmd == "install" {
		execInstallCmds(cmds)
	} else if subcmd == "update" {
		execUpdateCmds(cmds)
	}
}

func execUpdateCmds(cmds []ExecutablePackageCmd) {
	for _, v := range cmds {
		name := v.LoadedPackageData().Name
		fmt.Printf("[info] updating %s\n", name)
		e := v.UpdateCommand().Execute()
		if e != nil {
			fmt.Printf("[fail] err: %s\n", e.Error())
			fmt.Printf("[fail] faled to update %s\n", name)
		} else {
			fmt.Printf("[ ok ] %s has been updated\n", name)
		}
	}
}

func execInstallCmds(cmds []ExecutablePackageCmd) {
	for _, v := range cmds {
		name := v.LoadedPackageData().Name
		if !v.IsAlreadyInstalled() {
			fmt.Printf("[info] %s is not installed try to install\n", name)
			e := v.InstallCommand().Execute()
			if e != nil {
				fmt.Printf("[fail] err: %s\n", e.Error())
				fmt.Printf("[fail] faled to install %s\n", name)
			} else {
				fmt.Printf("[ ok ] %s has been installed\n", name)
			}
		} else {
			fmt.Printf("[ ok ] %s is installed\n", name)
		}
	}
}

func resolveDependency(origin []ExecutablePackageCmd) (result []ExecutablePackageCmd) {

	// https://ja.wikipedia.org/wiki/トポロジカルソート
	resulved := map[PackageName]bool{}
	for len(origin) != 0 {
		hit := false

		for i, v := range origin {
			isOK := true
			for _, d := range v.LoadedPackageData().Dependency {
				_, ok := resulved[d]
				isOK = isOK && ok
			}
			if isOK {
				target := origin[i]
				if i == len(origin)-1 {
					origin = origin[0 : len(origin)-1]
				} else {
					origin = append(origin[0:i], origin[i+1:]...)
				}
				result = append(result, target)
				resulved[v.LoadedPackageData().Name] = true
				hit = true
				break
			}
		}

		if hit == false {
			for _, v := range origin {
				fmt.Printf("[fail] fail to resolve dependency of %s that needs", string(v.LoadedPackageData().Name))
				for _, d := range v.LoadedPackageData().Dependency {
					fmt.Printf(" '%s'", string(d))
				}
				fmt.Printf("\n")
			}
			break
		}
	}

	return result
}

var Logger *log.Logger

func logf(format string, v ...interface{}) {
	if Logger == nil {
		log.Printf(format, v...)
		return
	}
	Logger.Printf(format, v...)
}

func LogData(c ExecutablePackageCmd) {
	if c.IsAlreadyInstalled() {
		logf("%s is alreadyt installed", string(c.LoadedPackageData().Name))
		return
	}
	logf(c.InstallCommand().Describe())
	logf(c.UpdateCommand().Describe())
}

type PackageName string
type GroupName string

type LoadedPackageData struct {
	Name PackageName

	BrewName     string
	BrewTap      string
	CheckCommand string

	Group          []GroupName
	PackageManager []PackageManagerName `toml:"manager"`
	Dependency     []PackageName
	Tasks          []string

	GoRepo string
}

func (l LoadedPackageData) Describe() (res string) {
	res = fmt.Sprintf("Package[%s]", l.Name)

	return res
}

func (l LoadedPackageData) BrewPackageName() string {
	if l.BrewName != "" {
		return l.BrewName
	}
	return string(l.Name)
}

type OS interface {
	Name() string
	Initialize()
	Managers() ([]PackageManager, error)
}

type Command interface {
	Execute() error
	Describe() string
}

type GeneralCommand struct {
	description      string
	cmds             []*exec.Cmd
	completeCallBack func()
}

func (g GeneralCommand) Execute() error {
	for _, c := range g.cmds {
		fmt.Printf("[info] run `%s`\n", c.String())
		e := c.Run()
		if e != nil {
			return e
		}
	}
	if g.completeCallBack != nil {
		g.completeCallBack()
	}
	return nil
}

func (g GeneralCommand) Describe() (res string) {
	for _, c := range strings.Split(g.description, "\n") {
		res += fmt.Sprintf("# %s\n", c)
	}

	for _, c := range g.cmds {
		res += fmt.Sprintf("[cmd]%s\n", c.String())
	}

	return res
}

type ExecutablePackageCmd interface {
	LoadedPackageData() LoadedPackageData
	IsAlreadyInstalled() bool
	UpdateCommand() Command
	InstallCommand() Command
}

type PackageManagerName string
type PackageManager interface {
	Name() PackageManagerName
	Initialize() error
	Inspect(LoadedPackageData) (ExecutablePackageCmd, error)
}
