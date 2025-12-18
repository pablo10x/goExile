package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
	"time"
)

// EnrollmentManager manages temporary enrollment keys
type EnrollmentManager struct {
	keys map[string]*EnrollmentKey
	mu   sync.RWMutex
}

// NewEnrollmentManager creates a new enrollment manager
func NewEnrollmentManager() *EnrollmentManager {
	em := &EnrollmentManager{
		keys: make(map[string]*EnrollmentKey),
	}

	// Start cleanup goroutine
	go em.cleanupExpiredKeys()

	return em
}

// GenerateKey creates a new enrollment key that expires in the given duration
func (em *EnrollmentManager) GenerateKey(createdBy string, duration time.Duration) (*EnrollmentKey, error) {
	em.mu.Lock()
	defer em.mu.Unlock()

	// Generate random key (32 hex characters)
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return nil, fmt.Errorf("failed to generate random key: %w", err)
	}

	key := hex.EncodeToString(bytes)

	enrollmentKey := &EnrollmentKey{
		Key:       key,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(duration),
		CreatedBy: createdBy,
		Used:      false,
	}

	em.keys[key] = enrollmentKey

	return enrollmentKey, nil
}

// ValidateAndUseKey validates an enrollment key and marks it as used
func (em *EnrollmentManager) ValidateAndUseKey(key string) (*EnrollmentKey, error) {
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
func (em *EnrollmentManager) ValidateKey(key string) (*EnrollmentKey, error) {
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
func (em *EnrollmentManager) GetKey(key string) (*EnrollmentKey, bool) {
	em.mu.RLock()
	defer em.mu.RUnlock()

	enrollmentKey, exists := em.keys[key]
	return enrollmentKey, exists
}

// GetActiveKeys returns all active (non-expired, non-used) keys
func (em *EnrollmentManager) GetActiveKeys() []*EnrollmentKey {
	em.mu.RLock()
	defer em.mu.RUnlock()

	var activeKeys []*EnrollmentKey
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
