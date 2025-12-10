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
	fullBinaryPath := filepath.Join(cfg.GameInstallDir, cfg.GameBinaryPath)

	// 1. Check if binary already exists
	if _, err := os.Stat(fullBinaryPath); err == nil {
		logger.Info("Game binary found, skipping installation", "path", fullBinaryPath)
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
	version, err := downloadFile(downloadURL, cfg.MasterAPIKey, tmpFile)
	if err != nil {
		tmpFile.Close()
		return fmt.Errorf("download failed: %w", err)
	}
	tmpFile.Close()

	// Save version info if available
	if version != "" {
		versionFile := filepath.Join(cfg.GameInstallDir, "version.txt")
		// We write to a temp location first or just overwrite after unzip? 
		// Unzip happens next. If we write now, unzip might overwrite or be fine.
		// Actually, let's write it AFTER unzip to ensure it persists.
		defer func() {
			if err := os.WriteFile(versionFile, []byte(version), 0644); err != nil {
				logger.Warn("Failed to save version file", "error", err)
			}
		}()
	}

	// 4. Unzip
	logger.Info("Extracting package...", "destination", cfg.GameInstallDir)
	if err := unzip(tmpFile.Name(), cfg.GameInstallDir); err != nil {
		return fmt.Errorf("extraction failed: %w", err)
	}

	// 5. Verify installation
	if _, err := os.Stat(fullBinaryPath); err != nil {
		// Attempt to auto-detect the binary if the path is wrong (e.g., missing or extra directory)
		targetName := filepath.Base(cfg.GameBinaryPath)
		var foundPath string

		_ = filepath.Walk(cfg.GameInstallDir, func(path string, info os.FileInfo, err error) error {
			if err == nil && !info.IsDir() && strings.EqualFold(info.Name(), targetName) {
				foundPath = path
				return io.EOF // Stop walking
			}
			return nil
		})

		if foundPath != "" {
			rel, _ := filepath.Rel(cfg.GameInstallDir, foundPath)
			logger.Warn("Game binary not found at configured path, but found elsewhere. Auto-correcting configuration.", 
				"configured", cfg.GameBinaryPath, 
				"found", rel)
			
			// Update config dynamically
			cfg.GameBinaryPath = rel
			fullBinaryPath = foundPath // Update local var for chmod below
		} else {
			// Debug: list files to help diagnose structure issues
			logger.Error("Binary verification failed. Listing installed files to debug structure:", "root", cfg.GameInstallDir)
			_ = filepath.Walk(cfg.GameInstallDir, func(path string, info os.FileInfo, err error) error {
				if err == nil {
					// Show relative path
					rel, _ := filepath.Rel(cfg.GameInstallDir, path)
					if rel != "." {
						logger.Error("Found file", "path", rel)
					}
				}
				return nil
			})
			return fmt.Errorf("installation completed but binary still missing at %s", fullBinaryPath)
		}
	}

	// 6. Make executable
	_ = os.Chmod(fullBinaryPath, 0755)

	logger.Info("Game server downloaded and installed successfully", "version", version)
	return nil
}

// UpdateTemplate checks if a newer version is available on the master server and updates the local template.
func UpdateTemplate(cfg *config.Config, logger *slog.Logger) (string, error) {
	// 1. Check local version
	versionFile := filepath.Join(cfg.GameInstallDir, "version.txt")
	localVersionBytes, err := os.ReadFile(versionFile)
	localVersion := ""
	if err == nil {
		localVersion = strings.TrimSpace(string(localVersionBytes))
	}

	downloadURL := fmt.Sprintf("%s/api/spawners/download?os=%s", cfg.MasterURL, runtime.GOOS)
	
	// 2. Check remote version (HEAD request)
	req, err := http.NewRequest("HEAD", downloadURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	if cfg.MasterAPIKey != "" {
		req.Header.Set("X-API-Key", cfg.MasterAPIKey)
	}
	
	client := &http.Client{Timeout: 5 * http.DefaultClient.Timeout} // Short timeout for check
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to check update: %w", err)
	}
	resp.Body.Close()
	
	remoteVersion := resp.Header.Get("X-Game-Version")
	if remoteVersion == "" {
		logger.Warn("Master server did not return X-Game-Version header. Skipping update check.", "current_local", localVersion)
		return localVersion, nil
	}

	if localVersion == remoteVersion {
		logger.Info("Local template is up to date", "version", localVersion)
		return localVersion, nil
	}

	logger.Info("Found new version, updating template...", "local", localVersion, "remote", remoteVersion)

	// 3. Download and Extract
	tmpFile, err := os.CreateTemp("", "gameserver-update-*.zip")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := downloadFile(downloadURL, cfg.MasterAPIKey, tmpFile); err != nil {
		tmpFile.Close()
		return "", fmt.Errorf("download failed: %w", err)
	}
	tmpFile.Close()

	if err := unzip(tmpFile.Name(), cfg.GameInstallDir); err != nil {
		return "", fmt.Errorf("extraction failed: %w", err)
	}

	// 4. Update version file
	if err := os.WriteFile(versionFile, []byte(remoteVersion), 0644); err != nil {
		logger.Warn("Failed to save version file", "error", err)
	}

	// 5. Re-apply permissions
	fullBinaryPath := filepath.Join(cfg.GameInstallDir, cfg.GameBinaryPath)
	_ = os.Chmod(fullBinaryPath, 0755)

	logger.Info("Template updated successfully", "version", remoteVersion)
	return remoteVersion, nil
}

func downloadFile(url, apiKey string, dest *os.File) (string, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("User-Agent", "Spawner/1.0")
	if apiKey != "" {
		req.Header.Set("X-API-Key", apiKey)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("server returned status: %s", resp.Status)
	}

	_, err = io.Copy(dest, resp.Body)
	return resp.Header.Get("X-Game-Version"), err
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