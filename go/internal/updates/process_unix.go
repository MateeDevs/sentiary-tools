//go:build !windows

package updates

import (
	"os/exec"
	"syscall"
)

func SpawnBackground(executable string) error {
	command := exec.Command(executable, BackgroundArgument)
	command.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	if err := command.Start(); err != nil {
		return err
	}
	return command.Process.Release()
}
