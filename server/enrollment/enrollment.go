package enrollment

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"exile/server/models"
)

// EnrollmentManager manages temporary enrollment keys
type EnrollmentManager struct {
	keys map[string]*models.EnrollmentKey
	mu   sync.RWMutex
}

// NewEnrollmentManager creates a new enrollment manager
func NewEnrollmentManager() *EnrollmentManager {
	em := &EnrollmentManager{
		keys: make(map[string]*models.EnrollmentKey),
	}

	// Start cleanup goroutine
	go em.cleanupExpiredKeys()

	return em
}

// GenerateKey creates a new enrollment key that expires in the given duration
func (em *EnrollmentManager) GenerateKey(createdBy string, duration time.Duration, autoApprove bool, allowedRegions string, tags string, maxUsages int) (*models.EnrollmentKey, error) {
	em.mu.Lock()
	defer em.mu.Unlock()

	// Generate random key (32 hex characters)
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return nil, fmt.Errorf("failed to generate random key: %w", err)
	}

	key := hex.EncodeToString(bytes)

	if maxUsages <= 0 {
		maxUsages = 1
	}

	enrollmentKey := &models.EnrollmentKey{
		Key:            key,
		CreatedAt:      time.Now(),
		ExpiresAt:      time.Now().Add(duration),
		CreatedBy:      createdBy,
		Used:           false,
		Status:         "active",
		AutoApprove:    autoApprove,
		AllowedRegions: allowedRegions,
		Tags:           tags,
		MaxUsages:      maxUsages,
		UsageCount:     0,
	}

	em.keys[key] = enrollmentKey

	return enrollmentKey, nil
}

// ClaimKey marks a key as pending and associates it with initial node data
func (em *EnrollmentManager) ClaimKey(key string, host string, port int) (*models.EnrollmentKey, error) {
	em.mu.Lock()
	defer em.mu.Unlock()

	enrollmentKey, exists := em.keys[key]
	if !exists {
		return nil, fmt.Errorf("invalid enrollment key")
	}

	if time.Now().After(enrollmentKey.ExpiresAt) {
		delete(em.keys, key)
		return nil, fmt.Errorf("enrollment key has expired")
	}

	if enrollmentKey.UsageCount >= enrollmentKey.MaxUsages {
		return nil, fmt.Errorf("enrollment key has been fully consumed")
	}

	// For multi-use keys, we don't block concurrent pending states easily unless we track per-request
	// But standard flow: 1 key = 1 node attempt at a time usually?
	// Actually, if MaxUsages > 1, multiple nodes might claim it.
	// We'll allow it to stay "active" until fully used?
	// Current logic sets status to "pending". This effectively locks the key for ONE node.
	// To support multi-use properly, we'd need a list of Pending nodes per key.
	// For now, let's keep it simple: If Status is Pending, it's busy.
	// UNLESS MaxUsages > 1, then maybe we clone the key state?
	
	// Simplification: Multi-use keys just remain "active" and we spawn a transient "pending" record?
	// That requires architectural change.
	// Let's stick to: ClaimKey locks it for this specific request.
	
	if enrollmentKey.Status != "active" {
		// If it's pending/approved, it's busy.
		return nil, fmt.Errorf("enrollment key is currently in use (status: %s)", enrollmentKey.Status)
	}

	// Update status to pending and store initial info
	enrollmentKey.Status = "pending"
	enrollmentKey.NodeInfo = &struct {
		ID           int    `json:"id,omitempty"`
		Region       string `json:"region"`
		Host         string `json:"host"`
		Port         int    `json:"port"`
		MaxInstances int    `json:"max_instances"`
		APIKey       string `json:"api_key,omitempty"`
	}{
		Host: host,
		Port: port,
	}

	return enrollmentKey, nil
}

// ApproveKey completes the registration with admin-provided config
func (em *EnrollmentManager) ApproveKey(key string, region string, maxInstances int, nodeID int, apiKey string) (*models.EnrollmentKey, error) {
	em.mu.Lock()
	defer em.mu.Unlock()

	enrollmentKey, exists := em.keys[key]
	if !exists {
		return nil, fmt.Errorf("invalid enrollment key")
	}

	if enrollmentKey.Status != "pending" {
		return nil, fmt.Errorf("key is not in pending status")
	}

	// Update usage
	enrollmentKey.UsageCount++
	
	// Mark as used for this instance
	now := time.Now()
	enrollmentKey.UsedAt = &now
	enrollmentKey.UsedBy = &nodeID // Last used by

	if enrollmentKey.NodeInfo != nil {
		enrollmentKey.NodeInfo.ID = nodeID
		enrollmentKey.NodeInfo.Region = region
		enrollmentKey.NodeInfo.MaxInstances = maxInstances
		enrollmentKey.NodeInfo.APIKey = apiKey
	}

	// If key allows more usages, reset to active
	if enrollmentKey.UsageCount < enrollmentKey.MaxUsages {
		enrollmentKey.Status = "active"
		enrollmentKey.Used = false // Not fully used yet
		// Clear NodeInfo for next user?
		// Yes, otherwise next ClaimKey sees old data
		enrollmentKey.NodeInfo = nil
	} else {
		enrollmentKey.Status = "approved"
		enrollmentKey.Used = true
	}

	return enrollmentKey, nil
}

// ValidateAndUseKey validates an enrollment key and marks it as used
func (em *EnrollmentManager) ValidateAndUseKey(key string) (*models.EnrollmentKey, error) {
	em.mu.Lock()
	defer em.mu.Unlock()

	enrollmentKey, exists := em.keys[key]
	if !exists {
		return nil, fmt.Errorf("invalid enrollment key")
	}

	if time.Now().After(enrollmentKey.ExpiresAt) {
		delete(em.keys, key)
		return nil, fmt.Errorf("enrollment key has expired")
	}

	if enrollmentKey.Used {
		return nil, fmt.Errorf("enrollment key has already been used")
	}

	// Mark as used
	now := time.Now()
	enrollmentKey.Used = true
	enrollmentKey.UsedAt = &now

	return enrollmentKey, nil
}

// ValidateKey checks if a key is valid without marking it as used
func (em *EnrollmentManager) ValidateKey(key string) (*models.EnrollmentKey, error) {
	em.mu.RLock()
	defer em.mu.RUnlock()

	enrollmentKey, exists := em.keys[key]
	if !exists {
		return nil, fmt.Errorf("invalid enrollment key")
	}

	if time.Now().After(enrollmentKey.ExpiresAt) {
		return nil, fmt.Errorf("enrollment key has expired")
	}

	if enrollmentKey.Used {
		return nil, fmt.Errorf("enrollment key has already been used")
	}

	return enrollmentKey, nil
}

// GetKey returns a key by its value (for checking status)
func (em *EnrollmentManager) GetKey(key string) (*models.EnrollmentKey, bool) {
	em.mu.RLock()
	defer em.mu.RUnlock()

	enrollmentKey, exists := em.keys[key]
	return enrollmentKey, exists
}

// GetActiveKeys returns all active (non-expired, non-used) keys
func (em *EnrollmentManager) GetActiveKeys() []*models.EnrollmentKey {
	em.mu.RLock()
	defer em.mu.RUnlock()

	var activeKeys []*models.EnrollmentKey
	now := time.Now()

	for _, key := range em.keys {
		if !key.Used && now.Before(key.ExpiresAt) {
			activeKeys = append(activeKeys, key)
		}
	}

	return activeKeys
}

// RevokeKey revokes an enrollment key, returns true if found and deleted
func (em *EnrollmentManager) RevokeKey(key string) bool {
	em.mu.Lock()
	defer em.mu.Unlock()

	if _, exists := em.keys[key]; !exists {
		return false
	}

	delete(em.keys, key)
	return true
}

// cleanupExpiredKeys removes expired keys every 30 seconds
func (em *EnrollmentManager) cleanupExpiredKeys() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		em.mu.Lock()
		now := time.Now()
		for key, enrollmentKey := range em.keys {
			// Remove expired keys or used keys older than 5 minutes
			if now.After(enrollmentKey.ExpiresAt) {
				delete(em.keys, key)
			} else if enrollmentKey.Used && enrollmentKey.UsedAt != nil && now.Sub(*enrollmentKey.UsedAt) > 5*time.Minute {
				delete(em.keys, key)
			}
		}
		em.mu.Unlock()
	}
}
