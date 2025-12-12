package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// configHandlers.go provides HTTP handlers for configuration management.

// GetAllConfigHandler returns all configuration settings
func GetAllConfigHandler(w http.ResponseWriter, r *http.Request) {
	configs, err := GetAllConfig(dbConn)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get configuration: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(configs); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// GetConfigByCategoryHandler returns configuration settings for a specific category
func GetConfigByCategoryHandler(w http.ResponseWriter, r *http.Request) {
	category := strings.TrimPrefix(r.URL.Path, "/api/config/")
	if category == "" {
		http.Error(w, "Category is required", http.StatusBadRequest)
		return
	}

	configs, err := GetConfigByCategory(dbConn, category)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get configuration for category %s: %v", category, err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(configs); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// GetConfigByKeyHandler returns a specific configuration setting by key
func GetConfigByKeyHandler(w http.ResponseWriter, r *http.Request) {
	key := strings.TrimPrefix(r.URL.Path, "/api/config/key/")
	if key == "" {
		http.Error(w, "Key is required", http.StatusBadRequest)
		return
	}

	config, err := GetConfigByKey(dbConn, key)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get configuration for key %s: %v", key, err), http.StatusInternalServerError)
		return
	}

	if config == nil {
		http.Error(w, "Configuration not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(config); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// UpdateConfigHandler updates a configuration setting
func UpdateConfigHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "PUT" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	key := strings.TrimPrefix(r.URL.Path, "/api/config/")
	if key == "" {
		http.Error(w, "Key is required", http.StatusBadRequest)
		return
	}

	var req struct {
		Value string `json:"value"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Get existing config to check if it's read-only
	existing, err := GetConfigByKey(dbConn, key)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get existing configuration: %v", err), http.StatusInternalServerError)
		return
	}

	if existing == nil {
		http.Error(w, "Configuration not found", http.StatusNotFound)
		return
	}

	if existing.IsReadOnly {
		http.Error(w, "Cannot update read-only configuration", http.StatusForbidden)
		return
	}

	// Get user from session (simplified - in production, you'd get this from proper auth)
	updatedBy := "admin" // This should come from the authenticated user

	if err := UpdateConfig(dbConn, key, req.Value, updatedBy); err != nil {
		http.Error(w, fmt.Sprintf("Failed to update configuration: %v", err), http.StatusInternalServerError)
		return
	}

	// Return updated config
	updated, err := GetConfigByKey(dbConn, key)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get updated configuration: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(updated); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// CreateConfigHandler creates a new configuration setting
func CreateConfigHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var config ServerConfig
	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if config.Key == "" || config.Value == "" || config.Type == "" || config.Category == "" {
		http.Error(w, "Key, value, type, and category are required", http.StatusBadRequest)
		return
	}

	// Set defaults
	config.UpdatedAt = time.Now()
	if config.UpdatedBy == "" {
		config.UpdatedBy = "admin" // This should come from authenticated user
	}

	if err := SaveConfig(dbConn, &config); err != nil {
		http.Error(w, fmt.Sprintf("Failed to create configuration: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(config); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}
