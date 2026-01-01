package auth

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"exile/server/utils"

	"github.com/pquerna/otp/totp"
	"golang.org/x/crypto/bcrypt"
)

const (
	AuthStepAuthenticated = "authenticated"
	AuthStepTOTP          = "totp"
	AuthStepEmail         = "email"
)

type AuthConfig struct {
	Enabled        bool
	Email          string
	HashedPassword string
	TOTPSecret     string
	IsProduction   bool
	SMTPHost       string
	SMTPPort       string
	SMTPUser       string
	SMTPPass       string
	SMTPFrom       string
}

type SessionData struct {
	Expiry     time.Time
	LastActive time.Time
	AuthStep   string
	EmailCode  string
}

type SessionStore struct {
	mu               sync.RWMutex
	sessions         map[string]SessionData
	maxSessions      int
	inactivityTimeout time.Duration
}

func NewSessionStore(isProduction bool) *SessionStore {
	max := 2
	if isProduction {
		max = 1
	}
	return &SessionStore{
		sessions:          make(map[string]SessionData),
		maxSessions:       max,
		inactivityTimeout: 1 * time.Hour,
	}
}

// CreateSession creates a new authentication session.
func (ss *SessionStore) CreateSession(initialStep string) (string, error) {
	token := make([]byte, 32)
	if _, err := rand.Read(token); err != nil {
		return "", fmt.Errorf("failed to generate session token: %w", err)
	}
	sessionID := base64.StdEncoding.EncodeToString(token)
	ss.mu.Lock()
	defer ss.mu.Unlock()
	ss.sessions[sessionID] = SessionData{
		Expiry:     time.Now().Add(24 * time.Hour),
		LastActive: time.Now(),
		AuthStep:   initialStep,
	}
	return sessionID, nil
}

func (ss *SessionStore) ValidateSession(sessionID string) (bool, string) {
	ss.mu.Lock()
	defer ss.mu.Unlock()
	data, exists := ss.sessions[sessionID]
	if !exists || time.Now().After(data.Expiry) {
		return false, ""
	}

	// Check inactivity timeout
	if time.Since(data.LastActive) > ss.inactivityTimeout {
		delete(ss.sessions, sessionID)
		return false, ""
	}

	// Update last active time
	data.LastActive = time.Now()
	ss.sessions[sessionID] = data

	return true, data.AuthStep
}

func (ss *SessionStore) MarkSessionAuthenticated(sessionID string) {
	ss.mu.Lock()
	defer ss.mu.Unlock()
	if data, exists := ss.sessions[sessionID]; exists {
		data.AuthStep = AuthStepAuthenticated
		ss.sessions[sessionID] = data
	}
}

func (ss *SessionStore) RevokeSession(sessionID string) {
	ss.mu.Lock()
	defer ss.mu.Unlock()
	delete(ss.sessions, sessionID)
}

func (ss *SessionStore) CleanupExpiredSessions() {
	ss.mu.Lock()
	defer ss.mu.Unlock()
	now := time.Now()
	for id, data := range ss.sessions {
		if now.After(data.Expiry) {
			delete(ss.sessions, id)
		}
	}
}

type RateLimiter struct {
	mu          sync.Mutex
	attempts    map[string]int
	maxAttempts int
}

func NewRateLimiter(max int, window time.Duration) *RateLimiter {
	return &RateLimiter{attempts: make(map[string]int), maxAttempts: max}
}

func (rl *RateLimiter) Allow(key string) (bool, int) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	rl.attempts[key]++
	return rl.attempts[key] <= rl.maxAttempts, rl.attempts[key]
}

func (rl *RateLimiter) Reset(key string) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	delete(rl.attempts, key)
}

var (
	LoginRateLimiter     = NewRateLimiter(5, 15*time.Minute)
	TwoFactorRateLimiter = NewRateLimiter(3, 15*time.Minute)
)

func GetAuthConfig() AuthConfig {
	isProd := os.Getenv("PRODUCTION_MODE") == "true"
	hashed, _ := bcrypt.GenerateFromPassword([]byte(utils.GetEnv("ADMIN_PASSWORD", "admin123")), bcrypt.DefaultCost)
	return AuthConfig{
		Enabled:        true,
		Email:          utils.GetEnv("ADMIN_EMAIL", "admin@example.com"),
		HashedPassword: string(hashed),
		TOTPSecret:     utils.GetEnv("ADMIN_2FA_SECRET", ""),
		IsProduction:   isProd,
	}
}

func HandleLogin(w http.ResponseWriter, r *http.Request, cfg AuthConfig, ss *SessionStore) {
	ip := utils.GetClientIP(r)
	if allowed, _ := LoginRateLimiter.Allow(ip); !allowed {
		http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
		return
	}
	email := r.FormValue("email")
	password := r.FormValue("password")
	if email == cfg.Email && bcrypt.CompareHashAndPassword([]byte(cfg.HashedPassword), []byte(password)) == nil {
		LoginRateLimiter.Reset(ip)
		step := AuthStepAuthenticated
		if cfg.TOTPSecret != "" && cfg.IsProduction {
			step = AuthStepTOTP
		}
		sid, _ := ss.CreateSession(step)
		http.SetCookie(w, &http.Cookie{
			Name:     "session",
			Value:    sid,
			Path:     "/",
			HttpOnly: true,
			Secure:   cfg.IsProduction,
			SameSite: http.SameSiteStrictMode,
		})

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{
			"status":    "ok",
			"next_step": step,
		})
		return
	}
	http.Error(w, "Invalid credentials", http.StatusUnauthorized)
}

func Handle2FAVerify(w http.ResponseWriter, r *http.Request, cfg AuthConfig, ss *SessionStore) {
	cookie, _ := r.Cookie("session")
	if cookie == nil {
		http.Error(w, "No session", http.StatusUnauthorized)
		return
	}
	valid, step := ss.ValidateSession(cookie.Value)
	if !valid || step != AuthStepTOTP {
		http.Error(w, "Invalid session or step", http.StatusUnauthorized)
		return
	}
	code := r.FormValue("code")
	if totp.Validate(code, cfg.TOTPSecret) {
		ss.MarkSessionAuthenticated(cookie.Value)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
		return
	}
	http.Error(w, "Invalid code", http.StatusUnauthorized)
}

func HandleEmailVerify(w http.ResponseWriter, r *http.Request, cfg AuthConfig, ss *SessionStore) {
	http.Error(w, "Not implemented", http.StatusNotImplemented)
}

func HandleLogout(w http.ResponseWriter, r *http.Request, ss *SessionStore) {
	if c, _ := r.Cookie("session"); c != nil {
		ss.RevokeSession(c.Value)
	}
	http.SetCookie(w, &http.Cookie{Name: "session", Value: "", Path: "/", MaxAge: -1})
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func AuthMiddleware(cfg AuthConfig, ss *SessionStore) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if c, _ := r.Cookie("session"); c != nil {
				if valid, step := ss.ValidateSession(c.Value); valid && step == AuthStepAuthenticated {
					next.ServeHTTP(w, r)
					return
				}
			}
			if strings.HasPrefix(r.URL.Path, "/api") {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			if strings.HasPrefix(r.URL.Path, "/api/auth") {
				next.ServeHTTP(w, r)
				return
			}
			http.Redirect(w, r, "/login", http.StatusSeeOther)
		})
	}
}
