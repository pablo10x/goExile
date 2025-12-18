package config

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Region            string
	Host              string // Hostname/IP reachable by Master Server
	GameBinaryPath    string
	GameDownloadURL   string
	GameDownloadToken string
	GameInstallDir    string
	Port              string // Port for the Spawner API itself
	StartingPort      int    // Starting port for game instances
	MaxInstances      int    // Maximum number of instances allowed
	MasterURL         string // URL of the Master Server (Registry)
	MasterAPIKey      string // API Key for Master Server authentication
	StateFilePath     string // Path to the JSON file storing instance state
	InstancesDir      string // Directory where game server instances are spawned
	EnrollmentKey     string // One-time enrollment key for initial registration
}

// Package-level flag variables
var (
	keyFlag    = flag.String("key", "", "One-time enrollment key for initial registration with master server")
	urlFlag    = flag.String("url", "", "URL of the Master Server")
	spFlag     = flag.Int("sp", 0, "Starting port for game instances")
	maxFlag    = flag.Int("max", 0, "Maximum number of instances")
	regionFlag = flag.String("region", "", "Region identifier (e.g. US-East)")

	// Spawner specific settings (no .env fallback for these initially, will be updated)
	hostFlag              = flag.String("host", "localhost", "Hostname/IP reachable by Master Server")
	gameBinaryPathFlag    = flag.String("game-binary", "", "Path to the game server binary relative to install dir")
	gameDownloadURLFlag   = flag.String("game-download-url", "", "URL to download game server files (if different from Master)")
	gameDownloadTokenFlag = flag.String("game-download-token", "", "Token for game server download (if required)")
	gameInstallDirFlag    = flag.String("install-dir", "", "Directory where game server is installed")
	portFlag              = flag.String("port", "8080", "Port for the Spawner API itself")
	stateFilePathFlag     = flag.String("state-file", "instances.json", "Path to the JSON file storing instance state")
	instancesDirFlag      = flag.String("instances-dir", "", "Directory where game server instances are spawned")
)

// Load reads configuration from environment variables and command-line flags.
// It assumes flag.Parse() has already been called.
func Load() (*Config, error) {
	// Try loading .env from local directory only
	_ = godotenv.Load(".env")

	conf := &Config{
		// These fields ONLY come from flags or hardcoded defaults (if flag not provided)
		Region:            *regionFlag,
		Host:              *hostFlag,
		GameDownloadURL:   *gameDownloadURLFlag,
		GameDownloadToken: *gameDownloadTokenFlag,
		Port:              *portFlag,
		StartingPort:      *spFlag,
		MaxInstances:      *maxFlag,
		StateFilePath:     *stateFilePathFlag,
		EnrollmentKey:     *keyFlag,

		// These fields can fall back to environment variables or flags
		MasterURL:    getEnv("MASTER_URL", *urlFlag),
		MasterAPIKey: getEnv("MASTER_API_KEY", ""),

		// These specific game server settings now also fall back to environment variables
		GameBinaryPath: getEnv("GAME_BINARY_PATH", *gameBinaryPathFlag),
		GameInstallDir: getEnv("GAME_INSTALL_DIR", *gameInstallDirFlag),
		InstancesDir:   getEnv("INSTANCES_DIR", *instancesDirFlag),
	}
	
	// Apply flag overrides to ensure flags always take precedence.
	// For flags that provide their own non-zero/non-empty defaults, those already take precedence implicitly.
	// This block handles cases where an empty env var should be overridden by a non-empty flag,
	// or where the flag's default is 0/"" but a non-zero/non-empty value was explicitly provided.

	if *urlFlag != "" { // Override env if flag was explicitly provided
		conf.MasterURL = *urlFlag
	}
	// For starting port and max instances, their flags have 0 default.
	// We need to prioritize flag > env > hardcoded default (7777, 10)
	if *spFlag != 0 {
		conf.StartingPort = *spFlag
	} else if conf.StartingPort == 0 { // if env also didn't set it (and flag default is 0)
		conf.StartingPort = 7777 // Hardcoded default
	}

	if *maxFlag != 0 {
		conf.MaxInstances = *maxFlag
	} else if conf.MaxInstances == 0 { // if env also didn't set it (and flag default is 0)
		conf.MaxInstances = 10 // Hardcoded default
	}

	// For string flags, an empty string means the flag wasn't provided, so env takes over.
	// Only override env if flag was explicitly provided.
	if *gameBinaryPathFlag != "" {
		conf.GameBinaryPath = *gameBinaryPathFlag
	}
	if *gameInstallDirFlag != "" {
		conf.GameInstallDir = *gameInstallDirFlag
	}
	if *instancesDirFlag != "" {
		conf.InstancesDir = *instancesDirFlag
	}

	if *regionFlag != "" {
		conf.Region = *regionFlag
	}

	return conf, nil
}

// Validate checks if the configuration is valid and sufficient to start the application.
func (c *Config) Validate() error {
	if c.Region == "" {
		return fmt.Errorf("REGION is required (use -region flag)")
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

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

func getEnvAsInt(key string, fallback int) int {
	valueStr := getEnv(key, "")
	if valueStr == "" {
		return fallback
	}
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return fallback
	}
	return value
}
