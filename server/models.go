package main

import "time"

// ErrorResponse is a minimal JSON structure used for error payloads.
type ErrorResponse struct {
	Error string `json:"error"`
}

// Spawner instance registered by the Spawner service
type Spawner struct {
	ID               int       `json:"id"`
	Region           string    `json:"region"`
	Host             string    `json:"host"`
	Port             int       `json:"port"`
	MaxInstances     int       `json:"max_instances"`
	CurrentInstances int       `json:"current_instances"`
	Status           string    `json:"status"`
	LastSeen         time.Time `json:"last_seen"`
	// System Metrics
	CpuUsage  float64 `json:"cpu_usage"`  // Percent
	MemUsed   uint64  `json:"mem_used"`   // Bytes
	MemTotal  uint64  `json:"mem_total"`  // Bytes
	DiskUsed  uint64  `json:"disk_used"`  // Bytes
	DiskTotal uint64  `json:"disk_total"` // Bytes
	GameVersion string `json:"game_version"`
}

// GameServerVersion represents a specific uploaded version of the game server package.
type GameServerVersion struct {
	ID         int       `json:"id" db:"id"`
	Filename   string    `json:"filename" db:"filename"`
	Version    string    `json:"version" db:"version"`
	Comment    string    `json:"comment" db:"comment"`
	UploadedAt time.Time `json:"uploaded_at" db:"-"` // Handled via unix timestamp in DB
	IsActive   bool      `json:"is_active" db:"is_active"`
}