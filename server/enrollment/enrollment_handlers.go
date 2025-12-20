package enrollment

import (
	"net/http"
	"os"
	"time"

	"exile/server/models"
	"exile/server/registry"
	"exile/server/utils"
)

// Global enrollment manager
var enrollmentManager *EnrollmentManager

// InitializeEnrollmentManager initializes the global enrollment manager
func InitializeEnrollmentManager() {
	enrollmentManager = NewEnrollmentManager()
}

// GenerateEnrollmentKeyHandler creates a new enrollment key
func GenerateEnrollmentKeyHandler(w http.ResponseWriter, r *http.Request) {
	if enrollmentManager == nil {
		utils.WriteError(w, r, http.StatusInternalServerError, "enrollment manager not initialized")
		return
	}

	// Get the user from the session (could be enhanced to get from context)
	createdBy := "admin"

	// Generate a key that expires in 2 minutes
	key, err := enrollmentManager.GenerateKey(createdBy, 2*time.Minute)
	if err != nil {
		utils.WriteError(w, r, http.StatusInternalServerError, "failed to generate enrollment key")
		return
	}

	utils.WriteJSON(w, http.StatusCreated, key)
}

// ListEnrollmentKeysHandler returns all active enrollment keys
func ListEnrollmentKeysHandler(w http.ResponseWriter, r *http.Request) {
	if enrollmentManager == nil {
		utils.WriteError(w, r, http.StatusInternalServerError, "enrollment manager not initialized")
		return
	}

	keys := enrollmentManager.GetActiveKeys()
	if keys == nil {
		keys = []*models.EnrollmentKey{}
	}
	utils.WriteJSON(w, http.StatusOK, keys)
}

// ValidateEnrollmentKeyHandler checks if a key is valid (used by spawner before full registration)
func ValidateEnrollmentKeyHandler(w http.ResponseWriter, r *http.Request) {
	if enrollmentManager == nil {
		utils.WriteError(w, r, http.StatusInternalServerError, "enrollment manager not initialized")
		return
	}

	var req struct {
		Key string `json:"key"`
	}

	if err := utils.DecodeJSON(r, &req); err != nil {
		utils.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	if req.Key == "" {
		utils.WriteError(w, r, http.StatusBadRequest, "enrollment key is required")
		return
	}

	// Check if key exists and is valid (without consuming it)
	key, exists := enrollmentManager.GetKey(req.Key)
	if !exists {
		utils.WriteError(w, r, http.StatusNotFound, "invalid enrollment key")
		return
	}

	if time.Now().After(key.ExpiresAt) {
		utils.WriteError(w, r, http.StatusGone, "enrollment key has expired")
		return
	}

	if key.Used {
		utils.WriteError(w, r, http.StatusConflict, "enrollment key has already been used")
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"valid":      true,
		"expires_at": key.ExpiresAt,
		"ttl":        time.Until(key.ExpiresAt).Seconds(),
	})
}

// GetEnrollmentKeyStatusHandler returns the current status of an enrollment key
// This is used by the dashboard to poll and check if a spawner has used the key
func GetEnrollmentKeyStatusHandler(w http.ResponseWriter, r *http.Request) {
	if enrollmentManager == nil {
		utils.WriteError(w, r, http.StatusInternalServerError, "enrollment manager not initialized")
		return
	}

	var req struct {
		Key string `json:"key"`
	}

	if err := utils.DecodeJSON(r, &req); err != nil {
		utils.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	if req.Key == "" {
		utils.WriteError(w, r, http.StatusBadRequest, "enrollment key is required")
		return
	}

	key, exists := enrollmentManager.GetKey(req.Key)
	if !exists {
		// Key might have been cleaned up after expiration
		utils.WriteJSON(w, http.StatusOK, map[string]interface{}{
			"status":  "expired",
			"used":    false,
			"expired": true,
		})
		return
	}

	now := time.Now()
	expired := now.After(key.ExpiresAt)
	ttl := key.ExpiresAt.Sub(now).Seconds()
	if ttl < 0 {
		ttl = 0
	}

	response := map[string]interface{}{
		"status":     "pending",
		"used":       key.Used,
		"expired":    expired,
		"expires_at": key.ExpiresAt,
		"ttl":        ttl,
	}

	if key.Used {
		response["status"] = "used"
		response["used_at"] = key.UsedAt
		if key.SpawnerInfo != nil {
			response["spawner_info"] = key.SpawnerInfo
		}
		if key.UsedBy != nil {
			response["spawner_id"] = *key.UsedBy
		}
	} else if expired {
		response["status"] = "expired"
	}

	utils.WriteJSON(w, http.StatusOK, response)
}

// RegisterSpawnerWithKeyHandler handles spawner registration using an enrollment key
// This endpoint does NOT require API key auth - the enrollment key IS the auth
func RegisterSpawnerWithKeyHandler(w http.ResponseWriter, r *http.Request) {
	if enrollmentManager == nil {
		utils.WriteError(w, r, http.StatusInternalServerError, "enrollment manager not initialized")
		return
	}

	var req struct {
		Key              string `json:"key"`
		Region           string `json:"region"`
		Host             string `json:"host"`
		Port             int    `json:"port"`
		MaxInstances     int    `json:"max_instances"`
		CurrentInstances int    `json:"current_instances"`
		Status           string `json:"status"`
	}

	if err := utils.DecodeJSON(r, &req); err != nil {
		utils.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	if req.Key == "" {
		utils.WriteError(w, r, http.StatusBadRequest, "enrollment key is required")
		return
	}

	if req.Host == "" {
		utils.WriteError(w, r, http.StatusBadRequest, "host is required")
		return
	}

	if req.Port < 1 || req.Port > 65535 {
		utils.WriteError(w, r, http.StatusBadRequest, "invalid port")
		return
	}

	// Validate and consume the enrollment key
	enrollmentKey, err := enrollmentManager.ValidateAndUseKey(req.Key)
	if err != nil {
		utils.WriteError(w, r, http.StatusUnauthorized, err.Error())
		return
	}

	// Create the spawner object
	s := models.Spawner{
		Region:           req.Region,
		Host:             req.Host,
		Port:             req.Port,
		MaxInstances:     req.MaxInstances,
		CurrentInstances: req.CurrentInstances,
		Status:           req.Status,
		LastSeen:         time.Now(),
	}

	// Register the spawner
	id, err := registry.GlobalRegistry.Register(&s)
	if err != nil {
		utils.WriteError(w, r, http.StatusInternalServerError, "failed to register spawner")
		return
	}

	// Update the enrollment key with spawner info for dashboard tracking
	enrollmentKey.UsedBy = &id
	enrollmentKey.SpawnerInfo = &struct {
		ID     int    `json:"id"`
		Region string `json:"region"`
		Host   string `json:"host"`
		Port   int    `json:"port"`
	}{
		ID:     id,
		Region: req.Region,
		Host:   req.Host,
		Port:   req.Port,
	}

	registry.GlobalStats.UpdateActiveServers(len(registry.GlobalRegistry.List()))

	// Return the spawner ID and the API key for future communications
	apiKey := getEnrollmentAPIKey()
	utils.WriteJSON(w, http.StatusCreated, map[string]interface{}{
		"id":      id,
		"api_key": apiKey,
		"message": "Spawner enrolled successfully",
	})
}

// getEnrollmentAPIKey retrieves the API key from environment
func getEnrollmentAPIKey() string {
	apiKey := os.Getenv("MASTER_API_KEY")
	if apiKey == "" {
		apiKey = "dev_master_key"
	}
	return apiKey
}

// RevokeEnrollmentKeyHandler revokes an active enrollment key
func RevokeEnrollmentKeyHandler(w http.ResponseWriter, r *http.Request) {
	if enrollmentManager == nil {
		utils.WriteError(w, r, http.StatusInternalServerError, "enrollment manager not initialized")
		return
	}

	var req struct {
		Key string `json:"key"`
	}

	if err := utils.DecodeJSON(r, &req); err != nil {
		utils.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	if req.Key == "" {
		utils.WriteError(w, r, http.StatusBadRequest, "enrollment key is required")
		return
	}

	if !enrollmentManager.RevokeKey(req.Key) {
		utils.WriteError(w, r, http.StatusNotFound, "enrollment key not found")
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]string{"message": "enrollment key revoked"})
}
