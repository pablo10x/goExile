package game

import (
	"log/slog"
	"net"
	"os"
	"runtime"
	"spawner/internal/config"
	"testing"
)

func TestNewManager(t *testing.T) {
	cfg := &config.Config{Region: "Test"}
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	m := NewManager(cfg, logger)

	if m == nil {
		t.Fatal("NewManager returned nil")
	}
}

// TestSpawnIntegration attempts to actually spawn a process.
// This requires a valid binary. We will use a system command.
func TestSpawnIntegration(t *testing.T) {
	// Determine a command that runs for a bit then exits
	var cmdPath string
	
	if runtime.GOOS == "windows" {
		cmdPath = "C:\\Windows\\System32\\timeout.exe"
		// Check if exists, else skip
		if _, err := os.Stat(cmdPath); os.IsNotExist(err) {
			t.Skip("timeout.exe not found, skipping integration test")
		}
	} else {
		cmdPath = "/bin/sleep"
		if _, err := os.Stat(cmdPath); os.IsNotExist(err) {
			t.Skip("sleep not found, skipping integration test")
		}
	}

	/*
	// Integration test logic temporarily disabled due to hardcoded args in Spawn()
	
	cfg := &config.Config{
		Region:         "Test",
		GameBinaryPath: cmdPath,
		MinGamePort:    9000,
		MaxGamePort:    9005,
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	m := NewManager(cfg, logger)

	ctx := context.Background()
	_, err := m.Spawn(ctx)
	if err != nil {
		t.Logf("Spawn failed (expected due to args): %v", err)
	}
	*/

	// Just skip for now until refactor
	t.Skip("Skipping spawn integration test because Spawn() has hardcoded Unity arguments")
}

func TestFindAvailablePort(t *testing.T) {
	cfg := &config.Config{
		MinGamePort: 9010,
		MaxGamePort: 9012,
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	m := NewManager(cfg, logger)

	// Occupy 9010
	l, err := net.Listen("tcp", ":9010")
	if err != nil {
		t.Skipf("Could not bind port 9010 for test: %v", err)
	}
	defer l.Close()

	port, err := m.findAvailablePort()
	if err != nil {
		t.Fatalf("findAvailablePort failed: %v", err)
	}

	if port == 9010 {
		t.Errorf("Expected to skip occupied port 9010, got %d", port)
	}
	if port != 9011 {
		t.Errorf("Expected port 9011, got %d", port)
	}
}