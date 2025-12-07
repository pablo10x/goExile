package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strconv"
)

// writeJSON encodes data as JSON and writes it to the ResponseWriter with
// the provided HTTP status code.
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("error encoding response: %v", err)
	}
}

// writeError sends a structured error response using ErrorResponse.
// It also records the error in the global dashboard stats.
func writeError(w http.ResponseWriter, r *http.Request, status int, message string) {
	path := "(unknown)"
	clientIP := "(unknown)"
	if r != nil {
		path = r.URL.Path
		// Strip port from RemoteAddr if present
		host, _, err := net.SplitHostPort(r.RemoteAddr)
		if err == nil {
			clientIP = host
		} else {
			clientIP = r.RemoteAddr
		}
	}
	GlobalStats.RecordError(path, status, message, clientIP)
	writeJSON(w, status, ErrorResponse{Error: message})
}

// parseID converts a string representation of an ID into an integer and
// validates it against a safe range.
func parseID(idStr string) (int, error) {
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return 0, fmt.Errorf("invalid ID format")
	}
	if id < 1 || id > maxIDValue {
		return 0, fmt.Errorf("ID out of valid range")
	}
	return id, nil
}

// decodeJSON decodes a request body into v using a size-limited reader
// and disallowing unknown fields for safety.
func decodeJSON(r *http.Request, v interface{}) error {
	lr := io.LimitReader(r.Body, maxBodySize)
	decoder := json.NewDecoder(lr)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(v); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}
	return nil
}