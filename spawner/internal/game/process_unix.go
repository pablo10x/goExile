//go:build !windows

package game

import (
	"os"
	"os/exec"
	"syscall"
)

func newGameCmd(binaryPath string, args []string, logFile *os.File) *exec.Cmd {
	// Prepend binary path for nohup
	fullArgs := append([]string{binaryPath}, args...)
	cmd := exec.Command("nohup", fullArgs...)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Stdout = logFile
	cmd.Stderr = logFile
	return cmd
}
