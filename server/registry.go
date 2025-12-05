package main

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/jmoiron/sqlx"
)

// Registry manages the lifecycle of Server instances in memory. It is
// safe for concurrent use and exposes simple CRUD-style operations used
// by the HTTP handlers.
type Registry struct {
	mu     sync.RWMutex
	nextID int
	items  map[int]*Server
}

var registry = &Registry{
	nextID: 1,
	items:  make(map[int]*Server),
}

// dbConn holds the optional SQLite DB connection when persistence is enabled.
var dbConn *sqlx.DB

// Register adds a new server to the registry, assigning a unique numeric
// ID and setting initial metadata such as LastSeen and Status.
// The function returns the assigned server ID.
func (r *Registry) Register(s *Server) int {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Prefer DB-assigned ID when persistence is enabled. This avoids
	// collision in multi-instance scenarios and makes the DB the source
	// of truth for ID allocation. We set metadata first and attempt an
	// INSERT; SaveServer will return the assigned id.
	s.LastSeen = time.Now().UTC()
	s.Status = "active"

	if dbConn != nil {
		// ensure we don't accidentally send a pre-filled ID for insert
		s.ID = 0
		id, err := SaveServer(dbConn, s)
		if err == nil && id > 0 {
			// Persisted successfully; record in-memory state.
			r.items[id] = s
			if id >= r.nextID {
				r.nextID = id + 1
			}
			return id
		}
		log.Printf("warning: SaveServer failed, falling back to in-memory id: %v", err)
		// fall through to in-memory assignment
	}

	// In-memory fallback (single-process)
	id := r.nextID
	r.nextID++
	s.ID = id
	r.items[id] = s

	// Best-effort persistence after fallback
	if dbConn != nil {
		if _, err := SaveServer(dbConn, s); err != nil {
			log.Printf("warning: failed to persist server id=%d after fallback: %v", id, err)
		}
	}

	return id
}

// UpdateHeartbeat refreshes the LastSeen timestamp for the server
// identified by id. If the server does not exist, an error is returned.
func (r *Registry) UpdateHeartbeat(id int) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	s, ok := r.items[id]
	if !ok {
		return fmt.Errorf("server not found")
	}
	s.LastSeen = time.Now().UTC()
	s.Status = "active"

	if dbConn != nil {
		if _, err := SaveServer(dbConn, s); err != nil {
			log.Printf("warning: failed to persist heartbeat for id=%d: %v", id, err)
		}
	}

	return nil
}

// Get returns a pointer to the Server and a boolean indicating whether it
// was found in the registry. The returned pointer references the in-memory
// object; callers should not modify it without taking care to avoid races.
func (r *Registry) Get(id int) (*Server, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	s, ok := r.items[id]
	return s, ok
}

// List returns a snapshot slice of all registered servers. The function
// copies values out of the internal map to avoid exposing internal state.
func (r *Registry) List() []Server {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]Server, 0, len(r.items))
	for _, v := range r.items {
		out = append(out, *v)
	}
	return out
}

// Delete removes the server with the given id from the registry. It
// returns true when a server was deleted, false if no server with that id
// existed.
func (r *Registry) Delete(id int) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.items[id]; !ok {
		return false
	}
	delete(r.items, id)

	if dbConn != nil {
		if err := DeleteServerDB(dbConn, id); err != nil {
			log.Printf("warning: failed to delete server id=%d from db: %v", id, err)
		}
	}

	return true
}

// Cleanup runs in a background goroutine and periodically removes servers
// whose LastSeen is older than the configured ttl. This prevents stale
// entries from accumulating after servers crash or lose network.
func (r *Registry) Cleanup(ttl time.Duration, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for range ticker.C {
		cutoff := time.Now().UTC().Add(-ttl)
		r.mu.Lock()
		for id, s := range r.items {
			if s.LastSeen.Before(cutoff) {
				log.Printf("removing expired server id=%d name=%s", id, s.Name)
				delete(r.items, id)
				if dbConn != nil {
					if err := DeleteServerDB(dbConn, id); err != nil {
						log.Printf("warning: failed to delete expired server id=%d from db: %v", id, err)
					}
				}
			}
		}
		r.mu.Unlock()
	}
}
