package updater

import (
	"archive/zip"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"spawner/internal/config"
	"strings"
)

// EnsureInstalled checks if the game binary exists. If not, it attempts to
// download the server package from the Master Server.
func EnsureInstalled(cfg *config.Config, logger *slog.Logger) error {
	// 1. Check if binary already exists
	if _, err := os.Stat(cfg.GameBinaryPath); err == nil {
		logger.Info("Game binary found, skipping installation", "path", cfg.GameBinaryPath)
		return nil
	}

	logger.Info("Game binary not found. Attempting download from Master Server...", "master_url", cfg.MasterURL)

	// 2. Ensure install directory exists
	if err := os.MkdirAll(cfg.GameInstallDir, 0755); err != nil {
		return fmt.Errorf("failed to create install directory: %w", err)
	}

	// 3. Download the zip file from Master Server
	downloadURL := fmt.Sprintf("%s/api/spawners/download?os=%s", cfg.MasterURL, runtime.GOOS)
	tmpFile, err := os.CreateTemp("", "gameserver-*.zip")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	logger.Info("Downloading game server package...", "url", downloadURL)
	if err := downloadFile(downloadURL, cfg.MasterAPIKey, tmpFile); err != nil {
		tmpFile.Close()
		return fmt.Errorf("download failed: %w", err)
	}
	tmpFile.Close()

	// 4. Unzip
	logger.Info("Extracting package...", "destination", cfg.GameInstallDir)
	if err := unzip(tmpFile.Name(), cfg.GameInstallDir); err != nil {
		return fmt.Errorf("extraction failed: %w", err)
	}

	// 5. Verify installation
	if _, err := os.Stat(cfg.GameBinaryPath); err != nil {
		return fmt.Errorf("installation completed but binary still missing at %s", cfg.GameBinaryPath)
	}

	// 6. Make executable
	_ = os.Chmod(cfg.GameBinaryPath, 0755)

	logger.Info("Game server downloaded and installed successfully")
	return nil
}

func downloadFile(url, apiKey string, dest *os.File) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("User-Agent", "Spawner/1.0")
	if apiKey != "" {
		req.Header.Set("X-API-Key", apiKey)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned status: %s %s", resp.Status,resp.Body)
	}

	_, err = io.Copy(dest, resp.Body)
	return err
}

func unzip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		fpath := filepath.Join(dest, f.Name)
		if !strings.HasPrefix(fpath, filepath.Clean(dest)+string(os.PathSeparator)) {
			return fmt.Errorf("illegal file path: %s", fpath)
		}

		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return err
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return err
		}

		_, err = io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()

		if err != nil {
			return err
		}
	}
	return nil
}