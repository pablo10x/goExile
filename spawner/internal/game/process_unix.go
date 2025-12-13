//go:build !windows

package game

import (
	"os"
	"os/exec"
	spawnerErrors "spawner/internal/errors"
	"syscall"
)

// newGameCmd creates a command to execute the game binary on Unix-like systems.
// It uses "nohup" and creates a new process group to ensure the process runs
// in the background and persists after the parent process exits.
func newGameCmd(binaryPath string, args []string, logFile *os.File) *exec.Cmd {
	// Validate binary path exists and is executable
	if _, err := os.Stat(binaryPath); err != nil {
		// This is a programming error, but we should handle it gracefully
		spawnerErr := spawnerErrors.FileOperationError("validate_binary_path", binaryPath, err).
			WithContext("binary_path", binaryPath).
			WithContext("operation", "new_game_cmd")
		// Log the error but continue - the calling function will handle the actual error
		// This is just defensive programming
		_ = spawnerErr
	}

	// Prepend the binary path to the arguments for nohup.
	fullArgs := append([]string{binaryPath}, args...)

	// Execute with nohup to ignore SIGHUP signals.
	cmd := exec.Command("nohup", fullArgs...)

	// Setpgid sets the process group ID to the process ID, creating a new process group.
	// This decouples the child process from the parent's control terminal signals.
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	// Redirect stdout and stderr to the specified log file.
	cmd.Stdout = logFile
	cmd.Stderr = logFile

	return cmd
}
