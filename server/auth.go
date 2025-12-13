package main

import (
	"crypto/rand"
	"encoding/base64"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// AuthConfig holds authentication credentials for the dashboard.
type AuthConfig struct {
	Enabled        bool
	Email          string
	HashedPassword string
	IsProduction   bool
}

// SessionStore manages active sessions.
type SessionStore struct {
	mu       sync.RWMutex
	sessions map[string]time.Time
}

// NewSessionStore creates a new session store.
func NewSessionStore() *SessionStore {
	return &SessionStore{
		sessions: make(map[string]time.Time),
	}
}

// CreateSession creates a new session token.
func (ss *SessionStore) CreateSession() (string, error) {
	token := make([]byte, 32)
	_, err := rand.Read(token)
	if err != nil {
		return "", err
	}

	sessionID := base64.StdEncoding.EncodeToString(token)

	ss.mu.Lock()
	ss.sessions[sessionID] = time.Now().Add(24 * time.Hour) // 24-hour session
	ss.mu.Unlock()

	return sessionID, nil
}

// ValidateSession checks if a session token is valid.
func (ss *SessionStore) ValidateSession(sessionID string) bool {
	ss.mu.RLock()
	defer ss.mu.RUnlock()

	expiresAt, exists := ss.sessions[sessionID]
	if !exists {
		return false
	}

	// Check if session has expired
	if time.Now().After(expiresAt) {
		// Don't delete here; let cleanup handle it
		return false
	}

	return true
}

// RevokeSession removes a session.
func (ss *SessionStore) RevokeSession(sessionID string) {
	ss.mu.Lock()
	defer ss.mu.Unlock()
	delete(ss.sessions, sessionID)
}

// CleanupExpiredSessions removes expired sessions periodically.
func (ss *SessionStore) CleanupExpiredSessions() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		ss.mu.Lock()
		now := time.Now()
		for sessionID, expiresAt := range ss.sessions {
			if now.After(expiresAt) {
				delete(ss.sessions, sessionID)
			}
		}
		ss.mu.Unlock()
	}
}

// GetAuthConfig returns the auth configuration based on environment.
func GetAuthConfig() AuthConfig {
	isProduction := os.Getenv("PRODUCTION_MODE") == "true"

	if !isProduction {
		log.Println("⚠️  Running in DEVELOPMENT mode. Security features are relaxed.")
		// Dev defaults
		adminPassword := getEnv("ADMIN_PASSWORD", "admin123")
		hashed, _ := bcrypt.GenerateFromPassword([]byte(adminPassword), bcrypt.DefaultCost)
		
		return AuthConfig{
			Enabled:        true,
			Email:          getEnv("ADMIN_EMAIL", "admin@example.com"),
			HashedPassword: string(hashed),
			IsProduction:   false,
		}
	}

	// PRODUCTION MODE
	log.Println("🔒 Running in PRODUCTION mode. Strict security enforced.")

	email := os.Getenv("ADMIN_EMAIL")
	if email == "" {
		log.Fatal("FATAL: ADMIN_EMAIL must be set in production mode")
	}

	password := os.Getenv("ADMIN_PASSWORD")
	passwordHash := os.Getenv("ADMIN_PASSWORD_HASH")

	var finalHash string

	if passwordHash != "" {
		finalHash = passwordHash
	} else if password != "" {
		log.Println("⚠️  ADMIN_PASSWORD used in plaintext. Consider using ADMIN_PASSWORD_HASH for better security.")
		hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			log.Fatalf("FATAL: Failed to hash password: %v", err)
		}
		finalHash = string(hashed)
	} else {
		log.Fatal("FATAL: ADMIN_PASSWORD or ADMIN_PASSWORD_HASH must be set in production mode")
	}

	return AuthConfig{
		Enabled:        true,
		Email:          email,
		HashedPassword: finalHash,
		IsProduction:   true,
	}
}

// getEnv gets an environment variable with a default value.
func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

// AuthMiddleware checks if user is authenticated before serving dashboard.
func AuthMiddleware(authConfig AuthConfig, sessionStore *SessionStore) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// In production, force HTTPS redirects if not behind a proxy handling it? 
			// For now, we focus on cookie security.

			// Check for valid session cookie
			cookie, err := r.Cookie("session")
			if err == nil && sessionStore.ValidateSession(cookie.Value) {
				// Valid session, proceed
				next.ServeHTTP(w, r)
				return
			}

			// No valid session
			if strings.HasPrefix(r.URL.Path, "/api") {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			
			// Redirect to login for pages
			http.Redirect(w, r, "/login", http.StatusSeeOther)
		})
	}
}
