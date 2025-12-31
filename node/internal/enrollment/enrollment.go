// Package enrollment handles the initial registration and API key retrieval for the node.
package enrollment

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"node/internal/config"
	"strings"
	"time"
)

// Result contains the result of a successful enrollment
type Result struct {
	ID     int    `json:"id"`
	APIKey string `json:"api_key"`
}

// Request represents the request body for enrollment
type Request struct {
	Key  string `json:"key"`
	Host string `json:"host"`
	Port int    `json:"port"`
}

// Enroll performs the enrollment process using the provided enrollment key
// It polls the master server until the enrollment is approved by an admin.
func Enroll(cfg *config.Config, logger *slog.Logger) (*Result, error) {
	if cfg.EnrollmentKey == "" {
		return nil, fmt.Errorf("enrollment key is required")
	}

	logger.Info("Starting enrollment process", "master_url", cfg.MasterURL)

	// Build enrollment URL
	baseURL := strings.TrimSuffix(cfg.MasterURL, "/")
	enrollURL := baseURL + "/api/enrollment/register"

	nodePort, _ := parsePort(cfg.Port)

	// Build request
	reqBody := Request{
		Key:  cfg.EnrollmentKey,
		Host: cfg.Host,
		Port: nodePort,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal enrollment request: %w", err)
	}

	// Poll until approved
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	logger.Info("Waiting for admin approval in dashboard...")

	for {
		// Create HTTP request
		req, err := http.NewRequest("POST", enrollURL, bytes.NewReader(jsonBody))
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")

		// Send request
		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			logger.Warn("Enrollment request failed, retrying...", "error", err)
		} else {
			defer func() { _ = resp.Body.Close() }()

			// Read response body
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return nil, fmt.Errorf("failed to read response: %w", err)
			}

			if resp.StatusCode == http.StatusOK {
				// Approved!
				var result struct {
					Status       string `json:"status"`
					ID           int    `json:"id"`
					APIKey       string `json:"api_key"`
					Region       string `json:"region"`
					MaxInstances int    `json:"max_instances"`
					Message      string `json:"message"`
				}
				if err := json.Unmarshal(body, &result); err != nil {
					return nil, fmt.Errorf("failed to parse enrollment response: %w", err)
				}

				if result.Status == "approved" {
					logger.Info("✅ Enrollment approved!",
						"node_id", result.ID,
						"region", result.Region,
						"max_instances", result.MaxInstances,
					)

					// Update local config with master-provided values
					cfg.Region = result.Region
					cfg.MaxInstances = result.MaxInstances

					return &Result{
						ID:     result.ID,
						APIKey: result.APIKey,
					}, nil
				}
			} else if resp.StatusCode == http.StatusAccepted {
				// Still pending, just continue loop
			} else {
				// Error
				var errResp struct {
					Error string `json:"error"`
				}
				if json.Unmarshal(body, &errResp) == nil && errResp.Error != "" {
					return nil, fmt.Errorf("enrollment failed: %s", errResp.Error)
				}
				return nil, fmt.Errorf("enrollment failed with status %d: %s", resp.StatusCode, string(body))
			}
		}

		select {
		case <-ticker.C:
			continue
		}
	}
}

// parsePort converts a string port to int
func parsePort(portStr string) (int, error) {
	var port int
	_, err := fmt.Sscanf(portStr, "%d", &port)
	return port, err
}
