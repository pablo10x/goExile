package utils

import (
	"log"
	"os"
	"strings"

	"github.com/jmoiron/sqlx"
)

// GetAllowedOrigins returns the list of allowed WebSocket origins.
// In Production Mode, it relies strictly on DB config and Environment variables.
// In Development Mode, it includes localhost defaults.
func GetAllowedOrigins(db *sqlx.DB) []string {
	isProduction := os.Getenv("PRODUCTION_MODE") == "true"
	
	// Start with defaults only if NOT in production
	var defaults []string
	if !isProduction {
		defaults = []string{
			"http://localhost:5173", // SvelteKit dev server
			"http://localhost:8081", // Backend server
			"http://127.0.0.1:5173",
			"http://127.0.0.1:8081",
			"http://tauri.localhost",  // Tauri Windows/Linux
			"https://tauri.localhost", // Tauri macOS/Production
			"tauri://localhost",       // Old Tauri
		}
	} else {
		// In production, start empty or potentially with the SERVER_HOST if relevant
		// But safer to start empty and rely on configuration
		defaults = []string{}
	}

	// 1. Load from DB (server_config table) if available
	if db != nil {
		// specific SQL to avoid circular dependency on config package if possible, 
		// or just use raw query here since it's a util.
		// However, to keep it clean, we'll try a simple query.
		var value string
		err := db.Get(&value, "SELECT value FROM server_config WHERE key = 'allowed_origins'")
		if err == nil && value != "" {
			defaults = append(defaults, SplitAndTrim(value, ",")...)
		}
	}

	// 2. Add custom origins from environment variable
	// Format: ALLOWED_ORIGINS=https://example.com,https://admin.example.com
	if customOrigins := GetEnv("ALLOWED_ORIGINS", ""); customOrigins != "" {
		for _, origin := range SplitAndTrim(customOrigins, ",") {
			if origin != "" {
				defaults = append(defaults, origin)
			}
		}
	}
	
	// Deduplicate
	seen := make(map[string]bool)
	result := make([]string, 0)
	for _, origin := range defaults {
		if !seen[origin] {
			seen[origin] = true
			result = append(result, origin)
		}
	}
	
	// If production and no origins found, log a warning
	if isProduction && len(result) == 0 {
		log.Println("⚠️  WARNING: Production mode is enabled but no ALLOWED_ORIGINS are configured. WebSocket connections may fail.")
	}

	return result
}

// SplitAndTrim splits a string and trims whitespace from each part
func SplitAndTrim(s, sep string) []string {
	parts := make([]string, 0)
	for _, part := range strings.Split(s, sep) {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			parts = append(parts, trimmed)
		}
	}
	return parts
}
