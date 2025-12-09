package main

import (
	"os"
	"time"
)

// Configuration constants used by the registry. These constants keep the
// runtime behaviour simple and predictable for small deployments.
const (
	// MaxBodySize is the maximum size accepted for request bodies. This
	// protects the server from large or inefficient payloads.
	MaxBodySize = 1 << 20 // 1MB

	// MaxIDValue provides a sanity limit for parsed ID values.
	MaxIDValue = 2147483647

	// ServerTTL defines how long a server is considered alive since its
	// last heartbeat. Servers older than this are removed by cleanup.
	ServerTTL = 60 * time.Second

	// CleanupInterval is how frequently the cleanup loop runs.
	CleanupInterval = 30 * time.Second
)

// GetDBPath returns the database path from the environment or a default value.
func GetDBPath() string {
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "registry.db"
	}
	return dbPath
}