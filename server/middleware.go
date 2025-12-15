package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"
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

		sw := &statusResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		
		// Estimate request size (Header + Body)
		reqSize := int64(0)
		if r.ContentLength != -1 {
			reqSize += r.ContentLength
		}
		reqSize += 500 // Overhead

		next.ServeHTTP(sw, r)

		GlobalStats.RecordRequest(sw.statusCode, reqSize, sw.length)

		// Centralized Error Logging
		if sw.statusCode >= 400 {
			message := http.StatusText(sw.statusCode)
			
			// Try to parse useful message from body
			if len(sw.body) > 0 {
				// Check for JSON error format: {"error": "message"}
				var errResp struct {
					Error string `json:"error"`
				}
				if json.Unmarshal(sw.body, &errResp) == nil && errResp.Error != "" {
					message = errResp.Error
				} else {
					// Use raw body if it's text/plain and short, otherwise stick to StatusText
					// Simple heuristic: if it contains unprintable chars, ignore it.
					// For now, just take it if it's short.
					bodyStr := string(sw.body)
					if len(bodyStr) < 200 {
						message = strings.TrimSpace(bodyStr)
					}
				}
			}

			clientIP := r.RemoteAddr
			if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
				clientIP = host
			}

			GlobalStats.RecordError(r.URL.Path, sw.statusCode, message, clientIP)
		}
	})
}

// APIKeyMiddleware secures Spawner-Master communication.
func APIKeyMiddleware(apiKey string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if apiKey != "" {
				clientKey := r.Header.Get("X-API-Key")
				if clientKey != apiKey {
					writeError(w, r, http.StatusUnauthorized, "invalid api key")
					return
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

// UnifiedAuthMiddleware allows access if EITHER a valid API Key is provided OR a valid Session exists.
func UnifiedAuthMiddleware(apiKey string, authConfig AuthConfig, sessionStore *SessionStore) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 1. Check API Key (Service-to-Service)
			if apiKey != "" {
				clientKey := r.Header.Get("X-API-Key")
				if clientKey == apiKey {
					next.ServeHTTP(w, r)
					return
				}
			}

			// 2. Check Session (User-to-Service)
			if authConfig.Enabled {
				cookie, err := r.Cookie("session")
				if err == nil {
					isValid, authStep := sessionStore.ValidateSession(cookie.Value)
					if isValid && authStep == AuthStepAuthenticated {
						next.ServeHTTP(w, r)
						return
					}
				}
			}

			// 3. Unauthorized
			writeError(w, r, http.StatusUnauthorized, "unauthorized: invalid api key or session")
		})
	}
}