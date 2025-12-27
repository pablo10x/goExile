package models

import "time"

// Player represents a user in the game.
// It is stored in the 'player_system' schema.
type Player struct {
	ID               int64     `json:"id" db:"id"`
	UID              string    `json:"uid" db:"uid"` // Firebase UID
	Name             string    `json:"name" db:"name"`
	DeviceID         string    `json:"device_id" db:"device_id"`
	XP               int64     `json:"xp" db:"xp"`
	Banned           bool      `json:"banned" db:"banned"`
	LastJoinedServer string    `json:"last_joined_server" db:"last_joined_server"`
	CreatedAt        time.Time `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time `json:"updated_at" db:"updated_at"`
	Online           bool      `json:"online"` // Computed field, not in DB

	// Enriched fields (not in DB table directly)
	Friends                []Player `json:"friends,omitempty"`
	IncomingFriendRequests []Player `json:"incoming_friend_requests,omitempty"`
	OutgoingFriendRequests []Player `json:"outgoing_friend_requests,omitempty"`
}

// Friendship represents an established link between two players.
type Friendship struct {
	Player1ID int64     `json:"player1_id" db:"player1_id"`
	Player2ID int64     `json:"player2_id" db:"player2_id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// FriendRequest represents a pending request.
type FriendRequest struct {
	ID         int64     `json:"id" db:"id"`
	SenderID   int64     `json:"sender_id" db:"sender_id"`
	ReceiverID int64     `json:"receiver_id" db:"receiver_id"`
	Status     string    `json:"status" db:"status"` // 'pending', 'rejected'
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}

// Report represents a user report.
type Report struct {
	ID                   int64     `json:"id" db:"id"`
	ReporterID           int64     `json:"reporter_id" db:"reporter_id"`
	ReportedUserID       int64     `json:"reported_user_id" db:"reported_user_id"`
	Reason               string    `json:"reason" db:"reason"`
	GameServerInstanceID string    `json:"game_server_instance_id" db:"game_server_instance_id"`
	Timestamp            time.Time `json:"timestamp" db:"timestamp"`

	// Enriched fields
	ReporterName     string `json:"reporter_name,omitempty" db:"reporter_name"`
	ReportedUserName string `json:"reported_user_name,omitempty" db:"reported_user_name"`
}
