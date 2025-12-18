package main

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"

	// Drivers
	_ "github.com/jackc/pgx/v5/stdlib"
)

// db.go provides a persistence layer for the registry using PostgreSQL.

// InitDB opens a database connection to PostgreSQL.
func InitDB(dsn string) (*sqlx.DB, error) {
	// Enforce pgx driver
	db, err := sqlx.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	// Connection pool settings
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(time.Minute * 5)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping db: %w", err)
	}

	if err := createTables(db); err != nil {
		db.Close()
		return nil, err
	}

	// ... rest of function ...

	// Migrations
	// Check if 'version' column exists in server_versions
	var hasVersionCol int
	err = db.QueryRow(`SELECT COUNT(*) FROM information_schema.columns WHERE table_name='server_versions' AND column_name='version'`).Scan(&hasVersionCol)
	if err == nil && hasVersionCol == 0 {
		fmt.Println("Migrating DB: Adding 'version' column to server_versions table...")
		if _, err := db.Exec(`ALTER TABLE server_versions ADD COLUMN version TEXT`); err != nil {
			if !strings.Contains(err.Error(), "duplicate column") && !strings.Contains(err.Error(), "already exists") {
				fmt.Printf("warning: failed to migrate server_versions table: %v\n", err)
			}
		}
	}

	// Ensure unique index exists
	if _, err := db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS idx_spawners_host_port ON spawners(host, port)`); err != nil {
		fmt.Printf("warning: failed to create unique index: %v\n", err)
	}

	// Check if 'game_version' column exists in spawners
	err = db.QueryRow(`SELECT COUNT(*) FROM information_schema.columns WHERE table_name='spawners' AND column_name='game_version'`).Scan(&hasVersionCol)
	if err == nil && hasVersionCol == 0 {
		fmt.Println("Migrating DB: Adding 'game_version' column to spawners table...")
		if _, err := db.Exec(`ALTER TABLE spawners ADD COLUMN game_version TEXT`); err != nil {
			if !strings.Contains(err.Error(), "duplicate column") && !strings.Contains(err.Error(), "already exists") {
				fmt.Printf("warning: failed to migrate spawners table: %v\n", err)
			}
		}
	}

	// Check if 'name' column exists in spawners
	var hasNameCol int
	err = db.QueryRow(`SELECT COUNT(*) FROM information_schema.columns WHERE table_name='spawners' AND column_name='name'`).Scan(&hasNameCol)
	if err == nil && hasNameCol == 0 {
		fmt.Println("Migrating DB: Adding 'name' column to spawners table...")
		if _, err := db.Exec(`ALTER TABLE spawners ADD COLUMN name TEXT`); err != nil {
			if !strings.Contains(err.Error(), "duplicate column") && !strings.Contains(err.Error(), "already exists") {
				fmt.Printf("warning: failed to migrate spawners table (name): %v\n", err)
			}
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

// AdvancedDBStats holds Postgres specific metrics
type AdvancedDBStats struct {
	DatabaseSize  string  `db:"db_size"`
	XactCommit    int64   `db:"xact_commit"`
	XactRollback  int64   `db:"xact_rollback"`
	BlksHit       int64   `db:"blks_hit"`
	BlksRead      int64   `db:"blks_read"`
	TupReturned   int64   `db:"tup_returned"`
	TupFetched    int64   `db:"tup_fetched"`
	TupInserted   int64   `db:"tup_inserted"`
	TupUpdated    int64   `db:"tup_updated"`
	TupDeleted    int64   `db:"tup_deleted"`
	CacheHitRatio float64 // Calculated
}

func GetAdvancedDBStats(db *sqlx.DB) (*AdvancedDBStats, error) {
	if db == nil {
		return nil, fmt.Errorf("db is nil")
	}

	// Get current DB name to filter stats
	var dbName string
	if err := db.Get(&dbName, "SELECT current_database()"); err != nil {
		return nil, err
	}

	var stats AdvancedDBStats
	query := `
        SELECT
            pg_size_pretty(pg_database_size($1)) as db_size,
            xact_commit,
            xact_rollback,
            blks_hit,
            blks_read,
            tup_returned,
            tup_fetched,
            tup_inserted,
            tup_updated,
            tup_deleted
        FROM pg_stat_database
        WHERE datname = $1
    `
	if err := db.Get(&stats, query, dbName); err != nil {
		return nil, err
	}

	if stats.BlksHit+stats.BlksRead > 0 {
		stats.CacheHitRatio = float64(stats.BlksHit) / float64(stats.BlksHit+stats.BlksRead) * 100
	}

	return &stats, nil
}

func createTables(db *sqlx.DB) error {
	pkType := "SERIAL PRIMARY KEY"

	q := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS spawners (
		id %s,
		name TEXT,
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
		id %s,
		filename TEXT NOT NULL,
		version TEXT,
		comment TEXT,
		uploaded_at INTEGER NOT NULL,
		is_active INTEGER DEFAULT 0
	);
	CREATE TABLE IF NOT EXISTS server_config (
		id %s,
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
		id %s,
		spawner_id INTEGER NOT NULL,
		instance_id TEXT NOT NULL,
		action TEXT NOT NULL,
		timestamp INTEGER NOT NULL,
		status TEXT,
		details TEXT,
		FOREIGN KEY(spawner_id) REFERENCES spawners(id) ON DELETE CASCADE
	);`, pkType, pkType, pkType, pkType)

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

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
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
			query := `INSERT INTO spawners (name, region, host, port, max_instances, current_instances, status, last_seen, game_version)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
				ON CONFLICT(host, port) DO UPDATE SET
				name=excluded.name,
				region=excluded.region,
				max_instances=excluded.max_instances,
				current_instances=excluded.current_instances,
				status=excluded.status,
				last_seen=excluded.last_seen,
				game_version=excluded.game_version
				RETURNING id`

			var id int64
			err = tx.QueryRow(query, s.Name, s.Region, s.Host, s.Port, s.MaxInstances, s.CurrentInstances, s.Status, s.LastSeen.Unix(), s.GameVersion).Scan(&id)
			if err != nil {
				return fmt.Errorf("upsert spawner: %w", err)
			}
			assignedID = int(id)
			if err := tx.Commit(); err != nil {
				return fmt.Errorf("commit: %w", err)
			}
			return nil
		}

		query := `INSERT INTO spawners (id, name, region, host, port, max_instances, current_instances, status, last_seen, game_version)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
			ON CONFLICT(id) DO UPDATE SET
			name=excluded.name,
			region=excluded.region,
			host=excluded.host,
			port=excluded.port,
			max_instances=excluded.max_instances,
			current_instances=excluded.current_instances,
			status=excluded.status,
			last_seen=excluded.last_seen,
			game_version=excluded.game_version`

		_, err = tx.Exec(query,
			s.ID, s.Name, s.Region, s.Host, s.Port, s.MaxInstances, s.CurrentInstances, s.Status, s.LastSeen.Unix(), s.GameVersion)
		if err != nil {
			return fmt.Errorf("upsert spawner (id): %w", err)
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
	rows, err := db.Queryx(`SELECT id, name, region, host, port, max_instances, current_instances, status, last_seen, game_version FROM spawners`)
	if err != nil {
		return nil, fmt.Errorf("query spawners: %w", err)
	}
	defer rows.Close()

	out := make([]Spawner, 0)
	for rows.Next() {
		var s Spawner

		var lastSeenUnix int64
		var gameVersion sql.NullString
		var name sql.NullString
		if err := rows.Scan(&s.ID, &name, &s.Region, &s.Host, &s.Port, &s.MaxInstances, &s.CurrentInstances, &s.Status, &lastSeenUnix, &gameVersion); err != nil {
			return nil, fmt.Errorf("scan: %w", err)
		}
		s.LastSeen = time.Unix(lastSeenUnix, 0).UTC()
		if gameVersion.Valid {
			s.GameVersion = gameVersion.String
		}
		if name.Valid {
			s.Name = name.String
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
		_, err := db.Exec(`DELETE FROM spawners WHERE id = $1`, id)
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
		_, err := db.Exec(`INSERT INTO instance_actions (spawner_id, instance_id, action, timestamp, status, details) VALUES ($1, $2, $3, $4, $5, $6)`,
			a.SpawnerID, a.InstanceID, a.Action, a.Timestamp.Unix(), a.Status, a.Details)
		if err != nil {
			return fmt.Errorf("save instance action: %w", err)
		}
		return nil
	}
	return execWithRetry(do)
}

func GetInstanceActions(db *sqlx.DB, spawnerID int, instanceID string) ([]InstanceAction, error) {

	rows, err := db.Queryx(`SELECT id, spawner_id, instance_id, action, timestamp, status, details FROM instance_actions WHERE spawner_id = $1 AND instance_id = $2 ORDER BY timestamp DESC LIMIT 50`, spawnerID, instanceID)
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
		row := db.QueryRowx(`SELECT id, name, region, host, port, max_instances, current_instances, status, last_seen, game_version FROM spawners WHERE id = $1`, id)
		var s Spawner
		var lastSeenUnix int64
		var gameVersion sql.NullString
		var name sql.NullString
		if err := row.Scan(&s.ID, &name, &s.Region, &s.Host, &s.Port, &s.MaxInstances, &s.CurrentInstances, &s.Status, &lastSeenUnix, &gameVersion); err != nil {
			if err == sql.ErrNoRows {
				return nil // Return nil if not found, handled by caller
			}
			return fmt.Errorf("scan spawner: %w", err)
		}
		s.LastSeen = time.Unix(lastSeenUnix, 0).UTC()
		if gameVersion.Valid {
			s.GameVersion = gameVersion.String
		}
		if name.Valid {
			s.Name = name.String
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
		_, err := db.Exec(`INSERT INTO server_versions (filename, version, comment, uploaded_at, is_active) VALUES ($1, $2, $3, $4, $5)`,
			v.Filename, v.Version, v.Comment, v.UploadedAt.Unix(), boolToInt(v.IsActive))
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

	out := make([]GameServerVersion, 0)

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
		if _, err := tx.Exec(`UPDATE server_versions SET is_active = 1 WHERE id = $1`, id); err != nil {
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
		if err := db.QueryRow(`SELECT filename FROM server_versions WHERE id = $1`, id).Scan(&filename); err != nil {
			return err
		}
		if _, err := db.Exec(`DELETE FROM server_versions WHERE id = $1`, id); err != nil {
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
		query := `INSERT INTO server_config
			(key, value, type, category, description, is_read_only, requires_restart, updated_at, updated_by)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
			ON CONFLICT(key) DO UPDATE SET
			value=excluded.value,
			updated_at=excluded.updated_at,
			updated_by=excluded.updated_by`

		_, err := db.Exec(query,
			c.Key, c.Value, c.Type, c.Category, c.Description, boolToInt(c.IsReadOnly), boolToInt(c.RequiresRestart), c.UpdatedAt.Unix(), c.UpdatedBy)
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

	rows, err := db.Queryx(`SELECT id, key, value, type, category, description, is_read_only, requires_restart, updated_at, updated_by FROM server_config WHERE category = $1 ORDER BY key`, category)
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
	err := db.QueryRow(`SELECT id, key, value, type, category, description, is_read_only, requires_restart, updated_at, updated_by FROM server_config WHERE key = $1`, key).
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
		_, err := db.Exec(`UPDATE server_config SET value = $1, updated_at = $2, updated_by = $3 WHERE key = $4 AND is_read_only = 0`,
			value, time.Now().Unix(), updatedBy, key)
		if err != nil {
			return fmt.Errorf("update config: %w", err)
		}
		return nil
	}
	return execWithRetry(do)
}

func ListTables(db *sqlx.DB) ([]string, error) {
	query := "SELECT table_name FROM information_schema.tables WHERE table_schema = 'public'"
	var tables []string
	err := db.Select(&tables, query)
	return tables, err
}

func ListSchemas(db *sqlx.DB) ([]string, error) {
	query := "SELECT schema_name FROM information_schema.schemata WHERE schema_name NOT IN ('information_schema', 'pg_catalog', 'pg_toast') AND schema_name NOT LIKE 'pg_temp_%' AND schema_name NOT LIKE 'pg_toast_temp_%'"
	var schemas []string
	err := db.Select(&schemas, query)
	return schemas, err
}

func ListAllTables(db *sqlx.DB) ([]map[string]string, error) {
	query := `
        SELECT
            n.nspname AS schema_name,
            c.relname AS table_name
        FROM
            pg_catalog.pg_class c
        JOIN
            pg_catalog.pg_namespace n ON n.oid = c.relnamespace
        WHERE
            c.relkind = 'r'
            AND n.nspname NOT IN ('pg_catalog', 'information_schema')
            AND n.nspname NOT LIKE 'pg_temp_%'
            AND n.nspname NOT LIKE 'pg_toast_temp_%'
        ORDER BY
            n.nspname,
            c.relname;
    `

	rows, err := db.Queryx(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := []map[string]string{}
	for rows.Next() {
		var schema, table string
		if err := rows.Scan(&schema, &table); err != nil {
			return nil, err
		}
		results = append(results, map[string]string{"schema": schema, "table": table})
	}
	return results, nil
}

func CreateSchema(db *sqlx.DB, name, owner string) error {
	// Basic validation to prevent SQL injection (alphanumeric + underscore)
	for _, r := range name {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_') {
			return fmt.Errorf("invalid schema name")
		}
	}

	query := fmt.Sprintf("CREATE SCHEMA %s", name)
	if owner != "" {
		// Validate owner name as well
		for _, r := range owner {
			if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_') {
				return fmt.Errorf("invalid owner name")
			}
		}
		query += fmt.Sprintf(" AUTHORIZATION %s", owner)
	}

	_, err := db.Exec(query)
	return err
}

func DropSchema(db *sqlx.DB, name string) error {
	for _, r := range name {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_') {
			return fmt.Errorf("invalid schema name")
		}
	}
	// CASCADE to drop tables inside
	_, err := db.Exec(fmt.Sprintf("DROP SCHEMA %s CASCADE", name))
	return err
}

type Role struct {
	Name        string `db:"rolname" json:"name"`
	Superuser   bool   `db:"rolsuper" json:"superuser"`
	Inherit     bool   `db:"rolinherit" json:"inherit"`
	CreateRole  bool   `db:"rolcreaterole" json:"create_role"`
	CreateDB    bool   `db:"rolcreatedb" json:"create_db"`
	CanLogin    bool   `db:"rolcanlogin" json:"can_login"`
	Replication bool   `db:"rolreplication" json:"replication"`
	BypassRLS   bool   `db:"rolbypassrls" json:"bypass_rls"`
}

func ListRoles(db *sqlx.DB) ([]Role, error) {
	var roles []Role
	err := db.Select(&roles, "SELECT rolname, rolsuper, rolinherit, rolcreaterole, rolcreatedb, rolcanlogin, rolreplication, rolbypassrls FROM pg_roles")
	return roles, err
}

func CreateRole(db *sqlx.DB, name, password string, options []string) error {
	// Validate role name using security helper
	if err := ValidateRoleName(name); err != nil {
		return err
	}

	// Validate password meets security requirements
	if err := ValidatePassword(password); err != nil {
		return fmt.Errorf("password validation failed: %w", err)
	}

	// Build options string with strict whitelist validation
	allowed := map[string]bool{
		"NOSUPERUSER": true, "CREATEDB": true, "NOCREATEDB": true,
		"CREATEROLE": true, "NOCREATEROLE": true, "INHERIT": true,
		"NOINHERIT": true, "LOGIN": true, "NOLOGIN": true,
		"NOREPLICATION": true, "NOBYPASSRLS": true,
	}
	// Note: SUPERUSER, REPLICATION, BYPASSRLS are intentionally excluded for security

	safeOpts := []string{}
	for _, opt := range options {
		opt = strings.ToUpper(strings.TrimSpace(opt))
		if allowed[opt] {
			safeOpts = append(safeOpts, opt)
		}
	}

	// Use quoted identifier for role name and parameterized password
	// PostgreSQL doesn't support $1 for passwords in CREATE ROLE, so we use format()
	// which is safer than string concatenation
	optStr := ""
	if len(safeOpts) > 0 {
		optStr = " " + strings.Join(safeOpts, " ")
	}

	// Use PostgreSQL's format() function with %I for identifier and %L for literal
	// This provides server-side escaping which is more robust
	q := fmt.Sprintf(`DO $$
        BEGIN
            EXECUTE format('CREATE ROLE %%I WITH LOGIN PASSWORD %%L%s', $1, $2);
        END $$`, optStr)

	_, err := db.Exec(q, name, password)
	return err
}

func DeleteRole(db *sqlx.DB, name string) error {
	// Validate role name using security helper
	if err := ValidateRoleName(name); err != nil {
		return err
	}

	// Use PostgreSQL's format() function with %I for identifier escaping
	q := `DO $$ BEGIN EXECUTE format('DROP ROLE %I', $1); END $$`
	_, err := db.Exec(q, name)
	return err
}

func ListTablesBySchema(db *sqlx.DB, schema string) ([]string, error) {
	// Use pg_catalog.pg_tables as the authoritative source
	query := "SELECT tablename FROM pg_catalog.pg_tables WHERE schemaname = $1 ORDER BY tablename"

	rows, err := db.Queryx(query, schema)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tables := []string{} // Initialize as empty slice
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		tables = append(tables, name)
	}
	return tables, nil
}

type ColumnInfo struct {
	Name         string `db:"column_name" json:"name"`
	DataType     string `db:"data_type" json:"type"`
	IsNullable   string `db:"is_nullable" json:"nullable"`
	IsPrimaryKey bool   `json:"is_pk"`
}

func ListColumns(db *sqlx.DB, schema, table string) ([]ColumnInfo, error) {
	// First get basic column info
	query := "SELECT column_name, data_type, is_nullable FROM information_schema.columns WHERE table_schema = $1 AND table_name = $2 ORDER BY ordinal_position"
	cols := []ColumnInfo{}
	err := db.Select(&cols, query, schema, table)
	if err != nil {
		return nil, err
	}

	// Then get primary key columns
	pkQuery := `
		SELECT kcu.column_name
		FROM information_schema.table_constraints tc
		JOIN information_schema.key_column_usage kcu
			ON tc.constraint_name = kcu.constraint_name
			AND tc.table_schema = kcu.table_schema
		WHERE tc.constraint_type = 'PRIMARY KEY'
			AND tc.table_schema = $1
			AND tc.table_name = $2
	`
	pkCols := []string{}
	rows, err := db.Query(pkQuery, schema, table)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var colName string
			if err := rows.Scan(&colName); err == nil {
				pkCols = append(pkCols, colName)
			}
		}
	}

	// Mark primary key columns
	pkSet := make(map[string]bool)
	for _, pk := range pkCols {
		pkSet[pk] = true
	}
	for i := range cols {
		cols[i].IsPrimaryKey = pkSet[cols[i].Name]
	}

	return cols, nil
}

func ExecuteSQL(db *sqlx.DB, query string) ([]map[string]interface{}, error) {
	rows, err := db.Queryx(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := []map[string]interface{}{}
	for rows.Next() {
		row := make(map[string]interface{})
		if err := rows.MapScan(row); err != nil {
			return nil, err
		}
		// Handle []byte for text columns (common in sqlx/drivers)
		for k, v := range row {
			if b, ok := v.([]byte); ok {
				row[k] = string(b)
			}
		}
		results = append(results, row)
	}
	return results, nil
}

func GetTableCounts(db *sqlx.DB) (map[string]int, error) {
	tables, err := ListTables(db)
	if err != nil {
		return nil, err
	}

	counts := make(map[string]int)
	for _, table := range tables {
		var count int
		// Postgres: SELECT count(*) FROM "table"
		if err := db.Get(&count, fmt.Sprintf("SELECT count(*) FROM %q", table)); err == nil {
			counts[table] = count
		}
	}
	return counts, nil
}

// DDL Helpers

func CreateTable(db *sqlx.DB, schema, name string, columns []string) error {
	// Use security helpers for validation
	if err := ValidateSchemaName(schema); err != nil {
		return err
	}
	if err := ValidateTableName(name); err != nil {
		return err
	}
	if len(columns) == 0 {
		return fmt.Errorf("columns definition required")
	}

	// Validate each column definition
	for _, col := range columns {
		if err := ValidateColumnDefinition(col); err != nil {
			return fmt.Errorf("invalid column definition: %w", err)
		}
	}

	q := fmt.Sprintf("CREATE TABLE %q.%q (%s)", schema, name, strings.Join(columns, ", "))
	_, err := db.Exec(q)
	return err
}

func DropTable(db *sqlx.DB, schema, name string) error {
	// Use security helpers for validation
	if err := ValidateSchemaName(schema); err != nil {
		return err
	}
	if err := ValidateTableName(name); err != nil {
		return err
	}
	q := fmt.Sprintf("DROP TABLE %q.%q", schema, name)
	_, err := db.Exec(q)
	return err
}

func AddColumn(db *sqlx.DB, schema, table, colName, colType string) error {
	// Use security helpers for validation
	if err := ValidateSchemaName(schema); err != nil {
		return err
	}
	if err := ValidateTableName(table); err != nil {
		return err
	}
	if err := ValidateColumnName(colName); err != nil {
		return err
	}
	// Validate column type against whitelist
	if err := ValidateSQLType(colType); err != nil {
		return fmt.Errorf("invalid column type: %w", err)
	}

	q := fmt.Sprintf("ALTER TABLE %q.%q ADD COLUMN %q %s", schema, table, colName, colType)
	_, err := db.Exec(q)
	return err
}

func DropColumn(db *sqlx.DB, schema, table, colName string) error {
	// Use security helpers for validation
	if err := ValidateSchemaName(schema); err != nil {
		return err
	}
	if err := ValidateTableName(table); err != nil {
		return err
	}
	if err := ValidateColumnName(colName); err != nil {
		return err
	}
	q := fmt.Sprintf("ALTER TABLE %q.%q DROP COLUMN %q", schema, table, colName)
	_, err := db.Exec(q)
	return err
}

func RenameColumn(db *sqlx.DB, schema, table, oldName, newName string) error {
	// Use security helpers for validation
	if err := ValidateSchemaName(schema); err != nil {
		return err
	}
	if err := ValidateTableName(table); err != nil {
		return err
	}
	if err := ValidateColumnName(oldName); err != nil {
		return err
	}
	if err := ValidateColumnName(newName); err != nil {
		return err
	}
	q := fmt.Sprintf("ALTER TABLE %q.%q RENAME COLUMN %q TO %q", schema, table, oldName, newName)
	_, err := db.Exec(q)
	return err
}

func AlterColumnType(db *sqlx.DB, schema, table, column, newType string) error {
	// Use security helpers for validation
	if err := ValidateSchemaName(schema); err != nil {
		return err
	}
	if err := ValidateTableName(table); err != nil {
		return err
	}
	if err := ValidateColumnName(column); err != nil {
		return err
	}
	// Validate new type against whitelist
	if err := ValidateSQLType(newType); err != nil {
		return fmt.Errorf("invalid column type: %w", err)
	}

	// Using CAST for safety when converting types
	q := fmt.Sprintf("ALTER TABLE %q.%q ALTER COLUMN %q TYPE %s USING %q::%s", schema, table, column, newType, column, newType)
	_, err := db.Exec(q)
	return err
}

func AlterColumnNullable(db *sqlx.DB, schema, table, column string, notNull bool) error {
	// Use security helpers for validation
	if err := ValidateSchemaName(schema); err != nil {
		return err
	}
	if err := ValidateTableName(table); err != nil {
		return err
	}
	if err := ValidateColumnName(column); err != nil {
		return err
	}
	action := "DROP NOT NULL"
	if notNull {
		action = "SET NOT NULL"
	}
	q := fmt.Sprintf("ALTER TABLE %q.%q ALTER COLUMN %q %s", schema, table, column, action)
	_, err := db.Exec(q)
	return err
}

func AlterColumnDefault(db *sqlx.DB, schema, table, column, defaultVal string) error {
	// Use security helpers for validation
	if err := ValidateSchemaName(schema); err != nil {
		return err
	}
	if err := ValidateTableName(table); err != nil {
		return err
	}
	if err := ValidateColumnName(column); err != nil {
		return err
	}

	var q string
	if defaultVal == "" {
		q = fmt.Sprintf("ALTER TABLE %q.%q ALTER COLUMN %q DROP DEFAULT", schema, table, column)
	} else {
		// Validate default value against allowlist of safe patterns
		if err := ValidateDefaultValue(defaultVal); err != nil {
			return fmt.Errorf("invalid default value: %w", err)
		}
		q = fmt.Sprintf("ALTER TABLE %q.%q ALTER COLUMN %q SET DEFAULT %s", schema, table, column, defaultVal)
	}
	_, err := db.Exec(q)
	return err
}

func isValidName(s string) bool {
	for _, r := range s {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_') {
			return false
		}
	}
	return len(s) > 0
}
