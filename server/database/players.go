package database

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"exile/server/models"

	"github.com/jmoiron/sqlx"
)

// InitPlayerSystem initializes the 'player_system' schema and tables.
func InitPlayerSystem(db *sqlx.DB) error {
	schema := `CREATE SCHEMA IF NOT EXISTS player_system;`
	if _, err := db.Exec(schema); err != nil {
		return fmt.Errorf("create schema: %w", err)
	}

	// Players Table
	playersTable := `CREATE TABLE IF NOT EXISTS player_system.players (
		id BIGSERIAL PRIMARY KEY,
		uid TEXT UNIQUE, -- Firebase UID
		name TEXT NOT NULL,
		device_id TEXT NOT NULL UNIQUE,
		xp BIGINT DEFAULT 0,
		last_joined_server TEXT DEFAULT '',
		created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
		updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
	);`
	if _, err := db.Exec(playersTable); err != nil {
		return fmt.Errorf("create players table: %w", err)
	}

	// Ensure ID sequence starts at 1,000,000 for nice IDs
	var maxID int64
	if err := db.Get(&maxID, "SELECT COALESCE(MAX(id), 0) FROM player_system.players"); err == nil && maxID < 1000000 {
		// Attempt to alter sequence. Ignore error if sequence name differs (unlikely with Postgres defaults)
		_, _ = db.Exec("ALTER SEQUENCE player_system.players_id_seq RESTART WITH 1000000")
	}

	// Migration: Add uid column if it doesn't exist
	_, err := db.Exec(`DO $$
	BEGIN
		IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema='player_system' AND table_name='players' AND column_name='uid') THEN
			ALTER TABLE player_system.players ADD COLUMN uid TEXT UNIQUE;
		END IF;
	END $$;`)
	if err != nil {
		log.Printf("Warning: Failed to add uid column to players: %v", err)
	}

	// Friendships Table (established friends)
	// Constraint: player1_id < player2_id to ensure unique pairings
	friendshipsTable := `CREATE TABLE IF NOT EXISTS player_system.friendships (
		player1_id BIGINT NOT NULL REFERENCES player_system.players(id) ON DELETE CASCADE,
		player2_id BIGINT NOT NULL REFERENCES player_system.players(id) ON DELETE CASCADE,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
		PRIMARY KEY (player1_id, player2_id),
		CONSTRAINT check_order CHECK (player1_id < player2_id)
	);`
	if _, err := db.Exec(friendshipsTable); err != nil {
		return fmt.Errorf("create friendships table: %w", err)
	}

	// Friend Requests Table (pending)
	requestsTable := `CREATE TABLE IF NOT EXISTS player_system.friend_requests (
		id BIGSERIAL PRIMARY KEY,
		sender_id BIGINT NOT NULL REFERENCES player_system.players(id) ON DELETE CASCADE,
		receiver_id BIGINT NOT NULL REFERENCES player_system.players(id) ON DELETE CASCADE,
		status TEXT DEFAULT 'pending',
		created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
		UNIQUE(sender_id, receiver_id)
	);`
	if _, err := db.Exec(requestsTable); err != nil {
		return fmt.Errorf("create friend_requests table: %w", err)
	}

	// Reports Table
	reportsTable := `CREATE TABLE IF NOT EXISTS player_system.reports (
		id BIGSERIAL PRIMARY KEY,
		reporter_id BIGINT NOT NULL REFERENCES player_system.players(id) ON DELETE CASCADE,
		reported_user_id BIGINT NOT NULL REFERENCES player_system.players(id) ON DELETE CASCADE,
		reason TEXT NOT NULL,
		game_server_instance_id TEXT DEFAULT '',
		timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW()
	);`
	if _, err := db.Exec(reportsTable); err != nil {
		return fmt.Errorf("create reports table: %w", err)
	}

	return nil
}

// -- Reports CRUD --

func CreateReport(db *sqlx.DB, r *models.Report) (int64, error) {
	var id int64
	query := `INSERT INTO player_system.reports (reporter_id, reported_user_id, reason, game_server_instance_id, timestamp)
			  VALUES ($1, $2, $3, $4, $5) RETURNING id`

	r.Timestamp = time.Now().UTC()

	err := db.QueryRow(query, r.ReporterID, r.ReportedUserID, r.Reason, r.GameServerInstanceID, r.Timestamp).Scan(&id)
	return id, err
}

func GetAllReports(db *sqlx.DB) ([]models.Report, error) {
	var reports []models.Report
	query := `
		SELECT r.*, p1.name as reporter_name, p2.name as reported_user_name
		FROM player_system.reports r
		JOIN player_system.players p1 ON r.reporter_id = p1.id
		JOIN player_system.players p2 ON r.reported_user_id = p2.id
		ORDER BY r.timestamp DESC
	`
	err := db.Select(&reports, query)
	return reports, err
}

// -- Player CRUD --

func CreatePlayer(db *sqlx.DB, p *models.Player) (int64, error) {
	var id int64
	query := `INSERT INTO player_system.players (uid, name, device_id, xp, last_joined_server, created_at, updated_at)
			  VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`

	p.CreatedAt = time.Now().UTC()
	p.UpdatedAt = time.Now().UTC()

	err := db.QueryRow(query, p.UID, p.Name, p.DeviceID, p.XP, p.LastJoinedServer, p.CreatedAt, p.UpdatedAt).Scan(&id)
	return id, err
}

func GetPlayerByUID(db *sqlx.DB, uid string) (*models.Player, error) {
	var p models.Player
	query := `SELECT * FROM player_system.players WHERE uid = $1`
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
	query := `SELECT * FROM player_system.players WHERE device_id = $1`
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
	query := `SELECT * FROM player_system.players WHERE id = $1`
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
	query := `SELECT * FROM player_system.players ORDER BY id DESC`
	err := db.Select(&players, query)
	return players, err
}

func UpdatePlayer(db *sqlx.DB, p *models.Player) error {
	p.UpdatedAt = time.Now().UTC()
	query := `UPDATE player_system.players SET uid=:uid, name=:name, device_id=:device_id, xp=:xp, last_joined_server=:last_joined_server, updated_at=:updated_at WHERE id=:id`
	_, err := db.NamedExec(query, p)
	return err
}

func DeletePlayer(db *sqlx.DB, id int64) error {
	_, err := db.Exec(`DELETE FROM player_system.players WHERE id = $1`, id)
	return err
}

// -- Friends System --

func SendFriendRequest(db *sqlx.DB, senderID, receiverID int64) error {
	if senderID == receiverID {
		return fmt.Errorf("cannot friend yourself")
	}
	// Check if already friends
	var count int
	p1, p2 := sortIDs(senderID, receiverID)
	err := db.Get(&count, `SELECT COUNT(*) FROM player_system.friendships WHERE player1_id=$1 AND player2_id=$2`, p1, p2)
	if err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf("already friends")
	}

	query := `INSERT INTO player_system.friend_requests (sender_id, receiver_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`
	_, err = db.Exec(query, senderID, receiverID)
	return err
}

func AcceptFriendRequest(db *sqlx.DB, senderID, receiverID int64) error {
	// 1. Delete request
	res, err := db.Exec(`DELETE FROM player_system.friend_requests WHERE sender_id=$1 AND receiver_id=$2`, senderID, receiverID)
	if err != nil {
		return err
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("request not found")
	}

	// 2. Create friendship
	p1, p2 := sortIDs(senderID, receiverID)
	_, err = db.Exec(`INSERT INTO player_system.friendships (player1_id, player2_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`, p1, p2)
	return err
}

func GetFriends(db *sqlx.DB, playerID int64) ([]models.Player, error) {
	var friends []models.Player
	// Select players where ID is either p1 or p2 in friendship table, excluding self
	query := `
		SELECT p.* FROM player_system.players p
		JOIN player_system.friendships f ON (p.id = f.player1_id OR p.id = f.player2_id)
		WHERE (f.player1_id = $1 OR f.player2_id = $1) AND p.id != $1
	`
	err := db.Select(&friends, query, playerID)
	return friends, err
}

func GetFriendRequests(db *sqlx.DB, playerID int64) (incoming []models.Player, outgoing []models.Player, err error) {
	// Incoming: where receiver_id = playerID
	queryIn := `
		SELECT p.* FROM player_system.players p
		JOIN player_system.friend_requests r ON p.id = r.sender_id
		WHERE r.receiver_id = $1
	`
	err = db.Select(&incoming, queryIn, playerID)
	if err != nil {
		return
	}

	// Outgoing: where sender_id = playerID
	queryOut := `
		SELECT p.* FROM player_system.players p
		JOIN player_system.friend_requests r ON p.id = r.receiver_id
		WHERE r.sender_id = $1
	`
	err = db.Select(&outgoing, queryOut, playerID)
	return
}

func DeleteFriendRequest(db *sqlx.DB, senderID, receiverID int64) error {
	_, err := db.Exec(`DELETE FROM player_system.friend_requests WHERE sender_id=$1 AND receiver_id=$2`, senderID, receiverID)
	return err
}

func RemoveFriendship(db *sqlx.DB, p1, p2 int64) error {
	id1, id2 := sortIDs(p1, p2)
	_, err := db.Exec(`DELETE FROM player_system.friendships WHERE player1_id=$1 AND player2_id=$2`, id1, id2)
	return err
}

func sortIDs(a, b int64) (int64, int64) {
	if a < b {
		return a, b
	}
	return b, a
}
