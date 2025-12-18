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
	regionFlag = flag.String("region", "", "Region identifier")
	// Other config flags can be added here
)

// Load reads configuration from environment variables and command-line flags.
// It assumes flag.Parse() has already been called.
func Load() (*Config, error) {
	// Try loading .env from local directory only
	_ = godotenv.Load(".env")

	conf := &Config{
		Region:            getEnv("REGION", *regionFlag),
		Host:              getEnv("SPAWNER_HOST", "localhost"),
		GameBinaryPath:    getEnv("GAME_BINARY_PATH", ""),
		GameDownloadURL:   getEnv("GAME_DOWNLOAD_URL", ""),
		GameDownloadToken: getEnv("GAME_DOWNLOAD_TOKEN", ""),
		GameInstallDir:    getEnv("GAME_INSTALL_DIR", "./game_server"),
		Port:              getEnv("SPAWNER_PORT", "8080"),
		StartingPort:      getEnvAsInt("STARTING_PORT", 7777),
		MaxInstances:      getEnvAsInt("MAX_INSTANCES", 10),
		MasterURL:         getEnv("MASTER_URL", "http://localhost:8081"),
		MasterAPIKey:      getEnv("MASTER_API_KEY", ""),
		StateFilePath:     getEnv("STATE_FILE_PATH", "instances.json"),
		InstancesDir:      getEnv("INSTANCES_DIR", "./instances"),
		EnrollmentKey:     *keyFlag,
	}

	// Override with flags if provided
	if *urlFlag != "" {
		conf.MasterURL = *urlFlag
	}
	if *spFlag != 0 {
		conf.StartingPort = *spFlag
	}
	if *maxFlag != 0 {
		conf.MaxInstances = *maxFlag
	}
	if *regionFlag != "" {
		conf.Region = *regionFlag
	}

	if err := conf.Validate(); err != nil {
		return nil, err
	}

	return conf, nil
}

// Validate checks if the configuration is valid and sufficient to start the application.
func (c *Config) Validate() error {
	if c.Region == "" {
		// Region is required for identity
		return fmt.Errorf("REGION is required (use -region flag or env)")
	}
	if c.GameBinaryPath == "" {
		return fmt.Errorf("GAME_BINARY_PATH is required")
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
		return fmt.Errorf("STARTING_PORT must be > 0")
	}
	if c.MaxInstances <= 0 {
		return fmt.Errorf("MAX_INSTANCES must be > 0")
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
