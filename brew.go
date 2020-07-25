package main

import (
	"os/exec"
	"strings"
)

type brewPackageManager struct {
	curentList []string
}

func (p *brewPackageManager) Name() PackageManagerName {
	return "brew"
}

func (p *brewPackageManager) Initialize() error {
	cmd := exec.Command("brew", "list", "--full-name")

	buf, err := cmd.Output()
	if err != nil {
		return err
	}

	p.curentList = []string{}

	for _, b := range strings.Split(string(buf), "\n") {
		p.curentList = append(p.curentList, b)
	}

	return nil
}

func (p *brewPackageManager) AddInstalledPackage(pd LoadedPackageData) {
	p.curentList = append(p.curentList, pd.BrewPackageName())
}

func (p *brewPackageManager) Inspect(pd LoadedPackageData) (ExecutablePackageCmd, error) {
	bep := &brewExecutablePackage{}

	bep.data = pd
	bep.bp = p

	if pd.BrewTap != "" {
		cmds := exec.Command("brew", "tap", pd.BrewTap)
		bep.installCmd = append(bep.installCmd, cmds)
	}

	bep.installCmd = append(
		bep.installCmd,
		exec.Command("brew", "install", pd.BrewPackageName()),
	)

	bep.updateCmd = append(
		bep.updateCmd,
		exec.Command("brew", "upgrade", pd.BrewPackageName()),
	)

	return bep, nil
}

type brewExecutablePackage struct {
	installCmd []*exec.Cmd
	updateCmd  []*exec.Cmd
	data       LoadedPackageData
	bp         *brewPackageManager
}

func (e *brewExecutablePackage) LoadedPackageData() LoadedPackageData {
	return e.data

}

func (e *brewExecutablePackage) IsAlreadyInstalled() bool {
	for _, v := range e.bp.curentList {
		if v == "" {
			continue
		}
		if strings.Contains(v, e.data.BrewPackageName()) {
			return true
		}
	}
	return false
}

func (e *brewExecutablePackage) UpdateCommand() Command {
	cmd := GeneralCommand{
		cmds:        e.updateCmd,
		description: e.data.Describe(),
	}

	return cmd
}

func (e *brewExecutablePackage) InstallCommand() Command {
	cmd := GeneralCommand{
		cmds:        e.installCmd,
		description: e.data.Describe(),
		completeCallBack: func() {
			e.bp.AddInstalledPackage(e.data)
		},
	}

	return cmd
}
