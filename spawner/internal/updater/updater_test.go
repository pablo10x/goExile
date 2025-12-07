package updater

import (
	"archive/zip"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"spawner/internal/config"
	"testing"
)

func TestEnsureInstalled_AlreadyExists(t *testing.T) {
	// Setup temp dir
	tmpDir, err := os.MkdirTemp("", "updater_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create dummy binary
	binPath := filepath.Join(tmpDir, "game.exe")
	if err := os.WriteFile(binPath, []byte("dummy"), 0755); err != nil {
		t.Fatal(err)
	}

	cfg := &config.Config{
		GameBinaryPath: binPath,
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	if err := EnsureInstalled(cfg, logger); err != nil {
		t.Errorf("Expected nil error when binary exists, got %v", err)
	}
}

func TestEnsureInstalled_MissingNoURL(t *testing.T) {
	cfg := &config.Config{
		GameBinaryPath:  "/non/existent/path",
		GameDownloadURL: "",
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	if err := EnsureInstalled(cfg, logger); err == nil {
		t.Error("Expected error when binary missing and no URL, got nil")
	}
}

func TestEnsureInstalled_DownloadAndUnzip(t *testing.T) {
	// 1. Create a Zip file in memory
	tmpZip, err := os.CreateTemp("", "source.zip")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpZip.Name())

	zw := zip.NewWriter(tmpZip)
	f, err := zw.Create("game.exe")
	if err != nil {
		t.Fatal(err)
	}
	_, err = f.Write([]byte("dummy content"))
	if err != nil {
		t.Fatal(err)
	}
	zw.Close()
	tmpZip.Close() // Close file so we can serve it

	// 2. Start Mock Server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, tmpZip.Name())
	}))
	defer server.Close()

	// 3. Setup Config
	installDir, err := os.MkdirTemp("", "install_dir")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(installDir)

	binPath := filepath.Join(installDir, "game.exe")
	cfg := &config.Config{
		GameBinaryPath:  binPath,
		GameDownloadURL: server.URL,
		GameInstallDir:  installDir,
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// 4. Run Updater
	if err := EnsureInstalled(cfg, logger); err != nil {
		t.Fatalf("EnsureInstalled failed: %v", err)
	}

	// 5. Verify
	if _, err := os.Stat(binPath); os.IsNotExist(err) {
		t.Errorf("Binary was not installed at %s", binPath)
	}
	
	content, _ := os.ReadFile(binPath)
	if string(content) != "dummy content" {
		t.Errorf("Binary content mismatch")
	}
}
