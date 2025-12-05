package main

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"

	// Use the pure-Go SQLite driver to avoid CGO dependency in builds.
	_ "modernc.org/sqlite"
)

// db.go provides a minimal persistence layer for the registry. The
// helpers are intentionally small and focused on reliability: open the
// database, ensure the schema exists, and provide simple CRUD helpers for
// Server records. The implementation uses SQL transactions for writes and
// stores timestamps as UNIX seconds for portability.

// InitDB opens (or creates) an SQLite database at the given path and
// ensures the servers table exists. The function also configures PRAGMAs
// to set a busy timeout and enable foreign keys, which improves behavior
// under contention and mirrors typical SQLite deployments.
func InitDB(path string) (*sqlx.DB, error) {
	// Open using the modernc driver (driver name is "sqlite").
	db, err := sqlx.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	// Set reasonable connection parameters for a small local DB.
	db.SetMaxOpenConns(1)
	db.SetConnMaxLifetime(time.Minute * 5)

	// Tune PRAGMAs for production-friendly defaults:
	// - busy_timeout: wait briefly when DB is locked
	// - foreign_keys: enforce FK constraints
	// - journal_mode = WAL: allow concurrent readers/writers
	// - synchronous = NORMAL: good durability/performance tradeoff
	if _, err := db.Exec("PRAGMA busy_timeout = 5000"); err != nil {
		db.Close()
		return nil, fmt.Errorf("set busy_timeout: %w", err)
	}
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		db.Close()
		return nil, fmt.Errorf("enable foreign_keys: %w", err)
	}
	if _, err := db.Exec("PRAGMA journal_mode = WAL"); err != nil {
		db.Close()
		return nil, fmt.Errorf("enable WAL journal: %w", err)
	}
	if _, err := db.Exec("PRAGMA synchronous = NORMAL"); err != nil {
		db.Close()
		return nil, fmt.Errorf("set synchronous: %w", err)
	}

	if err := createTables(db); err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}

// createTables creates the minimal schema required by the registry. The
// function is idempotent and safe to call on every startup.
func createTables(db *sqlx.DB) error {
	q := `CREATE TABLE IF NOT EXISTS servers (
		id INTEGER PRIMARY KEY,
		name TEXT NOT NULL,
		host TEXT NOT NULL,
		port INTEGER NOT NULL,
		max_players INTEGER NOT NULL,
		current_players INTEGER NOT NULL,
		region TEXT,
		status TEXT,
		last_seen INTEGER NOT NULL
	);`
	_, err := db.Exec(q)
	if err != nil {
		return fmt.Errorf("create tables: %w", err)
	}
	return nil
}

// CloseDB performs a checkpoint (to flush WAL) and closes the DB.
func CloseDB(db *sqlx.DB) error {
	if db == nil {
		return nil
	}
	// Try to checkpoint WAL to reduce WAL file growth.
	if _, err := db.Exec("PRAGMA wal_checkpoint(TRUNCATE)"); err != nil {
		// Non-fatal, log and continue with close.
		// Caller can decide how to handle.
	}
	return db.Close()
}

// execWithRetry executes the provided function with a simple retry/backoff
// strategy for transient errors (e.g., SQLITE_BUSY). It uses a small
// exponential backoff and a short deadline.
func execWithRetry(fn func() error) error {
	var lastErr error
	backoff := 50 * time.Millisecond
	for i := 0; i < 5; i++ {
		if err := fn(); err != nil {
			lastErr = err
			// If it's likely transient, sleep and retry.
			time.Sleep(backoff)
			backoff *= 2
			continue
		}
		return nil
	}
	return lastErr
}

// SaveServer inserts or replaces a server record. If s.ID == 0 the DB will
// assign an ID which is returned as the last insert id. If s.ID != 0 it will
// upsert by id.
func SaveServer(db *sqlx.DB, s *Server) (int, error) {
	if s == nil {
		return 0, fmt.Errorf("nil server")
	}
	// Use a transaction for safety. Wrap in retry to handle SQLITE_BUSY
	var assignedID int
	do := func() error {
		tx, err := db.Begin()
		if err != nil {
			return fmt.Errorf("begin tx: %w", err)
		}
		defer tx.Rollback()

		if s.ID == 0 {
			res, err := tx.Exec(`INSERT INTO servers (name, host, port, max_players, current_players, region, status, last_seen) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
				s.Name, s.Host, s.Port, s.MaxPlayers, s.CurrentPlayers, s.Region, s.Status, s.LastSeen.Unix())
			if err != nil {
				return fmt.Errorf("insert server: %w", err)
			}
			id, err := res.LastInsertId()
			if err != nil {
				return fmt.Errorf("last insert id: %w", err)
			}
			assignedID = int(id)
			if err := tx.Commit(); err != nil {
				return fmt.Errorf("commit: %w", err)
			}
			return nil
		}

		// Upsert by id (replace)
		_, err = tx.Exec(`REPLACE INTO servers (id, name, host, port, max_players, current_players, region, status, last_seen) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			s.ID, s.Name, s.Host, s.Port, s.MaxPlayers, s.CurrentPlayers, s.Region, s.Status, s.LastSeen.Unix())
		if err != nil {
			return fmt.Errorf("upsert server: %w", err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("commit: %w", err)
		}
		assignedID = s.ID
		return nil
	}

	if err := execWithRetry(do); err != nil {
		return 0, err
	}
	return assignedID, nil
}

// LoadServers returns all servers stored in the DB.
func LoadServers(db *sqlx.DB) ([]Server, error) {
	rows, err := db.Queryx(`SELECT id, name, host, port, max_players, current_players, region, status, last_seen FROM servers`)
	if err != nil {
		return nil, fmt.Errorf("query servers: %w", err)
	}
	defer rows.Close()

	var out []Server
	for rows.Next() {
		var s Server
		var lastSeenUnix int64
		if err := rows.Scan(&s.ID, &s.Name, &s.Host, &s.Port, &s.MaxPlayers, &s.CurrentPlayers, &s.Region, &s.Status, &lastSeenUnix); err != nil {
			return nil, fmt.Errorf("scan: %w", err)
		}
		s.LastSeen = time.Unix(lastSeenUnix, 0).UTC()
		out = append(out, s)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows err: %w", err)
	}
	return out, nil
}

// DeleteServer removes a server by id.
func DeleteServerDB(db *sqlx.DB, id int) error {
	do := func() error {
		_, err := db.Exec(`DELETE FROM servers WHERE id = ?`, id)
		if err != nil {
			return fmt.Errorf("delete server: %w", err)
		}
		return nil
	}
	return execWithRetry(do)
}

// GetServerByID returns a single server by id.
func GetServerByID(db *sqlx.DB, id int) (*Server, error) {
	// Simple query without transaction. Retry on transient failures.
	var out *Server
	do := func() error {
		row := db.QueryRowx(`SELECT id, name, host, port, max_players, current_players, region, status, last_seen FROM servers WHERE id = ?`, id)
		var s Server
		var lastSeenUnix int64
		if err := row.Scan(&s.ID, &s.Name, &s.Host, &s.Port, &s.MaxPlayers, &s.CurrentPlayers, &s.Region, &s.Status, &lastSeenUnix); err != nil {
			if err == sql.ErrNoRows {
				out = nil
				return nil
			}
			return fmt.Errorf("scan server: %w", err)
		}
		s.LastSeen = time.Unix(lastSeenUnix, 0).UTC()
		out = &s
		return nil
	}
	if err := execWithRetry(do); err != nil {
		return nil, err
	}
	return out, nil
}
