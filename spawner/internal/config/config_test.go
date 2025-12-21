package config

import (
	"os"
	"testing"
)

func TestConfig_Validate(t *testing.T) {
	// Setup a temporary file for binary check
	tmpFile, err := os.CreateTemp("", "game_binary*.exe")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Remove(tmpFile.Name()) }()

	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "Valid Config",
			config: Config{
				Region:         "EU",
				GameBinaryPath: tmpFile.Name(),
				StartingPort:   7777,
				MaxInstances:   10,
				MasterAPIKey:   "test-key",
			},
			wantErr: false,
		},
		{
			name: "Missing Region",
			config: Config{
				Region:         "",
				GameBinaryPath: tmpFile.Name(),
				StartingPort:   7777,
				MaxInstances:   10,
				MasterAPIKey:   "test-key",
			},
			wantErr: true,
		},
		{
			name: "Missing Binary",
			config: Config{
				Region:         "EU",
				GameBinaryPath: "/non/existent/path",
				StartingPort:   7777,
				MaxInstances:   10,
				MasterAPIKey:   "test-key",
			},
			wantErr: true,
		},
		{
			name: "Invalid Starting Port",
			config: Config{
				Region:         "EU",
				GameBinaryPath: tmpFile.Name(),
				StartingPort:   0,
				MaxInstances:   10,
				MasterAPIKey:   "test-key",
			},
			wantErr: true,
		},
		{
			name: "Invalid Max Instances",
			config: Config{
				Region:         "EU",
				GameBinaryPath: tmpFile.Name(),
				StartingPort:   7777,
				MaxInstances:   0,
				MasterAPIKey:   "test-key",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Config.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
