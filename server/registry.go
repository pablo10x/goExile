package main

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/jmoiron/sqlx"
)

// Registry manages the lifecycle of Spawner instances in memory.
// It acts as a write-through cache when persistence is enabled.
type Registry struct {
	mu     sync.RWMutex
	nextID int
	items  map[int]*Spawner
}

var registry = &Registry{
	nextID: 1,
	items:  make(map[int]*Spawner),
}

// dbConn holds the optional SQLite DB connection.
var dbConn *sqlx.DB

// Register adds a new spawner to the registry.
//
// Logic:
// 1. Sets initial metadata (LastSeen, Status).
// 2. Attempts to persist to DB first to get an authoritative ID.
// 3. Fails if DB is enabled but write fails.
// 4. Uses in-memory ID generation ONLY if DB is not configured (nil).
func (r *Registry) Register(s *Spawner) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	s.LastSeen = time.Now().UTC()
	s.Status = "Online"

	if dbConn != nil {
		s.ID = 0 // Ensure ID is 0 so DB assigns a new one
		id, err := SaveSpawner(dbConn, s)
		if err != nil {
			return 0, fmt.Errorf("failed to save spawner to DB: %w", err)
		}
		
		// Refresh with authoritative ID
		s.ID = id
		r.items[id] = s
		
		// Sync in-memory counter just in case
		if id >= r.nextID {
			r.nextID = id + 1
		}
		return id, nil
	}

	// In-memory mode only (no DB configured)
	id := r.nextID
	r.nextID++
	s.ID = id
	r.items[id] = s

	return id, nil
}

// UpdateHeartbeat refreshes the LastSeen timestamp and updates stats.
func (r *Registry) UpdateHeartbeat(id int, currentInstances, maxInstances int, status string, cpuUsage float64, memUsed, memTotal, diskUsed, diskTotal uint64, gameVersion string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	s, ok := r.items[id]
	if !ok {
		return fmt.Errorf("spawner not found")
	}
	s.LastSeen = time.Now().UTC()
	s.CurrentInstances = currentInstances
	s.MaxInstances = maxInstances
	s.Status = status
	
	s.CpuUsage = cpuUsage
	s.MemUsed = memUsed
	s.MemTotal = memTotal
	s.DiskUsed = diskUsed
	s.DiskTotal = diskTotal
	s.GameVersion = gameVersion

	if dbConn != nil {
		if _, err := SaveSpawner(dbConn, s); err != nil {
			log.Printf("warning: failed to persist heartbeat for id=%d: %v", id, err)
		}
	}

	return nil
}

// UpdateSpawnerStatus updates the status of a spawner and its last seen time.
func (r *Registry) UpdateSpawnerStatus(id int, newStatus string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	s, ok := r.items[id]
	if !ok {
		return fmt.Errorf("spawner not found")
	}
	s.Status = newStatus
	s.LastSeen = time.Now().UTC() // Update LastSeen as well
	if dbConn != nil {
		if _, err := SaveSpawner(dbConn, s); err != nil {
			log.Printf("warning: failed to persist status update for id=%d: %v", id, err)
		}
	}
	return nil
}

func (r *Registry) Get(id int) (*Spawner, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	s, ok := r.items[id]
	return s, ok
}

func (r *Registry) List() []Spawner {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]Spawner, 0, len(r.items))
	for _, v := range r.items {
		out = append(out, *v)
	}
	return out
}

func (r *Registry) Delete(id int) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.items[id]; !ok {
		return false
	}
	delete(r.items, id)

	if dbConn != nil {
		if err := DeleteSpawnerDB(dbConn, id); err != nil {
			log.Printf("warning: failed to delete spawner id=%d from db: %v", id, err)
		}
	}

	return true
}

// MonitorStatuses updates spawner statuses based on last seen time.
func (r *Registry) MonitorStatuses(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		r.mu.Lock()
		// now := time.Now().UTC()
		for _, s := range r.items {
			// If Offline (WS disconnected), don't change based on time
			if s.Status == "Offline" {
				continue
			}

			// Time-based status updates disabled (Degraded/Unresponsive).
			// Status is only "Online" (connected) or "Offline" (disconnected).
			/*
			since := now.Sub(s.LastSeen)
			
			var newStatus string
			// Heartbeat is 5s. 2 heartbeats = 10s.
			// Increased buffer to 20s (4 missed heartbeats) to prevent flapping
			if since < 20*time.Second { 
				newStatus = "Online"
			} else if since < 45*time.Second {
				newStatus = "Degraded"
			} else {
				newStatus = "Unresponsive"
			}

			// Only update if changed
			if s.Status != newStatus {
				s.Status = newStatus
			}
			*/
		}
		r.mu.Unlock()
	}
}

func (r *Registry) Cleanup(ttl time.Duration, interval time.Duration) {
	// Automatic cleanup disabled as per requirement.
	// Spawners persist until explicitly deleted.
}