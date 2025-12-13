package main

import (
	"bufio"
	"fmt"
	"net"
	"net/http"
)

// statusResponseWriter wraps http.ResponseWriter to capture status code and size
type statusResponseWriter struct {
	http.ResponseWriter
	statusCode int
	length     int64
}

func (w *statusResponseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *statusResponseWriter) Write(b []byte) (int, error) {
	n, err := w.ResponseWriter.Write(b)
	w.length += int64(n)
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

// StatsMiddleware tracks request bandwidth and status codes.
func StatsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip SSE to allow streaming
		if r.URL.Path == "/events" {
			next.ServeHTTP(w, r)
			return
		}

		sw := &statusResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		
		// Estimate request size (Header + Body)
		// This is an estimation. Exact wire size is hard to get at this layer.
		reqSize := int64(0)
		if r.ContentLength != -1 {
			reqSize += r.ContentLength
		}
		// Add some overhead for headers/method/url
		reqSize += 500 

		next.ServeHTTP(sw, r)

		GlobalStats.RecordRequest(sw.statusCode, reqSize, sw.length)
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
				if err == nil && sessionStore.ValidateSession(cookie.Value) {
					next.ServeHTTP(w, r)
					return
				}
			}

			// 3. Unauthorized
			// If request had an API key but it was wrong, we already fell through.
			// If request had no session or invalid session, we fall through here.
			writeError(w, r, http.StatusUnauthorized, "unauthorized: invalid api key or session")
		})
	}
}
