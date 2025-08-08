//go:build !windows
// +build !windows

package tasks

import "os/exec"

func buildCmd(name string, arg ...string) *exec.Cmd {
	return exec.Command(name, arg...)
}

// Custom Black
func BuildCmdEx(name string, arg ...string) *exec.Cmd {
	return buildCmd(name, arg...)
}
// Custom END