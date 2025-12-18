package main

import (
	"crypto/rand"
	"encoding/base64"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"
)

const (
	AuthStepAuthenticated = "authenticated"
	AuthStepTOTP          = "totp"
	AuthStepEmail         = "email"
)

// AuthConfig holds authentication credentials for the dashboard.
type AuthConfig struct {
	Enabled        bool
	Email          string
	HashedPassword string
	TOTPSecret     string
	IsProduction   bool

	// SMTP Settings
	SMTPHost string
	SMTPPort string
	SMTPUser string
	SMTPPass string
	SMTPFrom string
}

// SessionData holds session information.
type SessionData struct {
	Expiry    time.Time
	AuthStep  string // "authenticated", "totp", "email"
	EmailCode string
}

// SessionStore manages active sessions.
type SessionStore struct {
	mu       sync.RWMutex
	sessions map[string]SessionData
}

// NewSessionStore creates a new session store.
func NewSessionStore() *SessionStore {
	return &SessionStore{
		sessions: make(map[string]SessionData),
	}
}

// CreateSession creates a new session token.
func (ss *SessionStore) CreateSession(initialStep string) (string, error) {
	token := make([]byte, 32)
	_, err := rand.Read(token)
	if err != nil {
		return "", err
	}

	sessionID := base64.StdEncoding.EncodeToString(token)

	ss.mu.Lock()
	ss.sessions[sessionID] = SessionData{
		Expiry:   time.Now().Add(24 * time.Hour), // 24-hour session
		AuthStep: initialStep,
	}
	ss.mu.Unlock()

	return sessionID, nil
}

// ValidateSession checks if a session token is valid and returns its current auth step.
func (ss *SessionStore) ValidateSession(sessionID string) (isValid bool, authStep string) {
	ss.mu.RLock()
	defer ss.mu.RUnlock()

	data, exists := ss.sessions[sessionID]
	if !exists {
		return false, ""
	}

	// Check if session has expired
	if time.Now().After(data.Expiry) {
		return false, ""
	}

	return true, data.AuthStep
}

// SetSessionStep updates the authentication step for a session.
func (ss *SessionStore) SetSessionStep(sessionID string, step string) {
	ss.mu.Lock()
	defer ss.mu.Unlock()
	if data, exists := ss.sessions[sessionID]; exists {
		data.AuthStep = step
		ss.sessions[sessionID] = data
	}
}

// SetEmailCode stores the verification code for the email step.
func (ss *SessionStore) SetEmailCode(sessionID string, code string) {
	ss.mu.Lock()
	defer ss.mu.Unlock()
	if data, exists := ss.sessions[sessionID]; exists {
		data.EmailCode = code
		ss.sessions[sessionID] = data
	}
}

// VerifyEmailCode checks if the provided code matches the stored one.
func (ss *SessionStore) VerifyEmailCode(sessionID string, code string) bool {
	ss.mu.RLock()
	defer ss.mu.RUnlock()
	if data, exists := ss.sessions[sessionID]; exists {
		return data.EmailCode == code && code != ""
	}
	return false
}

// MarkSessionAuthenticated marks a session as fully authenticated.
func (ss *SessionStore) MarkSessionAuthenticated(sessionID string) {
	ss.mu.Lock()
	defer ss.mu.Unlock()
	if data, exists := ss.sessions[sessionID]; exists {
		data.AuthStep = AuthStepAuthenticated
		ss.sessions[sessionID] = data
	}
}

// RevokeSession removes a session.
func (ss *SessionStore) RevokeSession(sessionID string) {
	ss.mu.Lock()
	defer ss.mu.Unlock()
	delete(ss.sessions, sessionID)
}

// CleanupExpired sessions removes expired sessions periodically.
func (ss *SessionStore) CleanupExpiredSessions() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		ss.mu.Lock()
		now := time.Now()
		for sessionID, data := range ss.sessions {
			if now.After(data.Expiry) {
				delete(ss.sessions, sessionID)
			}
		}
		ss.mu.Unlock()
	}
}

// RateLimiter manages request rate limiting.
type RateLimiter struct {
	mu          sync.Mutex
	attempts    map[string]rateLimitEntry
	maxAttempts int
	window      time.Duration
}

type rateLimitEntry struct {
	count    int
	firstTry time.Time
}

// NewRateLimiter creates a new rate limiter.
func NewRateLimiter(maxAttempts int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		attempts:    make(map[string]rateLimitEntry),
		maxAttempts: maxAttempts,
		window:      window,
	}
}

// Allow checks if the action is allowed for the given key.
func (rl *RateLimiter) Allow(key string) (bool, int) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	entry, exists := rl.attempts[key]

	// If entry exists and window has passed, reset
	if exists && time.Since(entry.firstTry) > rl.window {
		delete(rl.attempts, key)
		exists = false
	}

	if !exists {
		rl.attempts[key] = rateLimitEntry{
			count:    1,
			firstTry: time.Now(),
		}
		return true, 1
	}

	if entry.count >= rl.maxAttempts {
		entry.count++
		rl.attempts[key] = entry
		return false, entry.count
	}

	entry.count++
	rl.attempts[key] = entry
	return true, entry.count
}

// Reset clears the limit for a key (e.g. on success).
func (rl *RateLimiter) Reset(key string) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	delete(rl.attempts, key)
}

// Global Limiters
var (
	LoginRateLimiter     = NewRateLimiter(5, 15*time.Minute) // 5 attempts per 15 mins
	TwoFactorRateLimiter = NewRateLimiter(3, 15*time.Minute) // 3 attempts per 15 mins
	EmailCodeRateLimiter = NewRateLimiter(3, 15*time.Minute) // 3 attempts per 15 mins for email code
)

// GetAuthConfig returns the auth configuration based on environment.
func GetAuthConfig() AuthConfig {
	isProduction := os.Getenv("PRODUCTION_MODE") == "true"

	// Common Config
	cfg := AuthConfig{
		Enabled:      true,
		IsProduction: isProduction,
		SMTPHost:     getEnv("SMTP_HOST", ""),
		SMTPPort:     getEnv("SMTP_PORT", "587"),
		SMTPUser:     getEnv("SMTP_USER", ""),
		SMTPPass:     getEnv("SMTP_PASS", ""),
		SMTPFrom:     getEnv("SMTP_FROM", ""),
	}

	// Override with SMTP_URL if present (e.g. smtp://user:pass@host:port)
	if smtpURL := os.Getenv("SMTP_URL"); smtpURL != "" {
		u, err := url.Parse(smtpURL)
		if err == nil {
			cfg.SMTPHost = u.Hostname()
			if p := u.Port(); p != "" {
				cfg.SMTPPort = p
			}
			cfg.SMTPUser = u.User.Username()
			if p, ok := u.User.Password(); ok {
				cfg.SMTPPass = p
			}
		}
	}

	if !isProduction {
		// Dev defaults
		adminPassword := getEnv("ADMIN_PASSWORD", "admin123")
		hashed, _ := bcrypt.GenerateFromPassword([]byte(adminPassword), bcrypt.DefaultCost)

		cfg.Email = getEnv("ADMIN_EMAIL", "admin@example.com")
		cfg.HashedPassword = string(hashed)
		cfg.TOTPSecret = getEnv("ADMIN_2FA_SECRET", "")

		return cfg
	}

	// PRODUCTION MODE
	cfg.Email = os.Getenv("ADMIN_EMAIL")
	if cfg.Email == "" {
		log.Fatal("FATAL: ADMIN_EMAIL must be set in production mode")
	}

	password := os.Getenv("ADMIN_PASSWORD")
	passwordHash := os.Getenv("ADMIN_PASSWORD_HASH")

	if passwordHash != "" {
		cfg.HashedPassword = passwordHash
	} else if password != "" {
		hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			log.Fatalf("FATAL: Failed to hash password: %v", err)
		}
		cfg.HashedPassword = string(hashed)
	} else {
		log.Fatal("FATAL: ADMIN_PASSWORD or ADMIN_PASSWORD_HASH must be set in production mode")
	}

	cfg.TOTPSecret = os.Getenv("ADMIN_2FA_SECRET")

	return cfg
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
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
			// Check for valid session cookie
			cookie, err := r.Cookie("session")
			if err == nil {
				isValid, authStep := sessionStore.ValidateSession(cookie.Value)
				if isValid {
					// Check Authentication Step
					if authStep == AuthStepAuthenticated {
						// Fully authenticated
						if r.URL.Path == "/login/2fa" {
							// Already authenticated, go to dashboard
							http.Redirect(w, r, "/", http.StatusSeeOther)
							return
						}
						next.ServeHTTP(w, r)
						return
					}

					// Pending Steps (TOTP or Email)
					// Allow access to 2FA page and API endpoints for verification
					allowedPaths := []string{
						"/login/2fa",
						"/api/login/2fa",   // TOTP Verify
						"/login/email",     // If separate page (optional)
						"/api/login/email", // Email Verify
						"/_app",            // Static assets
					}

					isAllowed := false
					for _, path := range allowedPaths {
						if r.URL.Path == path || strings.HasPrefix(r.URL.Path, path) {
							isAllowed = true
							break
						}
					}

					if isAllowed {
						next.ServeHTTP(w, r)
						return
					}

					// Redirect everything else to 2FA page
					if strings.HasPrefix(r.URL.Path, "/api") {
						http.Error(w, "Unauthorized: Verification Required", http.StatusUnauthorized)
						return
					}

					// Both TOTP and Email steps happen on the /login/2fa page (UI expands)
					http.Redirect(w, r, "/login/2fa", http.StatusSeeOther)
					return
				}
			}

			// No valid session
			if strings.HasPrefix(r.URL.Path, "/api") {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// Redirect to login for pages
			// Allow access to /login
			if strings.HasPrefix(r.URL.Path, "/login") {
				next.ServeHTTP(w, r)
				return
			}

			http.Redirect(w, r, "/login", http.StatusSeeOther)
		})
	}
}
