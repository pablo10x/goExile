//go:build windows

package game

import (
	"os"
	"os/exec"
)

func newGameCmd(binaryPath string, args []string, logFile *os.File) *exec.Cmd {
	cmd := exec.Command(binaryPath, args...)
	cmd.Stdout = logFile
	cmd.Stderr = logFile
	return cmd
}
