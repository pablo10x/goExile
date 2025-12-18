package enrollment

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"spawner/internal/config"
	"strings"
	"time"
)

// EnrollmentResult contains the result of a successful enrollment
type EnrollmentResult struct {
	ID     int    `json:"id"`
	APIKey string `json:"api_key"`
}

// EnrollmentRequest represents the request body for enrollment
type EnrollmentRequest struct {
	Key              string `json:"key"`
	Region           string `json:"region"`
	Host             string `json:"host"`
	Port             int    `json:"port"`
	MaxInstances     int    `json:"max_instances"`
	CurrentInstances int    `json:"current_instances"`
	Status           string `json:"status"`
}

// Enroll performs the enrollment process using the provided enrollment key
// Returns the API key and spawner ID on success
func Enroll(cfg *config.Config, logger *slog.Logger) (*EnrollmentResult, error) {
	if cfg.EnrollmentKey == "" {
		return nil, fmt.Errorf("enrollment key is required")
	}

	logger.Info("Starting enrollment process", "master_url", cfg.MasterURL)

	// First, validate the key
	if err := validateKey(cfg, logger); err != nil {
		return nil, fmt.Errorf("key validation failed: %w", err)
	}

	// Build enrollment URL
	baseURL := strings.TrimSuffix(cfg.MasterURL, "/")
	enrollURL := baseURL + "/api/enrollment/register"

	// Calculate max instances
	port, _ := parsePort(cfg.Port)
	maxInstances := cfg.MaxInstances
	if maxInstances < 1 {
		maxInstances = 1
	}

	// Build request
	reqBody := EnrollmentRequest{
		Key:              cfg.EnrollmentKey,
		Region:           cfg.Region,
		Host:             cfg.Host,
		Port:             port,
		MaxInstances:     maxInstances,
		CurrentInstances: 0,
		Status:           "Online",
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal enrollment request: %w", err)
	}

	logger.Info("Sending enrollment request",
		"url", enrollURL,
		"region", cfg.Region,
		"host", cfg.Host,
		"port", port,
		"max_instances", maxInstances,
	)

	// Create HTTP request
	req, err := http.NewRequest("POST", enrollURL, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Send request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("enrollment request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check status code
	if resp.StatusCode != http.StatusCreated {
		var errResp struct {
			Error string `json:"error"`
		}
		if json.Unmarshal(body, &errResp) == nil && errResp.Error != "" {
			return nil, fmt.Errorf("enrollment failed: %s", errResp.Error)
		}
		return nil, fmt.Errorf("enrollment failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse successful response
	var result struct {
		ID      int    `json:"id"`
		APIKey  string `json:"api_key"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse enrollment response: %w", err)
	}

	logger.Info("✅ Enrollment successful!",
		"spawner_id", result.ID,
		"message", result.Message,
	)

	return &EnrollmentResult{
		ID:     result.ID,
		APIKey: result.APIKey,
	}, nil
}

// validateKey checks if the enrollment key is valid before attempting registration
func validateKey(cfg *config.Config, logger *slog.Logger) error {
	baseURL := strings.TrimSuffix(cfg.MasterURL, "/")
	validateURL := baseURL + "/api/enrollment/validate"

	reqBody := map[string]string{"key": cfg.EnrollmentKey}
	jsonBody, _ := json.Marshal(reqBody)

	req, err := http.NewRequest("POST", validateURL, bytes.NewReader(jsonBody))
	if err != nil {
		return fmt.Errorf("failed to create validation request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("validation request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		var errResp struct {
			Error string `json:"error"`
		}
		if json.Unmarshal(body, &errResp) == nil && errResp.Error != "" {
			return fmt.Errorf("%s", errResp.Error)
		}
		return fmt.Errorf("validation failed with status %d", resp.StatusCode)
	}

	var validResp struct {
		Valid     bool    `json:"valid"`
		ExpiresAt string  `json:"expires_at"`
		TTL       float64 `json:"ttl"`
	}
	if err := json.Unmarshal(body, &validResp); err != nil {
		return fmt.Errorf("failed to parse validation response: %w", err)
	}

	if !validResp.Valid {
		return fmt.Errorf("enrollment key is not valid")
	}

	logger.Info("✅ Enrollment key validated",
		"ttl_seconds", validResp.TTL,
	)

	return nil
}

// parsePort converts a string port to int
func parsePort(portStr string) (int, error) {
	var port int
	_, err := fmt.Sscanf(portStr, "%d", &port)
	return port, err
}
