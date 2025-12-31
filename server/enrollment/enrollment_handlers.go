package enrollment

import (
	"net/http"
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

// ValidateEnrollmentKeyHandler checks if a key is valid (used by node before full registration)
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
// This is used by the dashboard to poll and check if a node has used the key
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
		"status":     key.Status, // Use actual status (active, pending, approved)
		"used":       key.Used,
		"expired":    expired,
		"expires_at": key.ExpiresAt,
		"ttl":        ttl,
	}

	// Always include NodeInfo if available (crucial for "pending" state)
	if key.NodeInfo != nil {
		response["node_info"] = key.NodeInfo
	}

	if key.Used {
		response["status"] = "used" // or "approved"
		response["used_at"] = key.UsedAt
		if key.UsedBy != nil {
			response["node_id"] = *key.UsedBy
		}
	} else if expired {
		response["status"] = "expired"
	}

	utils.WriteJSON(w, http.StatusOK, response)
}

// RegisterNodeWithKeyHandler handles node enrollment attempt
// If key is valid, it marks it as "pending" and waits for admin approval
func RegisterNodeWithKeyHandler(w http.ResponseWriter, r *http.Request) {
	if enrollmentManager == nil {
		utils.WriteError(w, r, http.StatusInternalServerError, "enrollment manager not initialized")
		return
	}

	var req struct {
		Key  string `json:"key"`
		Host string `json:"host"`
		Port int    `json:"port"`
	}

	if err := utils.DecodeJSON(r, &req); err != nil {
		utils.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	if req.Key == "" {
		utils.WriteError(w, r, http.StatusBadRequest, "enrollment key is required")
		return
	}

	// Check status
	key, exists := enrollmentManager.GetKey(req.Key)
	if !exists {
		utils.WriteError(w, r, http.StatusNotFound, "invalid enrollment key")
		return
	}

	if key.Status == "approved" && key.NodeInfo != nil {
		// Already approved, return config
		utils.WriteJSON(w, http.StatusOK, map[string]interface{}{
			"status":        "approved",
			"id":            key.NodeInfo.ID,
			"api_key":       key.NodeInfo.APIKey,
			"region":        key.NodeInfo.Region,
			"max_instances": key.NodeInfo.MaxInstances,
			"message":       "Node enrolled successfully",
		})
		return
	}

	// Claim/Update pending status
	_, err := enrollmentManager.ClaimKey(req.Key, req.Host, req.Port)
	if err != nil {
		utils.WriteError(w, r, http.StatusUnauthorized, err.Error())
		return
	}

	utils.WriteJSON(w, http.StatusAccepted, map[string]interface{}{
		"status":  "pending",
		"message": "Waiting for admin approval in dashboard",
	})
}

// ApproveEnrollmentHandler completes the registration from the dashboard
func ApproveEnrollmentHandler(w http.ResponseWriter, r *http.Request) {
	if enrollmentManager == nil {
		utils.WriteError(w, r, http.StatusInternalServerError, "enrollment manager not initialized")
		return
	}

	var req struct {
		Key          string `json:"key"`
		Region       string `json:"region"`
		MaxInstances int    `json:"max_instances"`
		Port         int    `json:"port"` // Optional override
	}

	if err := utils.DecodeJSON(r, &req); err != nil {
		utils.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	enrollKey, exists := enrollmentManager.GetKey(req.Key)
	if !exists || enrollKey.Status != "pending" {
		utils.WriteError(w, r, http.StatusBadRequest, "invalid or non-pending enrollment key")
		return
	}

	// Final data
	host := enrollKey.NodeInfo.Host
	port := enrollKey.NodeInfo.Port
	if req.Port > 0 {
		port = req.Port
	}

	// Generate a unique API key for this node
	apiKey, err := utils.GenerateRandomKey(32)
	if err != nil {
		utils.WriteError(w, r, http.StatusInternalServerError, "failed to generate api key")
		return
	}

	// Create the node object
	s := models.Node{
		Region:           req.Region,
		Host:             host,
		Port:             port,
		MaxInstances:     req.MaxInstances,
		CurrentInstances: 0,
		Status:           "Offline",
		LastSeen:         time.Now(),
		APIKey:           apiKey,
	}

	// Register the node in the global registry (persists to DB)
	id, err := registry.GlobalRegistry.Register(&s)
	if err != nil {
		utils.WriteError(w, r, http.StatusInternalServerError, "failed to register node: "+err.Error())
		return
	}

	// Approve the key so node can get the config on next poll
	_, err = enrollmentManager.ApproveKey(req.Key, req.Region, req.MaxInstances, id, apiKey)
	if err != nil {
		utils.WriteError(w, r, http.StatusInternalServerError, "failed to approve key: "+err.Error())
		return
	}

	registry.GlobalStats.UpdateActiveNodes(len(registry.GlobalRegistry.List()))

	utils.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"id":      id,
		"message": "Node approved and registered",
	})
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
