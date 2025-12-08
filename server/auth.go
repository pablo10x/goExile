package main

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

// AuthConfig holds authentication credentials for the dashboard.
type AuthConfig struct {
	Enabled bool
	Email   string
	Password string
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
	// Check if development mode
	//devMode := os.Getenv("DEV_MODE") != "" || os.Getenv("DEVELOPMENT") != ""
   devMode := true;
	// If development mode is enabled, auth is required
	if !devMode {
		// Production: webpages disabled entirely
		return AuthConfig{Enabled: false}
	}

	// Development: require auth with default credentials
	return AuthConfig{
		Enabled:  true,
		Email:    getEnv("ADMIN_EMAIL", "admin@example.com"),
		Password: getEnv("ADMIN_PASSWORD", "admin123"),
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
			// If auth is disabled, deny all web access
			if !authConfig.Enabled {
				http.Error(w, "Web dashboard is disabled in production mode", http.StatusForbidden)
				return
			}

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
