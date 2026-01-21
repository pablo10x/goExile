package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"exile/server/auth"
	"exile/server/database"
	"exile/server/models"
	"exile/server/utils"
	"exile/server/ws_player"

	"github.com/gorilla/mux"
)

// -- Player Handlers --

// AuthenticatePlayerHandler verifies a Firebase ID token and checks if the player account exists.
// It supports both application/json and application/x-www-form-urlencoded (Unity WWWForm).
//
// Flow:
//  1. Validates database and Firebase connections
//  2. Extracts and verifies the Firebase ID token from the request
//  3. Checks if a player account exists for the given Firebase UID
//  4. Automatically creates a player account if one doesn't exist and 'name' is provided
//  5. Generates a WebSocket authentication key for the session
//
// Request Parameters (JSON or Form):
//   - id_token (required): Firebase ID token for authentication
//   - name (optional): Display name for account creation/update
//   - device_id (optional): Device identifier for account linking
//
// Response (JSON):
//   - player: The full player object
//   - ws_auth_key: WebSocket authentication key for real-time connection
//   - ws_endpoint: Suggested WebSocket endpoint
func AuthenticatePlayerHandler(w http.ResponseWriter, r *http.Request) {
	// ==================== Validation ====================

	if database.DBConn == nil {
		utils.WriteError(w, r, http.StatusServiceUnavailable, "database not connected")
		return
	}

	if auth.FirebaseMgr == nil {
		utils.WriteError(w, r, http.StatusServiceUnavailable, "firebase not initialized")
		return
	}

	// ==================== Parse Request ====================

	var idToken, playerName, deviceID string

	contentType := r.Header.Get("Content-Type")
	if strings.Contains(contentType, "application/json") {
		var req struct {
			IDToken  string `json:"id_token"`
			Name     string `json:"name"`
			DeviceID string `json:"device_id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			utils.WriteError(w, r, http.StatusBadRequest, "invalid JSON request")
			return
		}
		idToken = req.IDToken
		playerName = req.Name
		deviceID = req.DeviceID
	} else {
		// Fallback to Form Data (Unity WWWForm)
		if err := r.ParseForm(); err != nil {
			utils.WriteError(w, r, http.StatusBadRequest, "failed to parse form data")
			return
		}
		idToken = r.FormValue("id_token")
		playerName = r.FormValue("name")
		deviceID = r.FormValue("device_id")
	}

	if idToken == "" {
		utils.WriteError(w, r, http.StatusBadRequest, "id_token is required")
		return
	}

	// ==================== Firebase Authentication ====================

	var uid string
	var err error

	// Dev Mode Bypass: If not in production and token is "dev_token_uid", skip Firebase verification
	isProd := strings.ToLower(os.Getenv("PRODUCTION_MODE")) == "true"
	if !isProd && strings.HasPrefix(idToken, "dev_token_") {
		uid = strings.TrimPrefix(idToken, "dev_token_")
	} else {
		uid, err = auth.FirebaseMgr.VerifyIDToken(idToken)
		if err != nil {
			utils.WriteError(w, r, http.StatusUnauthorized, "invalid token: "+err.Error())
			return
		}
	}

	// ==================== Player Lookup / Creation ====================

	p, err := database.GetPlayerByUID(database.DBConn, uid)
	if err != nil {
		utils.WriteError(w, r, http.StatusInternalServerError, "database error: "+err.Error())
		return
	}

	accountExists := true

	// Auto-Create player if they don't exist
	if p == nil {
		accountExists = false
		if playerName == "" {
			// Placeholder name, will be updated with ID below
			playerName = "Survivor-Pending"
		}
		if deviceID == "" {
			deviceID = "unknown_" + utils.GenerateRandomString(8)
		}

		newPlayer := &models.Player{
			UID:      uid,
			Name:     playerName,
			DeviceID: deviceID,
		}

		id, err := database.CreatePlayer(database.DBConn, newPlayer)
		if err != nil {
			utils.WriteError(w, r, http.StatusInternalServerError, "failed to create player: "+err.Error())
			return
		}
		newPlayer.ID = id
		p = newPlayer

		// If name was default/empty, set it to Survivor-{ID}
		if p.Name == "Survivor-Pending" {
			p.Name = fmt.Sprintf("Survivor-%d", p.ID)
			_ = database.UpdatePlayer(database.DBConn, p)
		}
	} else if playerName != "" && playerName != p.Name {
		// Update name if it changed (and user already exists)
		p.Name = playerName
		_ = database.UpdatePlayer(database.DBConn, p)
	}

	// ==================== WebSocket Session ====================

	wsKey := utils.GenerateRandomString(32)
	ws_player.GlobalPlayerWS.RegisterSession(p.ID, wsKey)

	// ==================== Response ====================

	response := map[string]interface{}{
		"player":         p,
		"account_exists": accountExists,
		"ws_auth_key":    wsKey,
		"ws_endpoint":    "/api/game/ws",
	}

	utils.WriteJSON(w, http.StatusOK, response)
}

// CreateOrGetPlayerHandler handles player login/registration via device_id
func CreateOrGetPlayerHandler(w http.ResponseWriter, r *http.Request) {
	if database.DBConn == nil {
		utils.WriteError(w, r, http.StatusServiceUnavailable, "database not connected")
		return
	}

	var req models.Player
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, r, http.StatusBadRequest, "invalid request")
		return
	}

	if req.DeviceID == "" {
		utils.WriteError(w, r, http.StatusBadRequest, "device_id is required")
		return
	}

	// Check if exists
	p, err := database.GetPlayerByDeviceID(database.DBConn, req.DeviceID)
	if err != nil {
		utils.WriteError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	if p != nil {
		// Update basic info if provided
		if req.Name != "" && req.Name != p.Name {
			p.Name = req.Name
			_ = database.UpdatePlayer(database.DBConn, p)
		}
		utils.WriteJSON(w, http.StatusOK, p)
		return
	}

	// Create new
	if req.Name == "" {
		req.Name = "Unknown Player"
	}
	id, err := database.CreatePlayer(database.DBConn, &req)
	if err != nil {
		utils.WriteError(w, r, http.StatusInternalServerError, "failed to create player: "+err.Error())
		return
	}
	req.ID = id
	utils.WriteJSON(w, http.StatusCreated, req)
}

func GetPlayerDetailsHandler(w http.ResponseWriter, r *http.Request) {
	if database.DBConn == nil {
		utils.WriteError(w, r, http.StatusServiceUnavailable, "database not connected")
		return
	}

	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.WriteError(w, r, http.StatusBadRequest, "invalid id")
		return
	}

	p, err := database.GetPlayerByID(database.DBConn, id)
	if err != nil {
		utils.WriteError(w, r, http.StatusInternalServerError, err.Error())
		return
	}
	if p == nil {
		utils.WriteError(w, r, http.StatusNotFound, "player not found")
		return
	}

	// Enrich with friends info
	friends, _ := database.GetFriends(database.DBConn, id)
	incoming, outgoing, _ := database.GetFriendRequests(database.DBConn, id)

	p.Friends = friends
	p.IncomingFriendRequests = incoming
	p.OutgoingFriendRequests = outgoing

	utils.WriteJSON(w, http.StatusOK, p)
}

func ListAllPlayersHandler(w http.ResponseWriter, r *http.Request) {
	if database.DBConn == nil {
		utils.WriteError(w, r, http.StatusServiceUnavailable, "database not connected")
		return
	}

	players, err := database.GetAllPlayers(database.DBConn)
	if err != nil {
		utils.WriteError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	for i := range players {
		players[i].Online = ws_player.GlobalPlayerWS.IsPlayerOnline(players[i].ID)
	}

	utils.WriteJSON(w, http.StatusOK, players)
}

func UpdatePlayerDetailsHandler(w http.ResponseWriter, r *http.Request) {
	if database.DBConn == nil {
		utils.WriteError(w, r, http.StatusServiceUnavailable, "database not connected")
		return
	}

	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.WriteError(w, r, http.StatusBadRequest, "invalid id")
		return
	}

	var req models.Player
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, r, http.StatusBadRequest, "invalid request")
		return
	}

	// Fetch existing to ensure it exists
	p, err := database.GetPlayerByID(database.DBConn, id)
	if err != nil {
		utils.WriteError(w, r, http.StatusInternalServerError, err.Error())
		return
	}
	if p == nil {
		utils.WriteError(w, r, http.StatusNotFound, "player not found")
		return
	}

	// Update fields
	p.Name = req.Name
	p.UID = req.UID
	p.DeviceID = req.DeviceID
	p.XP = req.XP
	// p.LastJoinedServer could be updated if we wanted to exposed it

	if err := database.UpdatePlayer(database.DBConn, p); err != nil {
		utils.WriteError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteJSON(w, http.StatusOK, p)
}

func DeletePlayerHandler(w http.ResponseWriter, r *http.Request) {
	if database.DBConn == nil {
		utils.WriteError(w, r, http.StatusServiceUnavailable, "database not connected")
		return
	}

	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.WriteError(w, r, http.StatusBadRequest, "invalid id")
		return
	}

	if err := database.DeletePlayer(database.DBConn, id); err != nil {
		utils.WriteError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]string{"message": "player deleted"})
}

func BanPlayerHandler(w http.ResponseWriter, r *http.Request) {
	if database.DBConn == nil {
		utils.WriteError(w, r, http.StatusServiceUnavailable, "database not connected")
		return
	}

	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.WriteError(w, r, http.StatusBadRequest, "invalid id")
		return
	}

	var req struct {
		Banned bool `json:"banned"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, r, http.StatusBadRequest, "invalid request")
		return
	}

	p, err := database.GetPlayerByID(database.DBConn, id)
	if err != nil {
		utils.WriteError(w, r, http.StatusInternalServerError, err.Error())
		return
	}
	if p == nil {
		utils.WriteError(w, r, http.StatusNotFound, "player not found")
		return
	}

	p.Banned = req.Banned
	if err := database.UpdatePlayer(database.DBConn, p); err != nil {
		utils.WriteError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteJSON(w, http.StatusOK, p)
}

// -- Friend System Handlers --

func SendFriendRequestHandler(w http.ResponseWriter, r *http.Request) {
	if database.DBConn == nil {
		utils.WriteError(w, r, http.StatusServiceUnavailable, "database not connected")
		return
	}

	var req struct {
		SenderID   int64 `json:"sender_id"`
		ReceiverID int64 `json:"receiver_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, r, http.StatusBadRequest, "invalid request")
		return
	}

	if err := database.SendFriendRequest(database.DBConn, req.SenderID, req.ReceiverID); err != nil {
		utils.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]string{"message": "request sent"})
}

func AcceptFriendRequestHandler(w http.ResponseWriter, r *http.Request) {
	if database.DBConn == nil {
		utils.WriteError(w, r, http.StatusServiceUnavailable, "database not connected")
		return
	}

	var req struct {
		SenderID   int64 `json:"sender_id"`
		ReceiverID int64 `json:"receiver_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, r, http.StatusBadRequest, "invalid request")
		return
	}

	// NOTE: In the request logic, 'Sender' is the one who sent the friend request.
	// 'Receiver' is the one accepting it (the current user).
	if err := database.AcceptFriendRequest(database.DBConn, req.SenderID, req.ReceiverID); err != nil {
		utils.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]string{"message": "request accepted"})
}
