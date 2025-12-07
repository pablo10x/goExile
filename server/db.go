package main

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"

	// Use the pure-Go SQLite driver to avoid CGO dependency in builds.
	_ "modernc.org/sqlite"
)

// db.go provides a minimal persistence layer for the registry.

// InitDB opens (or creates) an SQLite database at the given path.
func InitDB(path string) (*sqlx.DB, error) {
	db, err := sqlx.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	db.SetMaxOpenConns(1)
	db.SetConnMaxLifetime(time.Minute * 5)

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

func createTables(db *sqlx.DB) error {
	q := `CREATE TABLE IF NOT EXISTS spawners (
		id INTEGER PRIMARY KEY,
		region TEXT,
		host TEXT NOT NULL,
		port INTEGER NOT NULL,
		max_instances INTEGER NOT NULL,
		current_instances INTEGER NOT NULL,
		status TEXT,
		last_seen INTEGER NOT NULL
	);`
	_, err := db.Exec(q)
	if err != nil {
		return fmt.Errorf("create tables: %w", err)
	}
	return nil
}

func CloseDB(db *sqlx.DB) error {
	if db == nil {
		return nil
	}
	if _, err := db.Exec("PRAGMA wal_checkpoint(TRUNCATE)"); err != nil {
		// ignore
	}
	return db.Close()
}

func execWithRetry(fn func() error) error {
	var lastErr error
	backoff := 50 * time.Millisecond
	for i := 0; i < 5; i++ {
		if err := fn(); err != nil {
			lastErr = err
			time.Sleep(backoff)
			backoff *= 2
			continue
		}
		return nil
	}
	return lastErr
}

// SaveSpawner inserts or replaces a spawner record.
func SaveSpawner(db *sqlx.DB, s *Spawner) (int, error) {
	if s == nil {
		return 0, fmt.Errorf("nil spawner")
	}
	var assignedID int
	do := func() error {
		tx, err := db.Begin()
		if err != nil {
			return fmt.Errorf("begin tx: %w", err)
		}
		defer tx.Rollback()

		if s.ID == 0 {
			res, err := tx.Exec(`INSERT INTO spawners (region, host, port, max_instances, current_instances, status, last_seen) VALUES (?, ?, ?, ?, ?, ?, ?)`,
				s.Region, s.Host, s.Port, s.MaxInstances, s.CurrentInstances, s.Status, s.LastSeen.Unix())
			if err != nil {
				return fmt.Errorf("insert spawner: %w", err)
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

		_, err = tx.Exec(`REPLACE INTO spawners (id, region, host, port, max_instances, current_instances, status, last_seen) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
			s.ID, s.Region, s.Host, s.Port, s.MaxInstances, s.CurrentInstances, s.Status, s.LastSeen.Unix())
		if err != nil {
			return fmt.Errorf("upsert spawner: %w", err)
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

// LoadSpawners returns all spawners from DB.
func LoadSpawners(db *sqlx.DB) ([]Spawner, error) {
	rows, err := db.Queryx(`SELECT id, region, host, port, max_instances, current_instances, status, last_seen FROM spawners`)
	if err != nil {
		return nil, fmt.Errorf("query spawners: %w", err)
	}
	defer rows.Close()

	var out []Spawner
	for rows.Next() {
		var s Spawner
		var lastSeenUnix int64
		if err := rows.Scan(&s.ID, &s.Region, &s.Host, &s.Port, &s.MaxInstances, &s.CurrentInstances, &s.Status, &lastSeenUnix); err != nil {
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

func DeleteSpawnerDB(db *sqlx.DB, id int) error {
	do := func() error {
		_, err := db.Exec(`DELETE FROM spawners WHERE id = ?`, id)
		if err != nil {
			return fmt.Errorf("delete spawner: %w", err)
		}
		return nil
	}
	return execWithRetry(do)
}

// GetSpawnerByID returns a single spawner by id.
func GetSpawnerByID(db *sqlx.DB, id int) (*Spawner, error) {
	var out *Spawner
	do := func() error {
		row := db.QueryRowx(`SELECT id, region, host, port, max_instances, current_instances, status, last_seen FROM spawners WHERE id = ?`, id)
		var s Spawner
		var lastSeenUnix int64
		if err := row.Scan(&s.ID, &s.Region, &s.Host, &s.Port, &s.MaxInstances, &s.CurrentInstances, &s.Status, &lastSeenUnix); err != nil {
			if err == sql.ErrNoRows {
				return nil // Return nil if not found, handled by caller
			}
			return fmt.Errorf("scan spawner: %w", err)
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