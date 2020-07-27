package main

import (
	"os/exec"
	"strings"
)

type phonyPackageManager struct{}

func (p *phonyPackageManager) Name() PackageManagerName {
	return "phony"
}

func (p *phonyPackageManager) Initialize() error {
	return nil
}

func (p *phonyPackageManager) Inspect(pd LoadedPackageData) (ExecutablePackageCmd, error) {
	pep := &phonyExecutablePackage{}

	pep.data = pd

	for _, t := range pd.Tasks {
		ss := strings.Split(t, " ")
		cmds := exec.Command(ss[0], ss[1:]...)
		pep.command = append(pep.command, cmds)
	}

	return pep, nil
}

type phonyExecutablePackage struct {
	command []*exec.Cmd
	data    LoadedPackageData
}

func (e *phonyExecutablePackage) LoadedPackageData() LoadedPackageData {
	return e.data

}
func (e *phonyExecutablePackage) IsAlreadyInstalled() bool {
	cmd := strings.Split(e.LoadedPackageData().CheckCommand, " ")
	if len(cmd) == 1 {
		return false
	}

	ex := exec.Command(cmd[0], cmd[:1]...)
	err := ex.Run()
	if err != nil {
		println(err.Error())
		return false
	}

	return true
}
func (e *phonyExecutablePackage) UpdateCommand() Command {
	cmd := GeneralCommand{
		cmds:        e.command,
		description: e.data.Describe(),
	}

	return cmd
}
func (e *phonyExecutablePackage) InstallCommand() Command {
	cmd := GeneralCommand{
		cmds:        e.command,
		description: e.data.Describe(),
	}

	return cmd
}
