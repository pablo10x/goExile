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