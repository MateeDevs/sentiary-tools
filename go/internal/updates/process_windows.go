package updates

import (
	"os/exec"
	"syscall"
)

func SpawnBackground(executable string) error {
	command := exec.Command(executable, BackgroundArgument)
	command.SysProcAttr = &syscall.SysProcAttr{CreationFlags: 0x00000008 | 0x00000200, HideWindow: true}
	if err := command.Start(); err != nil {
		return err
	}
	return command.Process.Release()
}
