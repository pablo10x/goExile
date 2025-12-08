package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

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
	MinGamePort       int
	MaxGamePort       int
	MasterURL         string // URL of the Master Server (Registry)
	MasterAPIKey      string // API Key for Master Server authentication
	StateFilePath     string // Path to the JSON file storing instance state
	InstancesDir      string // Directory where game server instances are spawned
}

// Load reads configuration from environment variables.
// It attempts to load .env files from current and parent directories to support
// different running contexts (e.g., direct binary execution vs `go run`).
func Load() (*Config, error) {
	// Try loading .env from multiple locations
	_ = godotenv.Load(".env")
	_ = godotenv.Load("spawner/.env") // If running from project root
	_ = godotenv.Load("../../.env")   // If running from cmd/server
	_ = godotenv.Load("../.env")
	
	conf := &Config{
		Region:            getEnv("REGION", ""),
		Host:              getEnv("SPAWNER_HOST", "localhost"),
		GameBinaryPath:    getEnv("GAME_BINARY_PATH", ""),
		GameDownloadURL:   getEnv("GAME_DOWNLOAD_URL", ""),
		GameDownloadToken: getEnv("GAME_DOWNLOAD_TOKEN", ""),
		GameInstallDir:    getEnv("GAME_INSTALL_DIR", "./game_server"),
		Port:              getEnv("SPAWNER_PORT", "8080"),
		MinGamePort:       getEnvAsInt("MIN_GAME_PORT", 7777),
		MaxGamePort:       getEnvAsInt("MAX_GAME_PORT", 8000),
		MasterURL:         getEnv("MASTER_URL", "http://localhost:8081"),
		MasterAPIKey:      getEnv("MASTER_API_KEY", ""),
		StateFilePath:     getEnv("STATE_FILE_PATH", "instances.json"),
		InstancesDir:      getEnv("INSTANCES_DIR", "./instances"),
	}

	if err := conf.Validate(); err != nil {
		return nil, err
	}

	return conf, nil
}

// Validate checks if the configuration is valid and sufficient to start the application.
func (c *Config) Validate() error {
	if c.Region == "" {
		cwd, _ := os.Getwd()
		return fmt.Errorf("REGION is required (CWD: %s)", cwd)
	}
	if c.GameBinaryPath == "" {
		return fmt.Errorf("GAME_BINARY_PATH is required")
	}

	// Adjust binary path for Windows if needed
	if runtime.GOOS == "windows" {
		if !strings.HasSuffix(strings.ToLower(c.GameBinaryPath), ".exe") {
			c.GameBinaryPath += ".exe"
		}
	}
	
	// Check if binary exists
	// Try original path
	if _, err := os.Stat(c.GameBinaryPath); os.IsNotExist(err) {
		// Try adding spawner/ prefix (if running from root)
		altPath := filepath.Join("spawner", c.GameBinaryPath)
		if _, err2 := os.Stat(altPath); err2 == nil {
			c.GameBinaryPath = altPath // Update to working path
		} else {
			// If binary is missing, we MUST have a download URL to recover
			if c.GameDownloadURL == "" {
				return fmt.Errorf("GAME_BINARY_PATH does not exist: %s (and tried %s), and no GAME_DOWNLOAD_URL provided", c.GameBinaryPath, altPath)
			}
			// If we have a download URL, we also need an install directory
			if c.GameInstallDir == "" {
				return fmt.Errorf("GAME_INSTALL_DIR is required when downloading game server")
			}
		}
	}




	if c.MinGamePort >= c.MaxGamePort {
		return fmt.Errorf("MIN_GAME_PORT (%d) must be less than MAX_GAME_PORT (%d)", c.MinGamePort, c.MaxGamePort)
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