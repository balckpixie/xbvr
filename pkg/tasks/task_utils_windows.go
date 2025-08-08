//go:build windows
// +build windows

package tasks

import (
	"os/exec"
	"syscall"
)

func buildCmd(name string, arg ...string) *exec.Cmd {
	cmd := exec.Command(name, arg...)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	return cmd
}

// Custom Black
func BuildCmdEx(name string, arg ...string) *exec.Cmd {
	return buildCmd(name, arg...)
}
// Custom END
