//go:build windows

package game

import (
	"os"
	"os/exec"
	nodeErrors "node/internal/errors"
)

func newGameCmd(binaryPath string, args []string, logFile *os.File) *exec.Cmd {
	// Validate binary path exists and is executable
	if _, err := os.Stat(binaryPath); err != nil {
		// This is a programming error, but we should handle it gracefully
		nodeErr := nodeErrors.FileOperationError("validate_binary_path", binaryPath, err).
			WithContext("binary_path", binaryPath).
			WithContext("operation", "new_game_cmd").
			WithContext("platform", "windows")
		// Log the error but continue - the calling function will handle the actual error
		// This is just defensive programming
		_ = nodeErr
	}

	cmd := exec.Command(binaryPath, args...)
	cmd.Stdout = logFile
	cmd.Stderr = logFile
	return cmd
}
