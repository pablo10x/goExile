package main

import "time"

// Server instance registered by Unity game servers
type Server struct {
	ID             int       `json:"id"`
	Name           string    `json:"name"`
	Host           string    `json:"host"`
	Port           int       `json:"port"`
	MaxPlayers     int       `json:"max_players"`
	CurrentPlayers int       `json:"current_players"`
	Region         string    `json:"region"`
	Status         string    `json:"status"`
	LastSeen       time.Time `json:"last_seen"`
}

// ErrorResponse is a minimal JSON structure used for error payloads.
type ErrorResponse struct {
	Error string `json:"error"`
}
