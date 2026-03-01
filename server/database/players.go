package database

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"exile/server/models"

	"github.com/jmoiron/sqlx"
)

// InitPlayerSystem initializes the player management tables.
func InitPlayerSystem(db *sqlx.DB) error {
	isSQLite := db.DriverName() == "sqlite"
	schemaPrefix := "player_system."
	pkType := "BIGSERIAL PRIMARY KEY"
	tsType := "TIMESTAMP WITH TIME ZONE DEFAULT NOW()"
	boolType := "BOOLEAN DEFAULT FALSE"

	if isSQLite {
		schemaPrefix = "ps_" // SQLite doesn't support schemas well, use prefix
		pkType = "INTEGER PRIMARY KEY AUTOINCREMENT"
		tsType = "DATETIME DEFAULT CURRENT_TIMESTAMP"
		boolType = "INTEGER DEFAULT 0"
	} else {
		// Create schema for Postgres
		if _, err := db.Exec(`CREATE SCHEMA IF NOT EXISTS player_system;`); err != nil {
			return fmt.Errorf("create schema: %w", err)
		}
	}

	// Players Table
	playersTable := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %splayers (
		id %s,
		uid TEXT UNIQUE,
		name TEXT NOT NULL,
		device_id TEXT NOT NULL UNIQUE,
		xp BIGINT DEFAULT 0,
		banned %s,
		last_joined_server TEXT DEFAULT '',
		created_at %s,
		updated_at %s
	);`, schemaPrefix, pkType, boolType, tsType, tsType)

	if _, err := db.Exec(playersTable); err != nil {
		return fmt.Errorf("create players table: %w", err)
	}

	// Sequence adjustment (Postgres only)
	if !isSQLite {
		var maxID int64
		if err := db.Get(&maxID, fmt.Sprintf("SELECT COALESCE(MAX(id), 0) FROM %splayers", schemaPrefix)); err == nil && maxID < 1000000 {
			_, _ = db.Exec(fmt.Sprintf("ALTER SEQUENCE %splayers_id_seq RESTART WITH 1000000", schemaPrefix))
		}
	}

	// Migration: Add columns if they don't exist
	if isSQLite {
		// SQLite migrations are simpler but limited (no IF NOT EXISTS for columns)
		// We try to add and ignore error
		_, _ = db.Exec(fmt.Sprintf("ALTER TABLE %splayers ADD COLUMN uid TEXT", schemaPrefix))
		_, _ = db.Exec(fmt.Sprintf("ALTER TABLE %splayers ADD COLUMN banned INTEGER DEFAULT 0", schemaPrefix))
	} else {
		migration := fmt.Sprintf(`DO $$
		BEGIN
			IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema='player_system' AND table_name='players' AND column_name='uid') THEN
				ALTER TABLE player_system.players ADD COLUMN uid TEXT UNIQUE;
			END IF;
			IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema='player_system' AND table_name='players' AND column_name='banned') THEN
				ALTER TABLE player_system.players ADD COLUMN banned BOOLEAN DEFAULT FALSE;
			END IF;
		END $$;`)
		if _, err := db.Exec(migration); err != nil {
			log.Printf("Warning: Failed to run player migrations: %v", err)
		}
	}

	// Friendships Table
	friendshipsTable := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %sfriendships (
		player1_id BIGINT NOT NULL REFERENCES %splayers(id) ON DELETE CASCADE,
		player2_id BIGINT NOT NULL REFERENCES %splayers(id) ON DELETE CASCADE,
		created_at %s,
		PRIMARY KEY (player1_id, player2_id),
		CONSTRAINT check_order CHECK (player1_id < player2_id)
	);`, schemaPrefix, schemaPrefix, schemaPrefix, tsType)
	if _, err := db.Exec(friendshipsTable); err != nil {
		return fmt.Errorf("create friendships table: %w", err)
	}

	// Friend Requests Table
	requestsTable := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %sfriend_requests (
		id %s,
		sender_id BIGINT NOT NULL REFERENCES %splayers(id) ON DELETE CASCADE,
		receiver_id BIGINT NOT NULL REFERENCES %splayers(id) ON DELETE CASCADE,
		status TEXT DEFAULT 'pending',
		created_at %s,
		UNIQUE(sender_id, receiver_id)
	);`, schemaPrefix, pkType, schemaPrefix, schemaPrefix, tsType)
	if _, err := db.Exec(requestsTable); err != nil {
		return fmt.Errorf("create friend_requests table: %w", err)
	}

	// Reports Table
	reportsTable := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %sreports (
		id %s,
		reporter_id BIGINT NOT NULL REFERENCES %splayers(id) ON DELETE CASCADE,
		reported_user_id BIGINT NOT NULL REFERENCES %splayers(id) ON DELETE CASCADE,
		reason TEXT NOT NULL,
		game_server_instance_id TEXT DEFAULT '',
		timestamp %s
	);`, schemaPrefix, pkType, schemaPrefix, schemaPrefix, tsType)
	if _, err := db.Exec(reportsTable); err != nil {
		return fmt.Errorf("create reports table: %w", err)
	}

	return nil
}

func getPSPrefix(db *sqlx.DB) string {
	if db.DriverName() == "sqlite" {
		return "ps_"
	}
	return "player_system."
}

// -- Reports CRUD --

func CreateReport(db *sqlx.DB, r *models.Report) (int64, error) {
	var id int64
	prefix := getPSPrefix(db)
	query := fmt.Sprintf(`INSERT INTO %sreports (reporter_id, reported_user_id, reason, game_server_instance_id, timestamp)
			  VALUES ($1, $2, $3, $4, $5) RETURNING id`, prefix)

	r.Timestamp = time.Now().UTC()

	if db.DriverName() == "sqlite" {
		res, err := db.Exec(fmt.Sprintf(`INSERT INTO %sreports (reporter_id, reported_user_id, reason, game_server_instance_id, timestamp)
			  VALUES ($1, $2, $3, $4, $5)`, prefix), r.ReporterID, r.ReportedUserID, r.Reason, r.GameServerInstanceID, r.Timestamp.Unix())
		if err != nil {
			return 0, err
		}
		return res.LastInsertId()
	}

	err := db.QueryRow(query, r.ReporterID, r.ReportedUserID, r.Reason, r.GameServerInstanceID, r.Timestamp).Scan(&id)
	return id, err
}

func GetAllReports(db *sqlx.DB) ([]models.Report, error) {
	var reports []models.Report
	prefix := getPSPrefix(db)
	query := fmt.Sprintf(`
		SELECT r.*, p1.name as reporter_name, p2.name as reported_user_name
		FROM %sreports r
		JOIN %splayers p1 ON r.reporter_id = p1.id
		JOIN %splayers p2 ON r.reported_user_id = p2.id
		ORDER BY r.timestamp DESC
	`, prefix, prefix, prefix)
	err := db.Select(&reports, query)
	return reports, err
}

// -- Player CRUD --

func CreatePlayer(db *sqlx.DB, p *models.Player) (int64, error) {
	prefix := getPSPrefix(db)
	p.CreatedAt = time.Now().UTC()
	p.UpdatedAt = time.Now().UTC()

	if db.DriverName() == "sqlite" {
		query := fmt.Sprintf(`INSERT INTO %splayers (uid, name, device_id, xp, banned, last_joined_server, created_at, updated_at)
                          VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`, prefix)
		res, err := db.Exec(query, p.UID, p.Name, p.DeviceID, p.XP, boolToInt(p.Banned), p.LastJoinedServer, p.CreatedAt.Unix(), p.UpdatedAt.Unix())
		if err != nil {
			return 0, err
		}
		return res.LastInsertId()
	}

	var id int64
	query := fmt.Sprintf(`INSERT INTO %splayers (uid, name, device_id, xp, banned, last_joined_server, created_at, updated_at)
                          VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id`, prefix)
	err := db.QueryRow(query, p.UID, p.Name, p.DeviceID, p.XP, p.Banned, p.LastJoinedServer, p.CreatedAt, p.UpdatedAt).Scan(&id)
	return id, err
}

func GetPlayerByUID(db *sqlx.DB, uid string) (*models.Player, error) {
	var p models.Player
	prefix := getPSPrefix(db)
	query := fmt.Sprintf(`SELECT * FROM %splayers WHERE uid = $1`, prefix)
	err := db.Get(&p, query, uid)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &p, nil
}

func GetPlayerByDeviceID(db *sqlx.DB, deviceID string) (*models.Player, error) {
	var p models.Player
	prefix := getPSPrefix(db)
	query := fmt.Sprintf(`SELECT * FROM %splayers WHERE device_id = $1`, prefix)
	err := db.Get(&p, query, deviceID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &p, nil
}

func GetPlayerByID(db *sqlx.DB, id int64) (*models.Player, error) {
	var p models.Player
	prefix := getPSPrefix(db)
	query := fmt.Sprintf(`SELECT * FROM %splayers WHERE id = $1`, prefix)
	err := db.Get(&p, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &p, nil
}

func GetAllPlayers(db *sqlx.DB) ([]models.Player, error) {
	var players []models.Player
	prefix := getPSPrefix(db)
	query := fmt.Sprintf(`SELECT * FROM %splayers ORDER BY id DESC`, prefix)
	err := db.Select(&players, query)
	return players, err
}

func UpdatePlayer(db *sqlx.DB, p *models.Player) error {
	p.UpdatedAt = time.Now().UTC()
	prefix := getPSPrefix(db)
	query := fmt.Sprintf(`UPDATE %splayers SET uid=:uid, name=:name, device_id=:device_id, xp=:xp, banned=:banned, last_joined_server=:last_joined_server, updated_at=:updated_at WHERE id=:id`, prefix)
	_, err := db.NamedExec(query, p)
	return err
}

func DeletePlayer(db *sqlx.DB, id int64) error {
	prefix := getPSPrefix(db)
	_, err := db.Exec(fmt.Sprintf(`DELETE FROM %splayers WHERE id = $1`, prefix), id)
	return err
}

// -- Friends System --

func SendFriendRequest(db *sqlx.DB, senderID, receiverID int64) error {
	if senderID == receiverID {
		return fmt.Errorf("cannot friend yourself")
	}
	prefix := getPSPrefix(db)
	// Check if already friends
	var count int
	p1, p2 := sortIDs(senderID, receiverID)
	err := db.Get(&count, fmt.Sprintf(`SELECT COUNT(*) FROM %sfriendships WHERE player1_id=$1 AND player2_id=$2`, prefix), p1, p2)
	if err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf("already friends")
	}

	query := fmt.Sprintf(`INSERT INTO %sfriend_requests (sender_id, receiver_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`, prefix)
	_, err = db.Exec(query, senderID, receiverID)
	return err
}

func AcceptFriendRequest(db *sqlx.DB, senderID, receiverID int64) error {
	prefix := getPSPrefix(db)
	// 1. Delete request
	res, err := db.Exec(fmt.Sprintf(`DELETE FROM %sfriend_requests WHERE sender_id=$1 AND receiver_id=$2`, prefix), senderID, receiverID)
	if err != nil {
		return err
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("request not found")
	}

	// 2. Create friendship
	p1, p2 := sortIDs(senderID, receiverID)
	_, err = db.Exec(fmt.Sprintf(`INSERT INTO %sfriendships (player1_id, player2_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`, prefix), p1, p2)
	return err
}

func GetFriends(db *sqlx.DB, playerID int64) ([]models.Player, error) {
	var friends []models.Player
	prefix := getPSPrefix(db)
	// Select players where ID is either p1 or p2 in friendship table, excluding self
	query := fmt.Sprintf(`
		SELECT p.* FROM %splayers p
		JOIN %sfriendships f ON (p.id = f.player1_id OR p.id = f.player2_id)
		WHERE (f.player1_id = $1 OR f.player2_id = $1) AND p.id != $1
	`, prefix, prefix)
	err := db.Select(&friends, query, playerID)
	return friends, err
}

func GetFriendRequests(db *sqlx.DB, playerID int64) (incoming []models.Player, outgoing []models.Player, err error) {
	prefix := getPSPrefix(db)
	// Incoming: where receiver_id = playerID
	queryIn := fmt.Sprintf(`
		SELECT p.* FROM %splayers p
		JOIN %sfriend_requests r ON p.id = r.sender_id
		WHERE r.receiver_id = $1
	`, prefix, prefix)
	err = db.Select(&incoming, queryIn, playerID)
	if err != nil {
		return
	}

	// Outgoing: where sender_id = playerID
	queryOut := fmt.Sprintf(`
		SELECT p.* FROM %splayers p
		JOIN %sfriend_requests r ON p.id = r.receiver_id
		WHERE r.sender_id = $1
	`, prefix, prefix)
	err = db.Select(&outgoing, queryOut, playerID)
	return
}

func DeleteFriendRequest(db *sqlx.DB, senderID, receiverID int64) error {
	prefix := getPSPrefix(db)
	_, err := db.Exec(fmt.Sprintf(`DELETE FROM %sfriend_requests WHERE sender_id=$1 AND receiver_id=$2`, prefix), senderID, receiverID)
	return err
}

func RemoveFriendship(db *sqlx.DB, p1, p2 int64) error {
	prefix := getPSPrefix(db)
	id1, id2 := sortIDs(p1, p2)
	_, err := db.Exec(fmt.Sprintf(`DELETE FROM %sfriendships WHERE player1_id=$1 AND player2_id=$2`, prefix), id1, id2)
	return err
}

func sortIDs(a, b int64) (int64, int64) {
	if a < b {
		return a, b
	}
	return b, a
}
