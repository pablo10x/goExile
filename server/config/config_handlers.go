package config

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"exile/server/database"
	"exile/server/models"
	"exile/server/utils"
)

// GetAllConfigHandler returns all configuration settings
func GetAllConfigHandler(w http.ResponseWriter, r *http.Request) {
	configs, err := database.GetAllConfig(database.DBConn)
	if err != nil {
		utils.WriteError(w, r, http.StatusInternalServerError, fmt.Sprintf("Failed to get configuration: %v", err))
		return
	}

	utils.WriteJSON(w, http.StatusOK, configs)
}

// GetConfigByCategoryHandler returns configuration settings for a specific category
func GetConfigByCategoryHandler(w http.ResponseWriter, r *http.Request) {
	category := strings.TrimPrefix(r.URL.Path, "/api/config/category/")
	if category == "" {
		utils.WriteError(w, r, http.StatusBadRequest, "Category is required")
		return
	}

	configs, err := database.GetConfigByCategory(database.DBConn, category)
	if err != nil {
		utils.WriteError(w, r, http.StatusInternalServerError, fmt.Sprintf("Failed to get configuration for category %s: %v", category, err))
		return
	}

	utils.WriteJSON(w, http.StatusOK, configs)
}

// GetConfigByKeyHandler returns a specific configuration setting by key
func GetConfigByKeyHandler(w http.ResponseWriter, r *http.Request) {
	key := strings.TrimPrefix(r.URL.Path, "/api/config/")
	if key == "" {
		utils.WriteError(w, r, http.StatusBadRequest, "Key is required")
		return
	}

	config, err := database.GetConfigByKey(database.DBConn, key)
	if err != nil {
		utils.WriteError(w, r, http.StatusInternalServerError, fmt.Sprintf("Failed to get configuration for key %s: %v", key, err))
		return
	}

	if config == nil {
		utils.WriteError(w, r, http.StatusNotFound, "Configuration not found")
		return
	}

	utils.WriteJSON(w, http.StatusOK, config)
}

// UpdateConfigHandler updates a configuration setting
func UpdateConfigHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "PUT" {
		utils.WriteError(w, r, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	key := strings.TrimPrefix(r.URL.Path, "/api/config/")
	if key == "" {
		utils.WriteError(w, r, http.StatusBadRequest, "Key is required")
		return
	}

	var req struct {
		Value string `json:"value"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, r, http.StatusBadRequest, "Invalid request body")
		return
	}

	existing, err := database.GetConfigByKey(database.DBConn, key)
	if err != nil {
		utils.WriteError(w, r, http.StatusInternalServerError, fmt.Sprintf("Failed to get existing configuration: %v", err))
		return
	}

	if existing == nil {
		utils.WriteError(w, r, http.StatusNotFound, "Configuration not found")
		return
	}

	if existing.IsReadOnly {
		utils.WriteError(w, r, http.StatusForbidden, "Cannot update read-only configuration")
		return
	}

	updatedBy := "admin"
	if err := database.UpdateConfig(database.DBConn, key, req.Value, updatedBy); err != nil {
		utils.WriteError(w, r, http.StatusInternalServerError, fmt.Sprintf("Failed to update configuration: %v", err))
		return
	}

	updated, _ := database.GetConfigByKey(database.DBConn, key)
	utils.WriteJSON(w, http.StatusOK, updated)
}

// CreateConfigHandler creates a new configuration setting
func CreateConfigHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		utils.WriteError(w, r, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var config models.ServerConfig
	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		utils.WriteError(w, r, http.StatusBadRequest, "Invalid request body")
		return
	}

	if config.Key == "" || config.Value == "" || config.Type == "" || config.Category == "" {
		utils.WriteError(w, r, http.StatusBadRequest, "Key, value, type, and category are required")
		return
	}

	config.UpdatedAt = time.Now()
	if config.UpdatedBy == "" {
		config.UpdatedBy = "admin"
	}

	if err := database.SaveConfig(database.DBConn, &config); err != nil {
		utils.WriteError(w, r, http.StatusInternalServerError, fmt.Sprintf("Failed to create configuration: %v", err))
		return
	}

	utils.WriteJSON(w, http.StatusCreated, config)
}
