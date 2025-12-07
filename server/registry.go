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
// 3. Falls back to in-memory ID generation if DB is unavailable.
func (r *Registry) Register(s *Spawner) int {
	r.mu.Lock()
	defer r.mu.Unlock()

	s.LastSeen = time.Now().UTC()
	s.Status = "active"

	if dbConn != nil {
		s.ID = 0
		id, err := SaveSpawner(dbConn, s)
		if err == nil && id > 0 {
			r.items[id] = s
			if id >= r.nextID {
				r.nextID = id + 1
			}
			return id
		}
		log.Printf("warning: SaveSpawner failed, falling back to in-memory id: %v", err)
	}

	id := r.nextID
	r.nextID++
	s.ID = id
	r.items[id] = s

	if dbConn != nil {
		if _, err := SaveSpawner(dbConn, s); err != nil {
			log.Printf("warning: failed to persist spawner id=%d after fallback: %v", id, err)
		}
	}

	return id
}

// UpdateHeartbeat refreshes the LastSeen timestamp and updates stats.
func (r *Registry) UpdateHeartbeat(id int, currentInstances, maxInstances int, status string) error {
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

	if dbConn != nil {
		if _, err := SaveSpawner(dbConn, s); err != nil {
			log.Printf("warning: failed to persist heartbeat for id=%d: %v", id, err)
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

func (r *Registry) Cleanup(ttl time.Duration, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for range ticker.C {
		cutoff := time.Now().UTC().Add(-ttl)
		r.mu.Lock()
		for id, s := range r.items {
			if s.LastSeen.Before(cutoff) {
				log.Printf("removing expired spawner id=%d region=%s host=%s", id, s.Region, s.Host)
				delete(r.items, id)
				if dbConn != nil {
					if err := DeleteSpawnerDB(dbConn, id); err != nil {
						log.Printf("warning: failed to delete expired spawner id=%d from db: %v", id, err)
					}
				}
			}
		}
		r.mu.Unlock()
	}
}