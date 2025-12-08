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
	defer os.Remove(tmpFile.Name())

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
				MinGamePort:    7000,
				MaxGamePort:    8000,
			},
			wantErr: false,
		},
		{
			name: "Missing Region",
			config: Config{
				Region:         "",
				GameBinaryPath: tmpFile.Name(),
			},
			wantErr: true,
		},
		{
			name: "Missing Binary",
			config: Config{
				Region:         "EU",
				GameBinaryPath: "/non/existent/path",
			},
			wantErr: true,
		},
		{
			name: "Invalid Ports",
			config: Config{
				Region:         "EU",
				GameBinaryPath: tmpFile.Name(),
				MinGamePort:    9000,
				MaxGamePort:    8000,
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
