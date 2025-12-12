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

	// Check if 'version' column exists in server_versions (Migration)
	var hasVersionCol int
	err = db.QueryRow(`SELECT COUNT(*) FROM pragma_table_info('server_versions') WHERE name='version'`).Scan(&hasVersionCol)
	if err == nil && hasVersionCol == 0 {
		fmt.Println("Migrating DB: Adding 'version' column to server_versions table...")
		if _, err := db.Exec(`ALTER TABLE server_versions ADD COLUMN version TEXT`); err != nil {
			fmt.Printf("warning: failed to migrate server_versions table: %v\n", err)
		}
	}

	// Ensure unique index exists (migration for existing DBs)
	if _, err := db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS idx_spawners_host_port ON spawners(host, port)`); err != nil {
		fmt.Printf("warning: failed to create unique index: %v\n", err)
	}

	// Check if 'game_version' column exists in spawners (Migration)
	var hasGameVersionCol int
	err = db.QueryRow(`SELECT COUNT(*) FROM pragma_table_info('spawners') WHERE name='game_version'`).Scan(&hasGameVersionCol)
	if err == nil && hasGameVersionCol == 0 {
		fmt.Println("Migrating DB: Adding 'game_version' column to spawners table...")
		if _, err := db.Exec(`ALTER TABLE spawners ADD COLUMN game_version TEXT`); err != nil {
			fmt.Printf("warning: failed to migrate spawners table: %v\n", err)
		}
	}

	// Seed default configuration if table is empty
	var configCount int
	err = db.QueryRow(`SELECT COUNT(*) FROM server_config`).Scan(&configCount)
	if err == nil && configCount == 0 {
		fmt.Println("Seeding default configuration...")
		if err := SeedDefaultConfig(db); err != nil {
			fmt.Printf("warning: failed to seed default config: %v\n", err)
		}
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
		last_seen INTEGER NOT NULL,
		game_version TEXT,
		UNIQUE(host, port)
	);
	CREATE TABLE IF NOT EXISTS server_versions (
		id INTEGER PRIMARY KEY,
		filename TEXT NOT NULL,
		version TEXT,
		comment TEXT,
		uploaded_at INTEGER NOT NULL,
		is_active INTEGER DEFAULT 0
	);
	CREATE TABLE IF NOT EXISTS server_config (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		key TEXT UNIQUE NOT NULL,
		value TEXT NOT NULL,
		type TEXT NOT NULL DEFAULT 'string',
		category TEXT NOT NULL DEFAULT 'system',
		description TEXT,
		is_read_only INTEGER DEFAULT 0,
		requires_restart INTEGER DEFAULT 0,
		updated_at INTEGER NOT NULL,
		updated_by TEXT
	);
	CREATE TABLE IF NOT EXISTS instance_actions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		spawner_id INTEGER NOT NULL,
		instance_id TEXT NOT NULL,
		action TEXT NOT NULL,
		timestamp INTEGER NOT NULL,
		status TEXT,
		details TEXT,
		FOREIGN KEY(spawner_id) REFERENCES spawners(id) ON DELETE CASCADE
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
			// Use UPSERT to handle re-registration of existing spawner (same host/port)
			// effectively getting the existing ID or creating a new one.
			query := `INSERT INTO spawners (region, host, port, max_instances, current_instances, status, last_seen, game_version) 
				VALUES (?, ?, ?, ?, ?, ?, ?, ?)
				ON CONFLICT(host, port) DO UPDATE SET
				region=excluded.region,
				max_instances=excluded.max_instances,
				current_instances=excluded.current_instances,
				status=excluded.status,
				last_seen=excluded.last_seen,
				game_version=excluded.game_version
				RETURNING id`

			var id int64
			err := tx.QueryRow(query, s.Region, s.Host, s.Port, s.MaxInstances, s.CurrentInstances, s.Status, s.LastSeen.Unix(), s.GameVersion).Scan(&id)
			if err != nil {
				return fmt.Errorf("upsert spawner: %w", err)
			}
			assignedID = int(id)
			if err := tx.Commit(); err != nil {
				return fmt.Errorf("commit: %w", err)
			}
			return nil
		}

		_, err = tx.Exec(`REPLACE INTO spawners (id, region, host, port, max_instances, current_instances, status, last_seen, game_version) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			s.ID, s.Region, s.Host, s.Port, s.MaxInstances, s.CurrentInstances, s.Status, s.LastSeen.Unix(), s.GameVersion)
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
	rows, err := db.Queryx(`SELECT id, region, host, port, max_instances, current_instances, status, last_seen, game_version FROM spawners`)
	if err != nil {
		return nil, fmt.Errorf("query spawners: %w", err)
	}
	defer rows.Close()

	var out []Spawner
	for rows.Next() {
		var s Spawner
		var lastSeenUnix int64
		var gameVersion sql.NullString
		if err := rows.Scan(&s.ID, &s.Region, &s.Host, &s.Port, &s.MaxInstances, &s.CurrentInstances, &s.Status, &lastSeenUnix, &gameVersion); err != nil {
			return nil, fmt.Errorf("scan: %w", err)
		}
		s.LastSeen = time.Unix(lastSeenUnix, 0).UTC()
		if gameVersion.Valid {
			s.GameVersion = gameVersion.String
		}
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

// -- Instance Actions --

func SaveInstanceAction(db *sqlx.DB, a *InstanceAction) error {
	do := func() error {
		_, err := db.Exec(`INSERT INTO instance_actions (spawner_id, instance_id, action, timestamp, status, details) VALUES (?, ?, ?, ?, ?, ?)`,
			a.SpawnerID, a.InstanceID, a.Action, a.Timestamp.Unix(), a.Status, a.Details)
		if err != nil {
			return fmt.Errorf("save instance action: %w", err)
		}
		return nil
	}
	return execWithRetry(do)
}

func GetInstanceActions(db *sqlx.DB, spawnerID int, instanceID string) ([]InstanceAction, error) {
	rows, err := db.Queryx(`SELECT id, spawner_id, instance_id, action, timestamp, status, details FROM instance_actions WHERE spawner_id = ? AND instance_id = ? ORDER BY timestamp DESC LIMIT 50`, spawnerID, instanceID)
	if err != nil {
		return nil, fmt.Errorf("query instance actions: %w", err)
	}
	defer rows.Close()

	var out []InstanceAction
	for rows.Next() {
		var a InstanceAction
		var tsUnix int64
		if err := rows.Scan(&a.ID, &a.SpawnerID, &a.InstanceID, &a.Action, &tsUnix, &a.Status, &a.Details); err != nil {
			return nil, fmt.Errorf("scan instance action: %w", err)
		}
		a.Timestamp = time.Unix(tsUnix, 0).UTC()
		out = append(out, a)
	}
	return out, nil
}

// GetSpawnerByID returns a single spawner by id.
func GetSpawnerByID(db *sqlx.DB, id int) (*Spawner, error) {
	var out *Spawner
	do := func() error {
		row := db.QueryRowx(`SELECT id, region, host, port, max_instances, current_instances, status, last_seen, game_version FROM spawners WHERE id = ?`, id)
		var s Spawner
		var lastSeenUnix int64
		var gameVersion sql.NullString
		if err := row.Scan(&s.ID, &s.Region, &s.Host, &s.Port, &s.MaxInstances, &s.CurrentInstances, &s.Status, &lastSeenUnix, &gameVersion); err != nil {
			if err == sql.ErrNoRows {
				return nil // Return nil if not found, handled by caller
			}
			return fmt.Errorf("scan spawner: %w", err)
		}
		s.LastSeen = time.Unix(lastSeenUnix, 0).UTC()
		if gameVersion.Valid {
			s.GameVersion = gameVersion.String
		}
		out = &s
		return nil
	}
	if err := execWithRetry(do); err != nil {
		return nil, err
	}
	return out, nil
}

// -- Server Versions --

func SaveServerVersion(db *sqlx.DB, v *GameServerVersion) error {
	do := func() error {
		_, err := db.Exec(`INSERT INTO server_versions (filename, version, comment, uploaded_at, is_active) VALUES (?, ?, ?, ?, ?)`,
			v.Filename, v.Version, v.Comment, v.UploadedAt.Unix(), v.IsActive)
		if err != nil {
			return fmt.Errorf("insert version: %w", err)
		}
		return nil
	}
	return execWithRetry(do)
}

func ListServerVersions(db *sqlx.DB) ([]GameServerVersion, error) {
	rows, err := db.Queryx(`SELECT id, filename, version, comment, uploaded_at, is_active FROM server_versions ORDER BY uploaded_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("query versions: %w", err)
	}
	defer rows.Close()

	var out []GameServerVersion
	for rows.Next() {
		var v GameServerVersion
		var uploadedAtUnix int64
		// Handle NULL version if migration hasn't happened or older records
		var version sql.NullString
		if err := rows.Scan(&v.ID, &v.Filename, &version, &v.Comment, &uploadedAtUnix, &v.IsActive); err != nil {
			return nil, fmt.Errorf("scan version: %w", err)
		}
		if version.Valid {
			v.Version = version.String
		}
		v.UploadedAt = time.Unix(uploadedAtUnix, 0).UTC()
		out = append(out, v)
	}
	return out, nil
}

func SetActiveVersion(db *sqlx.DB, id int) error {
	do := func() error {
		tx, err := db.Begin()
		if err != nil {
			return err
		}
		defer tx.Rollback()

		// Deactivate all
		if _, err := tx.Exec(`UPDATE server_versions SET is_active = 0`); err != nil {
			return err
		}
		// Activate target
		if _, err := tx.Exec(`UPDATE server_versions SET is_active = 1 WHERE id = ?`, id); err != nil {
			return err
		}

		return tx.Commit()
	}
	return execWithRetry(do)
}

func DeleteServerVersion(db *sqlx.DB, id int) (string, error) {
	var filename string
	do := func() error {
		// Get filename first to return it (so caller can delete file)
		if err := db.QueryRow(`SELECT filename FROM server_versions WHERE id = ?`, id).Scan(&filename); err != nil {
			return err
		}
		if _, err := db.Exec(`DELETE FROM server_versions WHERE id = ?`, id); err != nil {
			return err
		}
		return nil
	}
	if err := execWithRetry(do); err != nil {
		return "", err
	}
	return filename, nil
}

func GetActiveServerVersion(db *sqlx.DB) (*GameServerVersion, error) {
	var v GameServerVersion
	var uploadedAtUnix int64
	var version sql.NullString
	err := db.QueryRow(`SELECT id, filename, version, comment, uploaded_at, is_active FROM server_versions WHERE is_active = 1 LIMIT 1`).
		Scan(&v.ID, &v.Filename, &version, &v.Comment, &uploadedAtUnix, &v.IsActive)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	if version.Valid {
		v.Version = version.String
	}
	v.UploadedAt = time.Unix(uploadedAtUnix, 0).UTC()
	return &v, nil
}

// -- Server Configuration --

func SeedDefaultConfig(db *sqlx.DB) error {
	defaultConfigs := []ServerConfig{
		{
			Key:             "server_port",
			Value:           "8081",
			Type:            "int",
			Category:        "system",
			Description:     "Port for the master server to listen on",
			IsReadOnly:      false,
			RequiresRestart: true,
			UpdatedBy:       "system",
		},
		{
			Key:             "server_ttl",
			Value:           "60s",
			Type:            "duration",
			Category:        "system",
			Description:     "How long a spawner is considered alive since last heartbeat",
			IsReadOnly:      false,
			RequiresRestart: false,
			UpdatedBy:       "system",
		},
		{
			Key:             "cleanup_interval",
			Value:           "30s",
			Type:            "duration",
			Category:        "system",
			Description:     "Frequency of cleanup routine for inactive spawners",
			IsReadOnly:      false,
			RequiresRestart: false,
			UpdatedBy:       "system",
		},
		{
			Key:             "max_body_size",
			Value:           "1MB",
			Type:            "string",
			Category:        "system",
			Description:     "Maximum size for request bodies",
			IsReadOnly:      false,
			RequiresRestart: true,
			UpdatedBy:       "system",
		},
		{
			Key:             "session_timeout",
			Value:           "24h",
			Type:            "duration",
			Category:        "security",
			Description:     "Session timeout for dashboard authentication",
			IsReadOnly:      false,
			RequiresRestart: false,
			UpdatedBy:       "system",
		},
		{
			Key:             "max_instances_per_spawner",
			Value:           "20",
			Type:            "int",
			Category:        "spawner",
			Description:     "Default maximum instances per spawner",
			IsReadOnly:      false,
			RequiresRestart: false,
			UpdatedBy:       "system",
		},
	}

	for _, config := range defaultConfigs {
		config.UpdatedAt = time.Now().UTC()
		if err := SaveConfig(db, &config); err != nil {
			return fmt.Errorf("seed default config %s: %w", config.Key, err)
		}
	}
	return nil
}

func SaveConfig(db *sqlx.DB, c *ServerConfig) error {
	do := func() error {
		_, err := db.Exec(`INSERT OR REPLACE INTO server_config 
			(key, value, type, category, description, is_read_only, requires_restart, updated_at, updated_by) 
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			c.Key, c.Value, c.Type, c.Category, c.Description, c.IsReadOnly, c.RequiresRestart, c.UpdatedAt.Unix(), c.UpdatedBy)
		if err != nil {
			return fmt.Errorf("save config: %w", err)
		}
		return nil
	}
	return execWithRetry(do)
}

func GetAllConfig(db *sqlx.DB) ([]ServerConfig, error) {
	rows, err := db.Queryx(`SELECT id, key, value, type, category, description, is_read_only, requires_restart, updated_at, updated_by FROM server_config ORDER BY category, key`)
	if err != nil {
		return nil, fmt.Errorf("query config: %w", err)
	}
	defer rows.Close()

	var out []ServerConfig
	for rows.Next() {
		var c ServerConfig
		var updatedAtUnix int64
		var updatedBy sql.NullString
		if err := rows.Scan(&c.ID, &c.Key, &c.Value, &c.Type, &c.Category, &c.Description, &c.IsReadOnly, &c.RequiresRestart, &updatedAtUnix, &updatedBy); err != nil {
			return nil, fmt.Errorf("scan config: %w", err)
		}
		c.UpdatedAt = time.Unix(updatedAtUnix, 0).UTC()
		if updatedBy.Valid {
			c.UpdatedBy = updatedBy.String
		}
		out = append(out, c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows err: %w", err)
	}
	return out, nil
}

func GetConfigByCategory(db *sqlx.DB, category string) ([]ServerConfig, error) {
	rows, err := db.Queryx(`SELECT id, key, value, type, category, description, is_read_only, requires_restart, updated_at, updated_by FROM server_config WHERE category = ? ORDER BY key`, category)
	if err != nil {
		return nil, fmt.Errorf("query config by category: %w", err)
	}
	defer rows.Close()

	var out []ServerConfig
	for rows.Next() {
		var c ServerConfig
		var updatedAtUnix int64
		var updatedBy sql.NullString
		if err := rows.Scan(&c.ID, &c.Key, &c.Value, &c.Type, &c.Category, &c.Description, &c.IsReadOnly, &c.RequiresRestart, &updatedAtUnix, &updatedBy); err != nil {
			return nil, fmt.Errorf("scan config: %w", err)
		}
		c.UpdatedAt = time.Unix(updatedAtUnix, 0).UTC()
		if updatedBy.Valid {
			c.UpdatedBy = updatedBy.String
		}
		out = append(out, c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows err: %w", err)
	}
	return out, nil
}

func GetConfigByKey(db *sqlx.DB, key string) (*ServerConfig, error) {
	var c ServerConfig
	var updatedAtUnix int64
	var updatedBy sql.NullString
	err := db.QueryRow(`SELECT id, key, value, type, category, description, is_read_only, requires_restart, updated_at, updated_by FROM server_config WHERE key = ?`, key).
		Scan(&c.ID, &c.Key, &c.Value, &c.Type, &c.Category, &c.Description, &c.IsReadOnly, &c.RequiresRestart, &updatedAtUnix, &updatedBy)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get config by key: %w", err)
	}
	c.UpdatedAt = time.Unix(updatedAtUnix, 0).UTC()
	if updatedBy.Valid {
		c.UpdatedBy = updatedBy.String
	}
	return &c, nil
}

func UpdateConfig(db *sqlx.DB, key, value, updatedBy string) error {
	do := func() error {
		_, err := db.Exec(`UPDATE server_config SET value = ?, updated_at = ?, updated_by = ? WHERE key = ? AND is_read_only = 0`,
			value, time.Now().Unix(), updatedBy, key)
		if err != nil {
			return fmt.Errorf("update config: %w", err)
		}
		return nil
	}
	return execWithRetry(do)
}
