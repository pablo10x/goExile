package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"exile/server/auth"
	"exile/server/database"
	"exile/server/models"
	"exile/server/utils"
	"exile/server/ws_player"

	"github.com/gorilla/mux"
)

// -- Player Handlers --

// AuthenticatePlayerHandler handles player login via Firebase ID Token.
//
// This handler performs three-stage player resolution:
// 1. Lookup by Firebase UID (primary identifier)
// 2. Lookup by DeviceID if UID not found (migration/linking scenario)
// 3. Create new player if neither UID nor DeviceID matches
//
// Request body supports both JSON and form-encoded data:
//   - id_token (required): Firebase ID token for authentication
//   - name (optional): Player display name
//   - device_id (optional for existing, required for new players): Unique device identifier
//
// Response includes:
//   - player: Full player object with friends and friend requests
//   - ws_auth_key: WebSocket authentication key for real-time connection
//   - ws_endpoint: WebSocket endpoint path
func AuthenticatePlayerHandler(w http.ResponseWriter, r *http.Request) {
	// Verify database connection is available
	if database.DBConn == nil {
		utils.WriteError(w, r, http.StatusServiceUnavailable, "database not connected")
		return
	}

	// Verify Firebase authentication is initialized
	if auth.FirebaseMgr == nil {
		utils.WriteError(w, r, http.StatusServiceUnavailable, "firebase not initialized")
		return
	}

	// Read request body into memory for potential re-parsing
	body, err := io.ReadAll(r.Body)
	if err != nil {
		utils.WriteError(w, r, http.StatusBadRequest, "failed to read body")
		return
	}

	// Restore the body so it can be read again by the decoder
	r.Body = io.NopCloser(bytes.NewBuffer(body))

	var idToken string

	var accountexist bool

	//form-encoded data
	if err := r.ParseForm(); err != nil {
		utils.WriteError(w, r, http.StatusBadRequest, "failed to parse form")
		return
	}
	idToken = r.FormValue("id_token")

	// Validate that ID token is provided
	if idToken == "" {
		utils.WriteError(w, r, http.StatusBadRequest, "id_token is required")
		return
	}

	// Verify the Firebase ID token and extract the user ID (UID)
	uid, err := auth.FirebaseMgr.VerifyIDToken(idToken)
	if err != nil {
		utils.WriteError(w, r, http.StatusUnauthorized, "invalid token: "+err.Error())
		return
	}

	// Stage 1: Attempt to find player by Firebase UID
	p, err := database.GetPlayerByUID(database.DBConn, uid)
	if err != nil {
		utils.WriteError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	if p != nil {
		accountexist = true
	}

	// Enrich player data with social information

	/* 	friends, _ := database.GetFriends(database.DBConn, p.ID)
	   	incoming, outgoing, _ := database.GetFriendRequests(database.DBConn, p.ID)

	   	p.Friends = friends
	   	p.IncomingFriendRequests = incoming
	   	p.OutgoingFriendRequests = outgoing*/

	// Generate WebSocket authentication key for real-time communication
	wsKey := utils.GenerateRandomString(32)
	ws_player.GlobalPlayerWS.RegisterSession(p.ID, wsKey)

	// Build response with player data and WebSocket credentials
	response := map[string]interface{}{
		"accountexist": accountexist,
		"ws_auth_key":  wsKey,
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
