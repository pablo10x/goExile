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
}

// GameServerVersion represents a specific uploaded version of the game server package.
type GameServerVersion struct {
	ID         int       `json:"id" db:"id"`
	Filename   string    `json:"filename" db:"filename"`
	Comment    string    `json:"comment" db:"comment"`
	UploadedAt time.Time `json:"uploaded_at" db:"-"` // Handled via unix timestamp in DB
	IsActive   bool      `json:"is_active" db:"is_active"`
}