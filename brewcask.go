package main

import (
	"os/exec"
	"strings"
)

type brewcaskPackageManager struct {
	curentList []string
}

func (p *brewcaskPackageManager) Name() PackageManagerName {
	return "brewcask"
}

func (p *brewcaskPackageManager) Initialize() error {
	cmd := exec.Command("brew", "cask", "list", "--full-name")

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

func (p *brewcaskPackageManager) AddInstalledPackage(pd LoadedPackageData) {
	p.curentList = append(p.curentList, pd.BrewPackageName())
}

func (p *brewcaskPackageManager) Inspect(pd LoadedPackageData) (ExecutablePackageCmd, error) {
	bep := &brewcaskExecutablePackage{}

	bep.data = pd
	bep.bp = p

	bep.installCmd = append(
		bep.installCmd,
		exec.Command("brew", "cask", "install", pd.BrewPackageName()),
	)

	bep.updateCmd = append(
		bep.updateCmd,
		exec.Command("brew", "cask", "upgrade", pd.BrewPackageName()),
	)

	return bep, nil
}

type brewcaskExecutablePackage struct {
	installCmd []*exec.Cmd
	updateCmd  []*exec.Cmd
	data       LoadedPackageData
	bp         *brewcaskPackageManager
}

func (e *brewcaskExecutablePackage) LoadedPackageData() LoadedPackageData {
	return e.data

}

func (e *brewcaskExecutablePackage) IsAlreadyInstalled() bool {
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

func (e *brewcaskExecutablePackage) UpdateCommand() Command {
	cmd := GeneralCommand{
		cmds:        e.updateCmd,
		description: e.data.Describe(),
	}

	return cmd
}

func (e *brewcaskExecutablePackage) InstallCommand() Command {
	cmd := GeneralCommand{
		cmds:        e.installCmd,
		description: e.data.Describe(),
		completeCallBack: func() {
			e.bp.AddInstalledPackage(e.data)
		},
	}

	return cmd
}
