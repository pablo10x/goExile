package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"

	// Drivers
	_ "github.com/jackc/pgx/v5/stdlib"

	"exile/server/models"
	"exile/server/utils"
)

var (
	dbConn       *sqlx.DB
	readOnlyConn *sqlx.DB // Separate connection for read-only queries if configured
)

// db.go provides a persistence layer for the registry using PostgreSQL.

var (
	DBConn         *sqlx.DB
	ReadOnlyDBConn *sqlx.DB
)

func InitDB(dsn string) error {
	var err error
	DBConn, err = sqlx.Connect("pgx", dsn)
	if err != nil {
		DBConn, err = sqlx.Connect("sqlite3", dsn)
		if err != nil {
			return fmt.Errorf("failed to connect to database: %w", err)
		}
	}

	DBConn.SetMaxOpenConns(25)
	DBConn.SetMaxIdleConns(25)
	DBConn.SetConnMaxLifetime(5 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err = DBConn.PingContext(ctx); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	// Schema migration
	if err = MigrateDB(DBConn); err != nil {
		return fmt.Errorf("failed to run database migrations: %w", err)
	}

	// Check if server_config table is empty, if so, populate initial config
	var configCount int
	err = DBConn.QueryRow(`SELECT COUNT(*) FROM server_config`).Scan(&configCount)
	if err == nil && configCount == 0 {
		// No initial config population logic here now, handled by InitConfig
	}
	return nil
}

// InitReadOnlyDB opens a separate read-only connection.
func InitReadOnlyDB(dsn string) error {
	var err error
	ReadOnlyDBConn, err = sqlx.Connect("pgx", dsn)
	if err != nil {
		ReadOnlyDBConn, err = sqlx.Connect("sqlite3", dsn)
		if err != nil {
			return fmt.Errorf("failed to connect to read-only database: %w", err)
		}
	}

	ReadOnlyDBConn.SetMaxOpenConns(25)
	ReadOnlyDBConn.SetMaxIdleConns(25)
	ReadOnlyDBConn.SetConnMaxLifetime(5 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err = ReadOnlyDBConn.PingContext(ctx); err != nil {
		return fmt.Errorf("failed to ping read-only database: %w", err)
	}
	return nil
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

func MigrateDB(db *sqlx.DB) error {
	pkType := "SERIAL PRIMARY KEY"

	queries := []string{
		fmt.Sprintf(`CREATE TABLE IF NOT EXISTS spawners (
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
	)`, pkType),
		fmt.Sprintf(`CREATE TABLE IF NOT EXISTS server_versions (
		id %s,
		filename TEXT NOT NULL,
		version TEXT,
		comment TEXT,
		uploaded_at INTEGER NOT NULL,
		is_active INTEGER DEFAULT 0
	)`, pkType),
		fmt.Sprintf(`CREATE TABLE IF NOT EXISTS server_config (
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
	)`, pkType),
		fmt.Sprintf(`CREATE TABLE IF NOT EXISTS instance_actions (
		id %s,
		spawner_id INTEGER NOT NULL,
		instance_id TEXT NOT NULL,
		action TEXT NOT NULL,
		timestamp INTEGER NOT NULL,
		status TEXT,
		details TEXT,
		FOREIGN KEY(spawner_id) REFERENCES spawners(id) ON DELETE CASCADE
	)`, pkType),
		fmt.Sprintf(`CREATE TABLE IF NOT EXISTS notes (
		id %s,
		title TEXT,
		content TEXT,
		color TEXT,
		status TEXT,
		rotation REAL DEFAULT 0,
		created_at INTEGER NOT NULL,
		updated_at INTEGER NOT NULL
	)`, pkType),
		fmt.Sprintf(`CREATE TABLE IF NOT EXISTS todos (
						id %s,
						parent_id INTEGER,
						content TEXT NOT NULL,
						done INTEGER DEFAULT 0,
						in_progress INTEGER DEFAULT 0,
						created_at INTEGER NOT NULL,
						deadline INTEGER,
						FOREIGN KEY(parent_id) REFERENCES todos(id) ON DELETE CASCADE
					)`, pkType),
		fmt.Sprintf(`CREATE TABLE IF NOT EXISTS todo_comments (
						id %s,
						todo_id INTEGER NOT NULL,
						content TEXT NOT NULL,
						author TEXT,
						created_at INTEGER NOT NULL,
						FOREIGN KEY(todo_id) REFERENCES todos(id) ON DELETE CASCADE
					)`, pkType), fmt.Sprintf(`CREATE TABLE IF NOT EXISTS system_logs (
		id %s,
		timestamp INTEGER NOT NULL,
		level TEXT NOT NULL,
		category TEXT NOT NULL,
		source TEXT,
		message TEXT,
		details TEXT,
		client_ip TEXT,
		path TEXT,
		method TEXT
	)`, pkType),
		fmt.Sprintf(`CREATE TABLE IF NOT EXISTS redeye_rules (
		id %s,
		name TEXT,
		cidr TEXT NOT NULL,
		port TEXT NOT NULL,
		path_pattern TEXT DEFAULT '',
		protocol TEXT NOT NULL,
		action TEXT NOT NULL,
		rate_limit INTEGER DEFAULT 0,
		burst INTEGER DEFAULT 0,
		enabled INTEGER DEFAULT 1,
		created_at INTEGER NOT NULL
	)`, pkType),
		fmt.Sprintf(`CREATE TABLE IF NOT EXISTS redeye_logs (
		id %s,
		rule_id INTEGER,
		source_ip TEXT NOT NULL,
		dest_port INTEGER NOT NULL,
		protocol TEXT NOT NULL,
		action TEXT NOT NULL,
		timestamp INTEGER NOT NULL,
		FOREIGN KEY(rule_id) REFERENCES redeye_rules(id) ON DELETE SET NULL
	)`, pkType),
		fmt.Sprintf(`CREATE TABLE IF NOT EXISTS redeye_anticheat_events (
		id %s,
		player_id TEXT,
		game_server_id INTEGER,
		event_type TEXT NOT NULL,
		details TEXT,
		client_ip TEXT NOT NULL,
		severity INTEGER DEFAULT 0,
		timestamp INTEGER NOT NULL
	)`, pkType),
		`CREATE TABLE IF NOT EXISTS redeye_ip_reputation (
		ip TEXT PRIMARY KEY,
		reputation_score INTEGER DEFAULT 0,
		total_events INTEGER DEFAULT 0,
		last_seen INTEGER NOT NULL,
		is_banned INTEGER DEFAULT 0,
		ban_reason TEXT,
		ban_expires_at INTEGER
	)`,
	}

	for _, q := range queries {
		if _, err := db.Exec(q); err != nil {
			return fmt.Errorf("create tables: %w", err)
		}
	}

	// Migrations for existing tables
	// Add in_progress to todos if missing
	_, _ = db.Exec("ALTER TABLE todos ADD COLUMN in_progress INTEGER DEFAULT 0")
	// Add deadline to todos if missing
	_, _ = db.Exec("ALTER TABLE todos ADD COLUMN deadline INTEGER")
	// Add parent_id to todos if missing
	_, _ = db.Exec("ALTER TABLE todos ADD COLUMN parent_id INTEGER")
	return nil
}

// -- System Logs --

func SaveSystemLog(db *sqlx.DB, logEntry *models.SystemLog) error {
	query := `INSERT INTO system_logs (timestamp, level, category, source, message, details, client_ip, path, method)
              VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
	_, err := db.Exec(query, logEntry.Timestamp.Unix(), logEntry.Level, logEntry.Category, logEntry.Source, logEntry.Message, logEntry.Details, logEntry.ClientIP, logEntry.Path, logEntry.Method)
	return err
}
func GetSystemLogs(db *sqlx.DB, category string, limit, offset int) ([]models.SystemLog, int64, error) {
	var logs []models.SystemLog
	var total int64

	whereClause := "WHERE 1=1"
	args := []interface{}{}

	if category != "" && category != "All" {
		whereClause += " AND category = $1"
		args = append(args, category)
	}

	// Count total
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM system_logs %s", whereClause)
	if err := db.Get(&total, countQuery, args...); err != nil {
		return nil, 0, fmt.Errorf("count logs: %w", err)
	}

	// Fetch logs
	query := fmt.Sprintf(`SELECT id, timestamp, level, category, source, message, details, client_ip, path, method 
                          FROM system_logs %s ORDER BY timestamp DESC LIMIT $%d OFFSET $%d`,
		whereClause, len(args)+1, len(args)+2)

	args = append(args, limit, offset)

	rows, err := db.Queryx(query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("query logs: %w", err)
	}
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		var l models.SystemLog
		var ts int64
		if err := rows.Scan(&l.ID, &ts, &l.Level, &l.Category, &l.Source, &l.Message, &l.Details, &l.ClientIP, &l.Path, &l.Method); err != nil {
			return nil, 0, err
		}
		l.Timestamp = time.Unix(ts, 0).UTC()
		logs = append(logs, l)
	}

	// Return empty slice instead of nil for JSON
	if logs == nil {
		logs = []models.SystemLog{}
	}

	return logs, total, nil
}

func GetSystemLogCounts(db *sqlx.DB) (map[string]int, error) {
	query := `SELECT category, COUNT(*) as count FROM system_logs GROUP BY category`
	rows, err := db.Queryx(query)
	if err != nil {
		return nil, fmt.Errorf("query counts: %w", err)
	}
	defer func() { _ = rows.Close() }()

	counts := make(map[string]int)
	var total int
	for rows.Next() {
		var category string
		var count int
		if err := rows.Scan(&category, &count); err != nil {
			return nil, err
		}
		counts[category] = count
		total += count
	}
	counts["All"] = total
	return counts, nil
}

func DeleteSystemLog(db *sqlx.DB, id int) error {
	_, err := db.Exec(`DELETE FROM system_logs WHERE id = $1`, id)
	return err
}

func ClearSystemLogs(db *sqlx.DB) error {
	_, err := db.Exec(`DELETE FROM system_logs`)
	return err
}

// ... existing functions ...

// -- Notes --

// GetNotes retrieves all sticky notes from the database.
func GetNotes(db *sqlx.DB) ([]models.Note, error) {
	// Use COALESCE to handle potential NULLs from migration or old data
	query := `
		SELECT 
			id, 
			COALESCE(title, '') as title, 
			COALESCE(content, '') as content, 
			COALESCE(color, 'yellow') as color, 
			COALESCE(status, 'normal') as status, 
			COALESCE(rotation, 0) as rotation, 
			COALESCE(created_at, CAST(extract(epoch from now()) AS INTEGER)) as created_at, 
			COALESCE(updated_at, CAST(extract(epoch from now()) AS INTEGER)) as updated_at 
		FROM notes 
		ORDER BY created_at DESC
	`
	rows, err := db.Queryx(query)
	if err != nil {
		return nil, fmt.Errorf("query notes: %w", err)
	}
	defer func() { _ = rows.Close() }()

	out := make([]models.Note, 0)
	for rows.Next() {
		var n models.Note
		var tsUnixCreated, tsUnixUpdated int64
		if err := rows.Scan(&n.ID, &n.Title, &n.Content, &n.Color, &n.Status, &n.Rotation, &tsUnixCreated, &tsUnixUpdated); err != nil {
			return nil, fmt.Errorf("scan note: %w", err)
		}
		n.CreatedAt = time.Unix(tsUnixCreated, 0).UTC()
		n.UpdatedAt = time.Unix(tsUnixUpdated, 0).UTC()
		out = append(out, n)
	}
	return out, nil
}

func SaveNote(db *sqlx.DB, n *models.Note) (int, error) {
	var id int
	do := func() error {
		// Set UpdatedAt right before saving
		n.UpdatedAt = time.Now().UTC()
		// Use int/float32 explicitly for Postgres INTEGER/REAL
		createdUnix := int(n.CreatedAt.Unix())
		updatedUnix := int(n.UpdatedAt.Unix())
		rotation32 := float32(n.Rotation)

		// Insert or Update
		if n.ID == 0 {
			query := `INSERT INTO notes (title, content, color, status, rotation, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`
			err := db.QueryRow(query, n.Title, n.Content, n.Color, n.Status, rotation32, createdUnix, updatedUnix).Scan(&id)
			if err != nil {
				return fmt.Errorf("insert query failed: %w", err)
			}
		} else {
			query := `UPDATE notes SET title=$1, content=$2, color=$3, status=$4, rotation=$5, updated_at=$6 WHERE id=$7`
			_, err := db.Exec(query, n.Title, n.Content, n.Color, n.Status, rotation32, updatedUnix, n.ID)
			if err != nil {
				return fmt.Errorf("update query failed: %w", err)
			}
			id = n.ID
		}
		return nil
	}
	if err := execWithRetry(do); err != nil {
		return 0, err
	}
	return id, nil
}

func DeleteNote(db *sqlx.DB, id int) error {
	do := func() error {
		_, err := db.Exec(`DELETE FROM notes WHERE id = $1`, id)
		return err
	}
	return execWithRetry(do)
}

// -- Todos --

// GetTodos retrieves the hierarchical list of tasks and comments.
func GetTodos(db *sqlx.DB) ([]models.Todo, error) {

	// Fetch all todos

	var all []models.Todo

	rows, err := db.Queryx(`SELECT id, parent_id, content, done, in_progress, created_at, deadline FROM todos ORDER BY created_at ASC`)

	if err != nil {

		return nil, fmt.Errorf("query todos: %w", err)

	}

	defer func() { _ = rows.Close() }()

	for rows.Next() {

		var t models.Todo

		var tsUnix int64

		var deadlineUnix, parentID *int64

		var doneInt, inProgressInt int

		if err := rows.Scan(&t.ID, &parentID, &t.Content, &doneInt, &inProgressInt, &tsUnix, &deadlineUnix); err != nil {

			return nil, fmt.Errorf("scan todo: %w", err)

		}

		t.CreatedAt = time.Unix(tsUnix, 0).UTC()

		t.Done = doneInt == 1

		t.InProgress = inProgressInt == 1

		if deadlineUnix != nil {

			d := time.Unix(*deadlineUnix, 0).UTC()

			t.Deadline = &d

		}

		if parentID != nil {

			pid := int(*parentID)

			t.ParentID = &pid

		}

		all = append(all, t)

	}

	// Fetch all comments

	var comments []models.TodoComment

	cRows, err := db.Queryx(`SELECT id, todo_id, content, author, created_at FROM todo_comments ORDER BY created_at ASC`)

	if err != nil {

		log.Printf("Warning: failed to query todo comments: %v", err)

	} else {

		defer func() { _ = cRows.Close() }()

		for cRows.Next() {
			var c models.TodoComment

			var createdAtUnix int64

			if err := cRows.Scan(&c.ID, &c.TodoID, &c.Content, &c.Author, &createdAtUnix); err != nil {

				log.Printf("Warning: failed to scan todo comment: %v", err)

				continue

			}

			c.CreatedAt = time.Unix(createdAtUnix, 0).UTC()

			comments = append(comments, c)

		}

	}
	// Map comments to todos

	commentMap := make(map[int][]models.TodoComment)

	for _, c := range comments {

		commentMap[c.TodoID] = append(commentMap[c.TodoID], c)

	}

	// Build hierarchy

	var rootNodes []models.Todo

	for i := range all {

		all[i].Comments = commentMap[all[i].ID]

	}

	for i := range all {

		if all[i].ParentID == nil {

			rootNodes = append(rootNodes, all[i])

		}

	}

	// Helper to recursively build tree

	var buildTree func(nodes []models.Todo) []models.Todo

	buildTree = func(nodes []models.Todo) []models.Todo {

		for i := range nodes {

			id := nodes[i].ID

			var children []models.Todo

			for _, t := range all {

				if t.ParentID != nil && *t.ParentID == id {

					children = append(children, t)

				}

			}

			if len(children) > 0 {

				nodes[i].SubTasks = buildTree(children)

			}

		}

		return nodes

	}

	if rootNodes == nil {

		return []models.Todo{}, nil

	}

	return buildTree(rootNodes), nil

}

func SaveTodo(db *sqlx.DB, t *models.Todo) (int, error) {

	var id int

	do := func() error {

		doneInt := 0

		if t.Done {

			doneInt = 1

		}

		inProgressInt := 0

		if t.InProgress {

			inProgressInt = 1

		}

		createdUnix := int(t.CreatedAt.Unix())

		if t.CreatedAt.IsZero() {

			createdUnix = int(time.Now().Unix())

		}

		var deadlineUnix *int64

		if t.Deadline != nil {

			d := t.Deadline.Unix()

			deadlineUnix = &d

		}

		if t.ID == 0 {
			query := `INSERT INTO todos (parent_id, content, done, in_progress, created_at, deadline) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`
			err := db.QueryRow(query, t.ParentID, t.Content, doneInt, inProgressInt, createdUnix, deadlineUnix).Scan(&id)
			if err != nil {
				return err
			}
		} else {
			query := `UPDATE todos SET parent_id=$1, content=$2, done=$3, in_progress=$4, deadline=$5 WHERE id=$6`
			_, err := db.Exec(query, t.ParentID, t.Content, doneInt, inProgressInt, deadlineUnix, t.ID)
			if err != nil {
				return err
			}
			id = t.ID
		}
		return nil
	}
	err := execWithRetry(do)
	return id, err
}

func SaveTodoComment(db *sqlx.DB, c *models.TodoComment) (int, error) {
	var id int
	createdUnix := int(time.Now().Unix())
	query := `INSERT INTO todo_comments (todo_id, content, author, created_at) VALUES ($1, $2, $3, $4) RETURNING id`
	err := db.QueryRow(query, c.TodoID, c.Content, c.Author, createdUnix).Scan(&id)
	return id, err
}

func DeleteTodoComment(db *sqlx.DB, id int) error {
	_, err := db.Exec(`DELETE FROM todo_comments WHERE id = $1`, id)
	return err
}
func DeleteTodo(db *sqlx.DB, id int) error {
	do := func() error {
		_, err := db.Exec(`DELETE FROM todos WHERE id = $1`, id)
		return err
	}
	return execWithRetry(do)
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
func SaveSpawner(db *sqlx.DB, s *models.Spawner) (int, error) {
	if s == nil {
		return 0, fmt.Errorf("nil spawner")
	}
	var assignedID int
	do := func() error {
		tx, err := db.Begin()
		if err != nil {
			return fmt.Errorf("begin tx: %w", err)
		}
		defer func() { _ = tx.Rollback() }()

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
func LoadSpawners(db *sqlx.DB) ([]models.Spawner, error) {
	rows, err := db.Queryx(`SELECT id, name, region, host, port, max_instances, current_instances, status, last_seen, game_version FROM spawners`)
	if err != nil {
		return nil, fmt.Errorf("query spawners: %w", err)
	}
	defer rows.Close()

	out := make([]models.Spawner, 0)
	for rows.Next() {
		var s models.Spawner

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

func SaveInstanceAction(db *sqlx.DB, a *models.InstanceAction) (int, error) {
	var id int
	do := func() error {
		query := `INSERT INTO instance_actions (spawner_id, instance_id, action, timestamp, status, details) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`
		err := db.QueryRow(query, a.SpawnerID, a.InstanceID, a.Action, a.Timestamp.Unix(), a.Status, a.Details).Scan(&id)
		if err != nil {
			return fmt.Errorf("save instance action: %w", err)
		}
		return nil
	}
	err := execWithRetry(do)
	return id, err
}
func GetInstanceActions(db *sqlx.DB, spawnerID int, instanceID string) ([]models.InstanceAction, error) {

	rows, err := db.Queryx(`SELECT id, spawner_id, instance_id, action, timestamp, status, details FROM instance_actions WHERE spawner_id = $1 AND instance_id = $2 ORDER BY timestamp DESC LIMIT 50`, spawnerID, instanceID)
	if err != nil {
		return nil, fmt.Errorf("query instance actions: %w", err)
	}
	defer rows.Close()

	var out []models.InstanceAction
	for rows.Next() {
		var a models.InstanceAction
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
func GetSpawnerByID(db *sqlx.DB, id int) (*models.Spawner, error) {
	var out *models.Spawner
	do := func() error {
		row := db.QueryRowx(`SELECT id, name, region, host, port, max_instances, current_instances, status, last_seen, game_version FROM spawners WHERE id = $1`, id)
		var s models.Spawner
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

func GetSpawnerByHostPort(db *sqlx.DB, host string, port int) (*models.Spawner, error) {
	var out *models.Spawner
	do := func() error {
		row := db.QueryRowx(`SELECT id, name, region, host, port, max_instances, current_instances, status, last_seen, game_version FROM spawners WHERE host = $1 AND port = $2`, host, port)
		var s models.Spawner
		var lastSeenUnix int64
		var gameVersion sql.NullString
		var name sql.NullString
		if err := row.Scan(&s.ID, &name, &s.Region, &s.Host, &s.Port, &s.MaxInstances, &s.CurrentInstances, &s.Status, &lastSeenUnix, &gameVersion); err != nil {
			if err == sql.ErrNoRows {
				return nil
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

func SaveServerVersion(db *sqlx.DB, v *models.GameServerVersion) error {
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

// ListServerVersions returns all uploaded server binary versions.
func ListServerVersions(db *sqlx.DB) ([]models.GameServerVersion, error) {

	rows, err := db.Queryx(`SELECT id, filename, version, comment, uploaded_at, is_active FROM server_versions ORDER BY uploaded_at DESC`)

	if err != nil {

		return nil, fmt.Errorf("query versions: %w", err)

	}

	defer func() { _ = rows.Close() }()

	out := make([]models.GameServerVersion, 0)
	for rows.Next() {

		var v models.GameServerVersion

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
		defer func() { _ = tx.Rollback() }()

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

func GetActiveServerVersion(db *sqlx.DB) (*models.GameServerVersion, error) {
	var v models.GameServerVersion
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
	defaultConfigs := []models.ServerConfig{
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
		{
			Key:             "allowed_origins",
			Value:           "http://localhost:5173,http://localhost:8081,http://127.0.0.1:5173,http://127.0.0.1:8081",
			Type:            "string",
			Category:        "system",
			Description:     "Comma-separated list of allowed WebSocket origins",
			IsReadOnly:      false,
			RequiresRestart: false,
			UpdatedBy:       "system",
		},
		{
			Key:             "health_check_interval",
			Value:           "10s",
			Type:            "duration",
			Category:        "system",
			Description:     "Frequency of health checks",
			IsReadOnly:      false,
			RequiresRestart: true,
			UpdatedBy:       "system",
		},
		{
			Key:             "maintenance_mode",
			Value:           "false",
			Type:            "bool",
			Category:        "system",
			Description:     "Enable maintenance mode (reject new connections)",
			IsReadOnly:      false,
			RequiresRestart: false,
			UpdatedBy:       "system",
		},
		{
			Key:             "redeye.auto_ban_enabled",
			Value:           "true",
			Type:            "bool",
			Category:        "redeye",
			Description:     "Automatically ban IPs that exceed reputation threshold",
			IsReadOnly:      false,
			RequiresRestart: false,
			UpdatedBy:       "system",
		},
		{
			Key:             "redeye.auto_ban_threshold",
			Value:           "100",
			Type:            "int",
			Category:        "redeye",
			Description:     "Reputation score threshold for auto-ban",
			IsReadOnly:      false,
			RequiresRestart: false,
			UpdatedBy:       "system",
		},
		{
			Key:             "redeye.alert_enabled",
			Value:           "true",
			Type:            "bool",
			Category:        "redeye",
			Description:     "Enable alerts for high severity events",
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

func SaveConfig(db *sqlx.DB, c *models.ServerConfig) error {
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

func GetAllConfig(db *sqlx.DB) ([]models.ServerConfig, error) {

	rows, err := db.Queryx(`SELECT id, key, value, type, category, description, is_read_only, requires_restart, updated_at, updated_by FROM server_config ORDER BY category, key`)
	if err != nil {
		return nil, fmt.Errorf("query config: %w", err)
	}
	defer rows.Close()

	var out []models.ServerConfig
	for rows.Next() {
		var c models.ServerConfig
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

func GetConfigByCategory(db *sqlx.DB, category string) ([]models.ServerConfig, error) {

	rows, err := db.Queryx(`SELECT id, key, value, type, category, description, is_read_only, requires_restart, updated_at, updated_by FROM server_config WHERE category = $1 ORDER BY key`, category)
	if err != nil {
		return nil, fmt.Errorf("query config by category: %w", err)
	}
	defer rows.Close()

	var out []models.ServerConfig
	for rows.Next() {
		var c models.ServerConfig
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

func GetConfigByKey(db *sqlx.DB, key string) (*models.ServerConfig, error) {
	var c models.ServerConfig
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
	if err := utils.ValidateRoleName(name); err != nil {
		return err
	}

	// Validate password meets security requirements
	if err := utils.ValidatePassword(password); err != nil {
		return fmt.Errorf("password validation failed: %w", err)
	}

	// Build options string with strict whitelist validation
	allowed := map[string]bool{
		"NOSUPERUSER": true, "CREATEDB": true, "NOCREATEDB": true,
		"CREATEROLE": true, "NOCREATEROLE": true, "INHERIT": true,
		"NOINHERIT": true, "LOGIN": true, "NOLOGIN": true,
		"NOREPLICATION": true, "NOBYPASSRLS": true,
	}
	// models.Note: SUPERUSER, REPLICATION, BYPASSRLS are intentionally excluded for security

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
	if err := utils.ValidateRoleName(name); err != nil {
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
	if err := utils.ValidateSchemaName(schema); err != nil {
		return err
	}
	if err := utils.ValidateTableName(name); err != nil {
		return err
	}
	if len(columns) == 0 {
		return fmt.Errorf("columns definition required")
	}

	// Validate each column definition
	for _, col := range columns {
		if err := utils.ValidateColumnDefinition(col); err != nil {
			return fmt.Errorf("invalid column definition: %w", err)
		}
	}

	q := fmt.Sprintf("CREATE TABLE %q.%q (%s)", schema, name, strings.Join(columns, ", "))
	_, err := db.Exec(q)
	return err
}

func DropTable(db *sqlx.DB, schema, name string) error {
	// Use security helpers for validation
	if err := utils.ValidateSchemaName(schema); err != nil {
		return err
	}
	if err := utils.ValidateTableName(name); err != nil {
		return err
	}
	q := fmt.Sprintf("DROP TABLE %q.%q", schema, name)
	_, err := db.Exec(q)
	return err
}

func AddColumn(db *sqlx.DB, schema, table, colName, colType string) error {
	// Use security helpers for validation
	if err := utils.ValidateSchemaName(schema); err != nil {
		return err
	}
	if err := utils.ValidateTableName(table); err != nil {
		return err
	}
	if err := utils.ValidateColumnName(colName); err != nil {
		return err
	}
	// Validate column type against whitelist
	if err := utils.ValidateSQLType(colType); err != nil {
		return fmt.Errorf("invalid column type: %w", err)
	}

	q := fmt.Sprintf("ALTER TABLE %q.%q ADD COLUMN %q %s", schema, table, colName, colType)
	_, err := db.Exec(q)
	return err
}

func DropColumn(db *sqlx.DB, schema, table, colName string) error {
	// Use security helpers for validation
	if err := utils.ValidateSchemaName(schema); err != nil {
		return err
	}
	if err := utils.ValidateTableName(table); err != nil {
		return err
	}
	if err := utils.ValidateColumnName(colName); err != nil {
		return err
	}
	q := fmt.Sprintf("ALTER TABLE %q.%q DROP COLUMN %q", schema, table, colName)
	_, err := db.Exec(q)
	return err
}

func RenameColumn(db *sqlx.DB, schema, table, oldName, newName string) error {
	// Use security helpers for validation
	if err := utils.ValidateSchemaName(schema); err != nil {
		return err
	}
	if err := utils.ValidateTableName(table); err != nil {
		return err
	}
	if err := utils.ValidateColumnName(oldName); err != nil {
		return err
	}
	if err := utils.ValidateColumnName(newName); err != nil {
		return err
	}
	q := fmt.Sprintf("ALTER TABLE %q.%q RENAME COLUMN %q TO %q", schema, table, oldName, newName)
	_, err := db.Exec(q)
	return err
}

func AlterColumnType(db *sqlx.DB, schema, table, column, newType string) error {
	// Use security helpers for validation
	if err := utils.ValidateSchemaName(schema); err != nil {
		return err
	}
	if err := utils.ValidateTableName(table); err != nil {
		return err
	}
	if err := utils.ValidateColumnName(column); err != nil {
		return err
	}
	// Validate new type against whitelist
	if err := utils.ValidateSQLType(newType); err != nil {
		return fmt.Errorf("invalid column type: %w", err)
	}

	// Using CAST for safety when converting types
	q := fmt.Sprintf("ALTER TABLE %q.%q ALTER COLUMN %q TYPE %s USING %q::%s", schema, table, column, newType, column, newType)
	_, err := db.Exec(q)
	return err
}

func AlterColumnNullable(db *sqlx.DB, schema, table, column string, notNull bool) error {
	// Use security helpers for validation
	if err := utils.ValidateSchemaName(schema); err != nil {
		return err
	}
	if err := utils.ValidateTableName(table); err != nil {
		return err
	}
	if err := utils.ValidateColumnName(column); err != nil {
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
	if err := utils.ValidateSchemaName(schema); err != nil {
		return err
	}
	if err := utils.ValidateTableName(table); err != nil {
		return err
	}
	if err := utils.ValidateColumnName(column); err != nil {
		return err
	}

	var q string
	if defaultVal == "" {
		q = fmt.Sprintf("ALTER TABLE %q.%q ALTER COLUMN %q DROP DEFAULT", schema, table, column)
	} else {
		// Validate default value against allowlist of safe patterns
		if err := utils.ValidateDefaultValue(defaultVal); err != nil {
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

// -- RedEye Rules --

func GetRedEyeRuleByID(db *sqlx.DB, id int) (*models.RedEyeRule, error) {
	var r models.RedEyeRule
	var tsUnix int64
	var enabledInt int
	query := `SELECT id, name, cidr, port, COALESCE(path_pattern, '') as path_pattern, protocol, action, rate_limit, burst, enabled, created_at FROM redeye_rules WHERE id = $1`
	err := db.QueryRowx(query, id).Scan(&r.ID, &r.Name, &r.CIDR, &r.Port, &r.PathPattern, &r.Protocol, &r.Action, &r.RateLimit, &r.Burst, &enabledInt, &tsUnix)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Rule not found
		}
		return nil, err
	}
	r.Enabled = enabledInt == 1
	r.CreatedAt = time.Unix(tsUnix, 0).UTC()
	return &r, nil
}

func CreateRedEyeRule(db *sqlx.DB, r *models.RedEyeRule) (int, error) {
	var id int
	do := func() error {
		query := `INSERT INTO redeye_rules (name, cidr, port, path_pattern, protocol, action, rate_limit, burst, enabled, created_at) 
                  VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) RETURNING id`
		err := db.QueryRow(query, r.Name, r.CIDR, r.Port, r.PathPattern, r.Protocol, r.Action, r.RateLimit, r.Burst, boolToInt(r.Enabled), time.Now().Unix()).Scan(&id)
		return err
	}
	if err := execWithRetry(do); err != nil {
		return 0, err
	}

	// Apply OS-level block if it's a DENY rule
	if r.Action == "DENY" {
		if err := utils.BlockIPSystem(r.CIDR); err != nil {
			log.Printf("RedEye: Failed to apply OS block for new rule (CIDR: %s): %v", r.CIDR, err)
		}
	}
	return id, nil
}

func GetRedEyeRules(db *sqlx.DB) ([]models.RedEyeRule, error) {
	rows, err := db.Queryx(`SELECT id, name, cidr, port, COALESCE(path_pattern, '') as path_pattern, protocol, action, rate_limit, burst, enabled, created_at FROM redeye_rules ORDER BY created_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("query redeye rules: %w", err)
	}
	defer rows.Close()

	out := make([]models.RedEyeRule, 0)
	for rows.Next() {
		var r models.RedEyeRule
		var tsUnix int64
		var enabledInt int
		if err := rows.Scan(&r.ID, &r.Name, &r.CIDR, &r.Port, &r.PathPattern, &r.Protocol, &r.Action, &r.RateLimit, &r.Burst, &enabledInt, &tsUnix); err != nil {
			return nil, err
		}
		r.Enabled = enabledInt == 1
		r.CreatedAt = time.Unix(tsUnix, 0).UTC()
		out = append(out, r)
	}
	return out, nil
}

func UpdateRedEyeRule(db *sqlx.DB, r *models.RedEyeRule) error {
	var oldRule *models.RedEyeRule
	var err error

	// Retrieve the old rule before updating
	oldRule, err = GetRedEyeRuleByID(db, r.ID)
	if err != nil {
		return fmt.Errorf("failed to get old rule for update: %w", err)
	}
	if oldRule == nil {
		return fmt.Errorf("rule with ID %d not found for update", r.ID)
	}

	do := func() error {
		query := `UPDATE redeye_rules SET name=$1, cidr=$2, port=$3, path_pattern=$4, protocol=$5, action=$6, rate_limit=$7, burst=$8, enabled=$9 WHERE id=$10`
		_, err := db.Exec(query, r.Name, r.CIDR, r.Port, r.PathPattern, r.Protocol, r.Action, r.RateLimit, r.Burst, boolToInt(r.Enabled), r.ID)
		return err
	}

	if err := execWithRetry(do); err != nil {
		return err
	}

	// Logic for UFW updates based on rule changes
	oldWasDeny := oldRule.Action == "DENY" && oldRule.Enabled
	newIsDeny := r.Action == "DENY" && r.Enabled

	// Case 1: Rule changed from DENY to non-DENY, or CIDR changed
	if oldWasDeny && (!newIsDeny || oldRule.CIDR != r.CIDR) {
		if err := utils.UnblockIPSystem(oldRule.CIDR); err != nil {
			log.Printf("RedEye: Failed to remove old OS block for rule ID %d (CIDR: %s): %v", r.ID, oldRule.CIDR, err)
		}
	}

	// Case 2: Rule changed to DENY, or CIDR changed for an existing DENY rule
	if newIsDeny && (!oldWasDeny || oldRule.CIDR != r.CIDR) {
		if err := utils.BlockIPSystem(r.CIDR); err != nil {
			log.Printf("RedEye: Failed to apply new OS block for rule ID %d (CIDR: %s): %v", r.ID, r.CIDR, err)
		}
	}

	return nil
}

func DeleteRedEyeRule(db *sqlx.DB, id int) error {
	var ruleToDelete *models.RedEyeRule
	var err error

	// Retrieve the rule before deleting
	ruleToDelete, err = GetRedEyeRuleByID(db, id)
	if err != nil {
		return fmt.Errorf("failed to get rule for deletion: %w", err)
	}
	// If ruleToDelete is nil, it means the rule wasn't found, which is fine for a delete operation,
	// but we can't unblock what doesn't exist or wasn't a DENY rule.

	do := func() error {
		_, err := db.Exec(`DELETE FROM redeye_rules WHERE id = $1`, id)
		return err
	}

	if err := execWithRetry(do); err != nil {
		return err
	}

	// If the deleted rule was a DENY rule, unblock the IP at the OS level
	if ruleToDelete != nil && ruleToDelete.Action == "DENY" {
		if err := utils.UnblockIPSystem(ruleToDelete.CIDR); err != nil {
			log.Printf("RedEye: Failed to remove OS block for deleted rule (CIDR: %s): %v", ruleToDelete.CIDR, err)
		}
	}
	return nil
}

// -- RedEye Logs --

func SaveRedEyeLog(db *sqlx.DB, l *models.RedEyeLog) error {
	query := `INSERT INTO redeye_logs (rule_id, source_ip, dest_port, protocol, action, timestamp) 
              VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := db.Exec(query, l.RuleID, l.SourceIP, l.DestPort, l.Protocol, l.Action, l.Timestamp.Unix())
	return err
}

func GetRedEyeLogs(db *sqlx.DB, limit, offset int) ([]models.RedEyeLog, int64, error) {
	var logs []models.RedEyeLog
	var total int64

	if err := db.Get(&total, "SELECT COUNT(*) FROM redeye_logs"); err != nil {
		return nil, 0, err
	}

	query := fmt.Sprintf(`SELECT id, rule_id, source_ip, dest_port, protocol, action, timestamp 
                          FROM redeye_logs ORDER BY timestamp DESC LIMIT $%d OFFSET $%d`, 1, 2)

	rows, err := db.Queryx(query, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	for rows.Next() {
		var l models.RedEyeLog
		var tsUnix int64
		if err := rows.Scan(&l.ID, &l.RuleID, &l.SourceIP, &l.DestPort, &l.Protocol, &l.Action, &tsUnix); err != nil {
			return nil, 0, err
		}
		l.Timestamp = time.Unix(tsUnix, 0).UTC()
		logs = append(logs, l)
	}

	if logs == nil {
		logs = []models.RedEyeLog{}
	}

	return logs, total, nil
}

func ClearRedEyeLogs(db *sqlx.DB) error {
	_, err := db.Exec(`DELETE FROM redeye_logs`)
	return err
}

// -- RedEye Anti-Cheat & Reputation --

func SaveAnticheatEvent(db *sqlx.DB, e *models.RedEyeAnticheatEvent) error {
	do := func() error {
		query := `INSERT INTO redeye_anticheat_events (player_id, game_server_id, event_type, details, client_ip, severity, timestamp)
                  VALUES ($1, $2, $3, $4, $5, $6, $7)`
		_, err := db.Exec(query, e.PlayerID, e.GameServerID, e.EventType, e.Details, e.ClientIP, e.Severity, e.Timestamp.Unix())
		return err
	}
	return execWithRetry(do)
}

func GetAnticheatEvents(db *sqlx.DB, limit, offset int) ([]models.RedEyeAnticheatEvent, int64, error) {
	var events []models.RedEyeAnticheatEvent
	var total int64

	if err := db.Get(&total, "SELECT COUNT(*) FROM redeye_anticheat_events"); err != nil {
		return nil, 0, err
	}

	query := fmt.Sprintf(`SELECT id, player_id, game_server_id, event_type, details, client_ip, severity, timestamp
                          FROM redeye_anticheat_events ORDER BY timestamp DESC LIMIT $%d OFFSET $%d`, 1, 2)
	rows, err := db.Queryx(query, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	for rows.Next() {
		var e models.RedEyeAnticheatEvent
		var tsUnix int64
		if err := rows.Scan(&e.ID, &e.PlayerID, &e.GameServerID, &e.EventType, &e.Details, &e.ClientIP, &e.Severity, &tsUnix); err != nil {
			return nil, 0, err
		}
		e.Timestamp = time.Unix(tsUnix, 0).UTC()
		events = append(events, e)
	}

	if events == nil {
		events = []models.RedEyeAnticheatEvent{}
	}
	return events, total, nil
}

func GetIPReputation(db *sqlx.DB, ip string) (*models.RedEyeIPReputation, error) {
	var r models.RedEyeIPReputation
	var tsUnix int64
	var banExpiresUnix sql.NullInt64
	var isBannedInt int

	query := `SELECT ip, reputation_score, total_events, last_seen, is_banned, ban_reason, ban_expires_at FROM redeye_ip_reputation WHERE ip = $1`
	err := db.QueryRowx(query, ip).Scan(&r.IP, &r.ReputationScore, &r.TotalEvents, &tsUnix, &isBannedInt, &r.BanReason, &banExpiresUnix)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Not found
		}
		return nil, err
	}
	r.LastSeen = time.Unix(tsUnix, 0).UTC()
	r.IsBanned = isBannedInt == 1
	if banExpiresUnix.Valid {
		t := time.Unix(banExpiresUnix.Int64, 0).UTC()
		r.BanExpiresAt = &t
	}
	return &r, nil
}

func UpdateIPReputation(db *sqlx.DB, r *models.RedEyeIPReputation) error {
	do := func() error {
		var banExpiresUnix *int64
		if r.BanExpiresAt != nil {
			t := r.BanExpiresAt.Unix()
			banExpiresUnix = &t
		}

		query := `INSERT INTO redeye_ip_reputation (ip, reputation_score, total_events, last_seen, is_banned, ban_reason, ban_expires_at)
                  VALUES ($1, $2, $3, $4, $5, $6, $7)
                  ON CONFLICT(ip) DO UPDATE SET
                  reputation_score=excluded.reputation_score,
                  total_events=excluded.total_events,
                  last_seen=excluded.last_seen,
                  is_banned=excluded.is_banned,
                  ban_reason=excluded.ban_reason,
                  ban_expires_at=excluded.ban_expires_at`

		_, err := db.Exec(query, r.IP, r.ReputationScore, r.TotalEvents, r.LastSeen.Unix(), boolToInt(r.IsBanned), r.BanReason, banExpiresUnix)
		return err
	}
	return execWithRetry(do)
}

func GetBannedIPList(db *sqlx.DB) ([]string, error) {
	var ips []string
	// Select IPs where is_banned=1 AND (ban_expires_at IS NULL OR ban_expires_at > NOW)
	query := `SELECT ip FROM redeye_ip_reputation WHERE is_banned = 1 AND (ban_expires_at IS NULL OR ban_expires_at > $1)`
	err := db.Select(&ips, query, time.Now().Unix())
	return ips, err
}

func GetBannedIPsFull(db *sqlx.DB) ([]models.RedEyeIPReputation, error) {
	query := `SELECT ip, reputation_score, total_events, last_seen, is_banned, ban_reason, ban_expires_at 
              FROM redeye_ip_reputation 
              WHERE is_banned = 1 AND (ban_expires_at IS NULL OR ban_expires_at > $1)
              ORDER BY last_seen DESC`

	rows, err := db.Queryx(query, time.Now().Unix())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []models.RedEyeIPReputation
	for rows.Next() {
		var r models.RedEyeIPReputation
		var tsUnix int64
		var banExpiresUnix sql.NullInt64
		var isBannedInt int
		if err := rows.Scan(&r.IP, &r.ReputationScore, &r.TotalEvents, &tsUnix, &isBannedInt, &r.BanReason, &banExpiresUnix); err != nil {
			return nil, err
		}
		r.LastSeen = time.Unix(tsUnix, 0).UTC()
		r.IsBanned = isBannedInt == 1
		if banExpiresUnix.Valid {
			t := time.Unix(banExpiresUnix.Int64, 0).UTC()
			r.BanExpiresAt = &t
		}
		out = append(out, r)
	}
	// Return empty slice instead of nil for JSON consistency
	if out == nil {
		out = []models.RedEyeIPReputation{}
	}
	return out, nil
}

func UnbanIP(db *sqlx.DB, ip string) error {
	_, err := db.Exec(`UPDATE redeye_ip_reputation SET is_banned = 0, ban_expires_at = NULL WHERE ip = $1`, ip)
	if err != nil {
		return err
	}

	// Also remove OS-level block
	if err := utils.UnblockIPSystem(ip); err != nil {
		log.Printf("RedEye: Failed to remove OS block for unbanned IP (CIDR: %s): %v", ip, err)
	}
	return nil
}

// RedEyeStats holds summary statistics for the dashboard
type RedEyeStats struct {
	TotalRules      int     `json:"total_rules"`
	ActiveBans      int     `json:"active_bans"`
	Events24h       int     `json:"events_24h"`
	Logs24h         int     `json:"logs_24h"`
	ReputationCount int     `json:"reputation_count"`
	Entropy         float64 `json:"entropy"`
	ThreatLevel     string  `json:"threat_level"`
	Uptime          string  `json:"uptime"`
}

func GetRedEyeStats(db *sqlx.DB) (*RedEyeStats, error) {
	stats := &RedEyeStats{}

	// Run queries in parallel or sequence (sequence is fine for sqlite/low load)
	// Rules
	if err := db.Get(&stats.TotalRules, "SELECT COUNT(*) FROM redeye_rules"); err != nil {
		return nil, err
	}
	// Bans
	if err := db.Get(&stats.ActiveBans, "SELECT COUNT(*) FROM redeye_ip_reputation WHERE is_banned = 1"); err != nil {
		return nil, err
	}
	// Reputation entries
	if err := db.Get(&stats.ReputationCount, "SELECT COUNT(*) FROM redeye_ip_reputation"); err != nil {
		return nil, err
	}

	// Events 24h
	yesterday := time.Now().Add(-24 * time.Hour).Unix()
	if err := db.Get(&stats.Events24h, "SELECT COUNT(*) FROM redeye_anticheat_events WHERE timestamp > $1", yesterday); err != nil {
		return nil, err
	}
	// Logs 24h
	if err := db.Get(&stats.Logs24h, "SELECT COUNT(*) FROM redeye_logs WHERE timestamp > $1", yesterday); err != nil {
		return nil, err
	}

	// Calculate simulated entropy
	stats.Entropy = 0.0042 // Baseline
	if stats.Events24h > 0 {
		stats.Entropy += float64(stats.Events24h) * 0.0001
	}

	// Determine threat level
	stats.ThreatLevel = "Low"
	if stats.ActiveBans > 10 || stats.Events24h > 50 {
		stats.ThreatLevel = "Medium"
	}
	if stats.ActiveBans > 50 || stats.Events24h > 200 {
		stats.ThreatLevel = "High"
	}

	// Subsystem uptime (hardcoded for now as requested by UI aesthetic)
	stats.Uptime = "99.99%"

	return stats, nil
}
