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

// AuthenticatePlayerHandler handles player login via Firebase ID Token
func AuthenticatePlayerHandler(w http.ResponseWriter, r *http.Request) {
	if database.DBConn == nil {
		utils.WriteError(w, r, http.StatusServiceUnavailable, "database not connected")
		return
	}

	if auth.FirebaseMgr == nil {
		utils.WriteError(w, r, http.StatusServiceUnavailable, "firebase not initialized")
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		utils.WriteError(w, r, http.StatusBadRequest, "failed to read body")
		return
	}

	// Restore the body so it can be read again
	r.Body = io.NopCloser(bytes.NewBuffer(body))

	var idToken, name, deviceID string
	contentType := r.Header.Get("Content-Type")

	if bytes.Contains([]byte(contentType), []byte("application/json")) {
		var req struct {
			IDToken  string `json:"id_token"`
			Name     string `json:"name"`
			DeviceID string `json:"device_id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			utils.WriteError(w, r, http.StatusBadRequest, "invalid request")
			return
		}
		idToken = req.IDToken
		name = req.Name
		deviceID = req.DeviceID
	} else {
		// Fallback to form data
		if err := r.ParseForm(); err != nil {
			utils.WriteError(w, r, http.StatusBadRequest, "failed to parse form")
			return
		}
		idToken = r.FormValue("id_token")
		name = r.FormValue("name")
		deviceID = r.FormValue("device_id")
	}

	if idToken == "" {
		utils.WriteError(w, r, http.StatusBadRequest, "id_token is required")
		return
	}

	uid, err := auth.FirebaseMgr.VerifyIDToken(idToken)
	if err != nil {
		utils.WriteError(w, r, http.StatusUnauthorized, "invalid token: "+err.Error())
		return
	}

	// 1. Try to find player by UID
	p, err := database.GetPlayerByUID(database.DBConn, uid)
	if err != nil {
		utils.WriteError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	if p != nil {
		// Existing player found by UID
		// Update info if needed
		updated := false
		if name != "" && p.Name != name {
			p.Name = name
			updated = true
		}
		// If DeviceID changed or wasn't set (migration scenario)
		if deviceID != "" && p.DeviceID != deviceID {
			p.DeviceID = deviceID
			updated = true
		}

		if updated {
			_ = database.UpdatePlayer(database.DBConn, p)
		}
	} else {
		// 2. Try to find player by DeviceID (migration or first link)
		if deviceID != "" {
			p, err = database.GetPlayerByDeviceID(database.DBConn, deviceID)
			if err != nil {
				utils.WriteError(w, r, http.StatusInternalServerError, err.Error())
				return
			}
		}

		if p != nil {
			// Found by DeviceID, but UID was empty. Link them.
			p.UID = uid
			if name != "" {
				p.Name = name
			}
			_, err := database.DBConn.Exec(`UPDATE player_system.players SET uid=$1, name=$2, updated_at=NOW() WHERE id=$3`, uid, p.Name, p.ID)
			if err != nil {
				utils.WriteError(w, r, http.StatusInternalServerError, "failed to link uid: "+err.Error())
				return
			}
		} else {
			// 3. Create new player
			newPlayer := &models.Player{
				UID:      uid,
				Name:     name,
				DeviceID: deviceID,
			}
			if newPlayer.Name == "" {
				newPlayer.Name = "Unknown Player"
			}
			if newPlayer.DeviceID == "" {
				utils.WriteError(w, r, http.StatusBadRequest, "device_id required for new account")
				return
			}

			id, err := database.CreatePlayer(database.DBConn, newPlayer)
			if err != nil {
				utils.WriteError(w, r, http.StatusInternalServerError, "failed to create player: "+err.Error())
				return
			}
			newPlayer.ID = id
			p = newPlayer
		}
	}

	// Enrich with friends info
	friends, _ := database.GetFriends(database.DBConn, p.ID)
	incoming, outgoing, _ := database.GetFriendRequests(database.DBConn, p.ID)

	p.Friends = friends
	p.IncomingFriendRequests = incoming
	p.OutgoingFriendRequests = outgoing

	// Generate WS Session Key
	wsKey := utils.GenerateRandomString(32)
	ws_player.GlobalPlayerWS.RegisterSession(p.ID, wsKey)

	response := map[string]interface{}{
		"player":      p,
		"ws_auth_key": wsKey,
		"ws_endpoint": "/api/game/ws",
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
