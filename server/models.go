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
	CpuUsage    float64 `json:"cpu_usage"`  // Percent
	MemUsed     uint64  `json:"mem_used"`   // Bytes
	MemTotal    uint64  `json:"mem_total"`  // Bytes
	DiskUsed    uint64  `json:"disk_used"`  // Bytes
	DiskTotal   uint64  `json:"disk_total"` // Bytes
	GameVersion string  `json:"game_version"`
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

// ServerConfig represents a configuration setting for the server.
type ServerConfig struct {
	ID              int       `json:"id" db:"id"`
	Key             string    `json:"key" db:"key"`
	Value           string    `json:"value" db:"value"`
	Type            string    `json:"type" db:"type"`         // string, int, bool, duration
	Category        string    `json:"category" db:"category"` // system, spawner, security
	Description     string    `json:"description" db:"description"`
	IsReadOnly      bool      `json:"is_read_only" db:"is_read_only"`
	RequiresRestart bool      `json:"requires_restart" db:"requires_restart"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
	UpdatedBy       string    `json:"updated_by" db:"updated_by"`
}

// InstanceAction represents a recorded action performed on an instance.
type InstanceAction struct {
	ID         int       `json:"id" db:"id"`
	SpawnerID  int       `json:"spawner_id" db:"spawner_id"`
	InstanceID string    `json:"instance_id" db:"instance_id"`
	Action     string    `json:"action" db:"action"`
	Timestamp  time.Time `json:"timestamp" db:"timestamp"`
	Status     string    `json:"status" db:"status"` // "success" or "failed"
	Details    string    `json:"details" db:"details"`
}
