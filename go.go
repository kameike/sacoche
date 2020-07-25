package main

import (
	"os/exec"
)

type goPackageManager struct{}

func (p *goPackageManager) Name() PackageManagerName {
	return "go"
}

func (p *goPackageManager) Initialize() error {
	return nil
}

func (p *goPackageManager) Inspect(pd LoadedPackageData) (ExecutablePackageCmd, error) {
	bep := &goExecutablePackage{}

	bep.data = pd
	bep.bp = p

	bep.installCmd = append(
		bep.installCmd,
		exec.Command("go", "get", pd.GoRepo),
	)

	bep.updateCmd = append(
		bep.updateCmd,
		exec.Command("go", "get", "-u", pd.GoRepo),
	)

	return bep, nil
}

type goExecutablePackage struct {
	installCmd []*exec.Cmd
	updateCmd  []*exec.Cmd
	data       LoadedPackageData
	bp         *goPackageManager
}

func (e *goExecutablePackage) LoadedPackageData() LoadedPackageData {
	return e.data

}

func (e *goExecutablePackage) IsAlreadyInstalled() bool {
	cmd := exec.Command("which", string(e.LoadedPackageData().Name))
	err := cmd.Run()
	if err == nil {
		return true
	} else {
		return false
	}
}

func (e *goExecutablePackage) UpdateCommand() Command {
	cmd := GeneralCommand{
		cmds:        e.updateCmd,
		description: e.data.Describe(),
	}

	return cmd
}

func (e *goExecutablePackage) InstallCommand() Command {
	cmd := GeneralCommand{
		cmds:        e.installCmd,
		description: e.data.Describe(),
	}

	return cmd
}
