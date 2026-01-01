package middleware

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"exile/server/auth"
	"exile/server/logging"
	"exile/server/registry"
	"exile/server/utils"
)

// statusResponseWriter wraps http.ResponseWriter to capture status code, size, and body snippet
type statusResponseWriter struct {
	http.ResponseWriter
	statusCode int
	length     int64
	body       []byte
}

func (w *statusResponseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *statusResponseWriter) Write(b []byte) (int, error) {
	n, err := w.ResponseWriter.Write(b)
	w.length += int64(n)

	// Capture body if status implies error, up to a limit
	if w.statusCode >= 400 && len(w.body) < 1024 {
		limit := 1024 - len(w.body)
		if len(b) < limit {
			limit = len(b)
		}
		w.body = append(w.body, b[:limit]...)
	}

	return n, err
}

// Hijack implements the http.Hijacker interface, required for WebSockets.
func (w *statusResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hijacker, ok := w.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, fmt.Errorf("underlying ResponseWriter does not support Hijacker")
	}
	return hijacker.Hijack()
}

// Flush implements the http.Flusher interface.
func (w *statusResponseWriter) Flush() {
	if flusher, ok := w.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

// StatsMiddleware tracks request bandwidth, status codes, and logs errors.
func StatsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip SSE to allow streaming
		if r.URL.Path == "/events" {
			next.ServeHTTP(w, r)
			return
		}

		// Disable caching for all other API/HTML responses
		w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")

		sw := &statusResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		// Estimate request size (Header + Body)
		reqSize := int64(0)
		if r.ContentLength != -1 {
			reqSize += r.ContentLength
		}
		reqSize += 500 // Overhead

		next.ServeHTTP(sw, r)

		registry.GlobalStats.RecordRequest(sw.statusCode, reqSize, sw.length)

		// Centralized Error Logging
		if sw.statusCode >= 400 {
			message := http.StatusText(sw.statusCode)
			details := ""

			// Try to parse useful message from body
			if len(sw.body) > 0 {
				// Check for JSON error format: {"error": "message"}
				var errResp struct {
					Error string `json:"error"`
				}
				if json.Unmarshal(sw.body, &errResp) == nil && errResp.Error != "" {
					message = errResp.Error
					details = errResp.Error
				} else {
					// Use raw body if it's text/plain and short, otherwise stick to StatusText
					// Simple heuristic: if it contains unprintable chars, ignore it.
					// For now, just take it if it's short.
					bodyStr := string(sw.body)
					if len(bodyStr) < 200 {
						message = strings.TrimSpace(bodyStr)
						details = message
					}
				}
			}

			clientIP := utils.GetClientIP(r)
			category := logging.DetermineCategory(r.URL.Path, sw.statusCode)

			// Log to persistent storage and update stats
			logging.Logger.Log(logging.LogLevelError, category, message, details, r.URL.Path, r.Method, clientIP, sw.statusCode)
		}
	})
}

// APIKeyMiddleware secures Node-Master communication.
func APIKeyMiddleware(apiKey string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			clientKey := r.Header.Get("X-API-Key")
			if clientKey != "" {
				// Check Master Key
				if apiKey != "" && clientKey == apiKey {
					next.ServeHTTP(w, r)
					return
				}
				// Check Node Keys
				if _, valid := registry.GlobalRegistry.ValidateNodeKey(clientKey); valid {
					next.ServeHTTP(w, r)
					return
				}
			}
			utils.WriteError(w, r, http.StatusUnauthorized, "invalid api key")
		})
	}
}

// UnifiedAuthMiddleware allows access if EITHER a valid API Key is provided OR a valid Session exists.
func UnifiedAuthMiddleware(apiKey string, authConfig auth.AuthConfig, sessionStore *auth.SessionStore) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 1. Check API Key (Service-to-Service)
			clientKey := r.Header.Get("X-API-Key")
			if clientKey != "" {
				// Check Master Key
				if apiKey != "" && clientKey == apiKey {
					next.ServeHTTP(w, r)
					return
				}
				// Check Node Keys
				if _, valid := registry.GlobalRegistry.ValidateNodeKey(clientKey); valid {
					next.ServeHTTP(w, r)
					return
				}
			}

			// 2. Check Session (User-to-Service)
			if authConfig.Enabled {
				cookie, err := r.Cookie("session")
				if err == nil {
					isValid, authStep := sessionStore.ValidateSession(cookie.Value)
					if isValid && authStep == auth.AuthStepAuthenticated {
						next.ServeHTTP(w, r)
						return
					}
				}
			}

			// 3. Unauthorized
			utils.WriteError(w, r, http.StatusUnauthorized, "unauthorized: invalid api key or session")
		})
	}
}

// Auth_GameMiddleware secures Game Client communication using a specific Game API Key.
func Auth_GameMiddleware(gameAPIKey string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if gameAPIKey == "" {
				utils.WriteError(w, r, http.StatusServiceUnavailable, "game api key not configured on server")
				return
			}

			clientKey := r.Header.Get("X-Game-API-Key")
			if clientKey != gameAPIKey {
				utils.WriteError(w, r, http.StatusUnauthorized, "invalid game api key")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// GlobalRateLimitMiddleware applies a blanket rate limit to all API requests.
func GlobalRateLimitMiddleware(next http.Handler) http.Handler {
	// Simple in-memory token bucket
	type clientLimiter struct {
		tokens     float64
		lastUpdate time.Time
		mu         sync.Mutex
	}
	
	var (
		limiters = make(map[string]*clientLimiter)
		mu       sync.RWMutex
	)

	// Settings: 100 requests per minute, burst of 20
	const (
		rate  = 100.0 / 60.0 // tokens per second
		burst = 20.0
	)

	go func() {
		// Cleanup routine
		for {
			time.Sleep(10 * time.Minute)
			mu.Lock()
			for ip, lim := range limiters {
				if time.Since(lim.lastUpdate) > 10*time.Minute {
					delete(limiters, ip)
				}
			}
			mu.Unlock()
		}
	}()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip rate limiting for static assets or non-API routes if applied globally
		if !strings.HasPrefix(r.URL.Path, "/api") {
			next.ServeHTTP(w, r)
			return
		}

		ip := utils.GetClientIP(r)

		mu.RLock()
		lim, exists := limiters[ip]
		mu.RUnlock()

		if !exists {
			mu.Lock()
			if lim, exists = limiters[ip]; !exists {
				lim = &clientLimiter{
					tokens:     burst,
					lastUpdate: time.Now(),
				}
				limiters[ip] = lim
			}
			mu.Unlock()
		}

		lim.mu.Lock()
		now := time.Now()
		elapsed := now.Sub(lim.lastUpdate).Seconds()
		lim.tokens += elapsed * rate
		if lim.tokens > burst {
			lim.tokens = burst
		}
		lim.lastUpdate = now

		if lim.tokens >= 1 {
			lim.tokens--
			lim.mu.Unlock()
			next.ServeHTTP(w, r)
		} else {
			lim.mu.Unlock()
			utils.WriteError(w, r, http.StatusTooManyRequests, "global rate limit exceeded")
		}
	})
}

// SecurityHeadersMiddleware adds security-related headers to responses.
func SecurityHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		// CSP allows SvelteKit scripts/styles. Adjust if needed.
		// 'unsafe-inline' is often needed for Svelte unless nonces are strictly used everywhere.
		// 'unsafe-eval' might be needed for some dev tools or specific libs.
		// For now, a reasonable baseline:
		w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; img-src 'self' data: blob:; connect-src 'self' ws: wss:;")

		// HSTS (Strict-Transport-Security) - Should only be sent over HTTPS
		// If behind a proxy that terminates TLS, we might want to set this if X-Forwarded-Proto is https
		if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
			w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}

		next.ServeHTTP(w, r)
	})
}
