// Package config provides configuration loading and management for the node.
package config

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

// Config holds the configuration parameters for the node.
type Config struct {
	Region            string
	Host              string // Hostname/IP reachable by Master Server
	GameBinaryPath    string
	GameDownloadURL   string
	GameDownloadToken string
	GameInstallDir    string
	Port              string // Port for the Node API itself
	StartingPort      int    // Starting port for game instances
	MaxInstances      int    // Maximum number of instances allowed
	MasterURL         string // URL of the Master Server (Registry)
	MasterAPIKey      string // API Key for Master Server authentication
	StateFilePath     string // Path to the JSON file storing instance state
	InstancesDir      string // Directory where game server instances are spawned
	EnrollmentKey     string // One-time enrollment key for initial registration
	IsDraining        bool   // Drain mode status
}

// Package-level flag variables
var (
	keyFlag  = flag.String("key", "", "One-time enrollment key for initial registration with master server")
	urlFlag  = flag.String("url", "", "URL of the Master Server")
	mFlag    = flag.String("m", "", "Alias for -url (URL of the Master Server)")
	hostFlag = flag.String("host", "localhost", "Hostname/IP reachable by Master Server")
	portFlag = flag.String("port", "8080", "Port for the Node API itself")

	// These are now managed by Master but kept for backward compatibility or local-only mode
	spFlag     = flag.Int("sp", 0, "Starting port for game instances (Managed by Master)")
	maxFlag    = flag.Int("max", 0, "Maximum number of instances (Managed by Master)")
	regionFlag = flag.String("region", "", "Region identifier (e.g. US-East) (Managed by Master)")

	gameBinaryPathFlag    = flag.String("game-binary", "", "Path to the game server binary relative to install dir")
	gameDownloadURLFlag   = flag.String("game-download-url", "", "URL to download game server files (if different from Master)")
	gameDownloadTokenFlag = flag.String("game-download-token", "", "Token for game server download (if required)")
	gameInstallDirFlag    = flag.String("install-dir", "", "Directory where game server is installed")
	stateFilePathFlag     = flag.String("state-file", "instances.json", "Path to the JSON file storing instance state")
	instancesDirFlag      = flag.String("instances-dir", "", "Directory where game server instances are spawned")
)

// Load reads configuration from environment variables and command-line flags.
func Load() (*Config, error) {
	// Try loading .env from local directory only
	_ = godotenv.Load(".env")

	// Determine Master URL (mFlag > urlFlag > env)
	masterURL := getEnv("MASTER_URL", *urlFlag)
	if *mFlag != "" {
		masterURL = *mFlag
	}

	conf := &Config{
		Region:            getEnv("REGION", *regionFlag),
		Host:              *hostFlag,
		GameDownloadURL:   *gameDownloadURLFlag,
		Port:              *portFlag,
		StartingPort:      *spFlag,
		MaxInstances:      *maxFlag,
		StateFilePath:     *stateFilePathFlag,
		EnrollmentKey:     *keyFlag,
		MasterURL:         masterURL,
		MasterAPIKey:      getEnv("MASTER_API_KEY", ""),
		GameBinaryPath:    getEnv("GAME_BINARY_PATH", *gameBinaryPathFlag),
		GameInstallDir:    getEnv("GAME_INSTALL_DIR", *gameInstallDirFlag),
		InstancesDir:      getEnv("INSTANCES_DIR", *instancesDirFlag),
		GameDownloadToken: getEnv("GAME_DOWNLOAD_TOKEN", *gameDownloadTokenFlag),
		IsDraining:        getEnv("IS_DRAINING", "false") == "true",
	}

	// Set defaults if not provided
	if conf.StartingPort == 0 {
		conf.StartingPort = 7777
	}
	if conf.MaxInstances == 0 {
		conf.MaxInstances = 10
	}

	return conf, nil
}

// Validate checks if the configuration is valid and sufficient to start the application.
func (c *Config) Validate() error {
	// If we have an enrollment key, we don't need Region/MaxInstances yet as they will be provided by Master
	if c.EnrollmentKey == "" {
		if c.Region == "" {
			return fmt.Errorf("REGION is required (use -region flag or register via dashboard)")
		}
		if c.MaxInstances <= 0 {
			return fmt.Errorf("MAX_INSTANCES must be > 0")
		}
	}

	if c.GameBinaryPath == "" {
		return fmt.Errorf("GAME_BINARY_PATH is required (use -game-binary flag or env)")
	}

	// If no API key is configured, an enrollment key is required for initial setup
	if c.MasterAPIKey == "" && c.EnrollmentKey == "" {
		return fmt.Errorf("either MASTER_API_KEY or -key (enrollment key) is required")
	}

	// Check if binary exists in the install directory
	fullBinaryPath := filepath.Join(c.GameInstallDir, c.GameBinaryPath)
	if _, err := os.Stat(fullBinaryPath); os.IsNotExist(err) {
		// If binary is missing, we MUST have a master URL to recover (download)
		if c.MasterURL == "" {
			return fmt.Errorf("binary missing at %s and no MASTER_URL provided", fullBinaryPath)
		}
		// We can't validate much else if it's missing, we rely on Updater to fetch it
	}

	if c.StartingPort <= 0 {
		return fmt.Errorf("STARTING_PORT must be > 0 (use -sp flag or env)")
	}
	if c.MaxInstances <= 0 {
		return fmt.Errorf("MAX_INSTANCES must be > 0 (use -max flag or env)")
	}

	return nil
}

func getEnv(key string, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

// SaveConfigToEnv saves configuration settings to the .env file
func SaveConfigToEnv(updates map[string]string) error {
	envFile := ".env"

	// Read existing content
	content, err := os.ReadFile(envFile)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	lines := []string{}
	if len(content) > 0 {
		lines = splitLines(string(content))
	}

	// Track which keys we've updated
	processed := make(map[string]bool)

	// Update existing lines
	for i, line := range lines {
		for key, value := range updates {
			prefix := key + "="
			if len(line) >= len(prefix) && line[:len(prefix)] == prefix {
				lines[i] = prefix + value
				processed[key] = true
			}
		}
	}

	// Add new keys
	for key, value := range updates {
		if !processed[key] {
			lines = append(lines, key+"="+value)
		}
	}

	// Write back
	newContent := ""
	for i, line := range lines {
		newContent += line
		if i < len(lines)-1 {
			newContent += "\n"
		}
	}

	return os.WriteFile(envFile, []byte(newContent), 0600)
}

// splitLines splits a string into lines
func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			line := s[start:i]
			if len(line) > 0 && line[len(line)-1] == '\r' {
				line = line[:len(line)-1]
			}
			lines = append(lines, line)
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}
