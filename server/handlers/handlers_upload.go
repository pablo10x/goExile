package handlers

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"exile/server/database"
	"exile/server/models"
	"exile/server/utils"

	"github.com/gorilla/mux"
)

// ServeGameServerFile serves the currently active game_server.zip to spawners.
func ServeGameServerFile(w http.ResponseWriter, r *http.Request) {
	// If DB is connected, try to find the active version
	var filename string = "game_server.zip" // default fallback

	if database.DBConn != nil {
		active, err := database.GetActiveServerVersion(database.DBConn)
		if err != nil {
			// Log error but attempt fallback
			log.Printf("ServeGameServerFile: Error getting active version: %v", err)
		} else if active != nil {
			filename = active.Filename
			w.Header().Set("X-Game-Version", active.Version)
			w.Header().Set("Access-Control-Expose-Headers", "X-Game-Version") // Ensure CORS doesn't block it
		}
	}

	path := filepath.Join("files", filename)

	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// Try fallback if active version is missing
		if filename != "game_server.zip" {
			path = filepath.Join("files", "game_server.zip")
			if _, err := os.Stat(path); os.IsNotExist(err) {
				http.Error(w, "No game server package available", http.StatusNotFound)
				return
			}
		} else {
			http.Error(w, "No game server package available", http.StatusNotFound)
			return
		}
	}

	http.ServeFile(w, r, path)
}

// HandleUploadGameServer accepts a file upload and saves it as a new version.
func HandleUploadGameServer(w http.ResponseWriter, r *http.Request) {
	if database.DBConn == nil {
		utils.WriteError(w, r, http.StatusServiceUnavailable, "Database not connected, cannot track versions")
		return
	}

	// 1. Parse Multipart Form (100MB limit)
	if err := r.ParseMultipartForm(100 << 20); err != nil {
		utils.WriteError(w, r, http.StatusBadRequest, "File too large or invalid form")
		return
	}

	// 2. Retrieve the file
	file, handler, err := r.FormFile("file")
	if err != nil {
		utils.WriteError(w, r, http.StatusBadRequest, "Error retrieving file")
		return
	}
	defer file.Close()

	comment := r.FormValue("comment")
	versionStr := r.FormValue("version")

	// 3. Generate unique filename
	timestamp := time.Now().Unix()
	ext := filepath.Ext(handler.Filename)
	if ext == "" {
		ext = ".zip"
	}
	newFilename := fmt.Sprintf("game_server_%d%s", timestamp, ext)

	// 4. Ensure files directory exists
	if err := os.MkdirAll("./files", 0755); err != nil {
		utils.WriteError(w, r, http.StatusInternalServerError, "Failed to create directory")
		return
	}

	// 5. Create the destination file
	dstPath := filepath.Join("files", newFilename)
	dst, err := os.Create(dstPath)
	if err != nil {
		utils.WriteError(w, r, http.StatusInternalServerError, "Failed to create file on server")
		return
	}
	defer dst.Close()

	// 6. Copy the uploaded file
	if _, err := io.Copy(dst, file); err != nil {
		utils.WriteError(w, r, http.StatusInternalServerError, "Failed to save file")
		return
	}

	// 7. Save metadata to DB
	version := &models.GameServerVersion{
		Filename:   newFilename,
		Version:    versionStr,
		Comment:    comment,
		UploadedAt: time.Now().UTC(),
		IsActive:   false, // Default to inactive, user must activate
	}

	// If it's the first version, maybe make it active?
	// For safety, let's keep it inactive unless it's the very first one.
	versions, _ := database.ListServerVersions(database.DBConn)
	if len(versions) == 0 {
		version.IsActive = true
	}

	if err := database.SaveServerVersion(database.DBConn, version); err != nil {
		// Try to cleanup file if DB insert fails
		os.Remove(dstPath)
		utils.WriteError(w, r, http.StatusInternalServerError, "Failed to save version metadata")
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]string{
		"message":  "File uploaded successfully",
		"filename": newFilename,
	})
}

func ListVersions(w http.ResponseWriter, r *http.Request) {
	if database.DBConn == nil {
		utils.WriteError(w, r, http.StatusServiceUnavailable, "Database not connected")
		return
	}

	versions, err := database.ListServerVersions(database.DBConn)
	if err != nil {
		utils.WriteError(w, r, http.StatusInternalServerError, "Failed to list versions")
		return
	}

	utils.WriteJSON(w, http.StatusOK, versions)
}

func HandleSetActiveVersion(w http.ResponseWriter, r *http.Request) {
	if database.DBConn == nil {
		utils.WriteError(w, r, http.StatusServiceUnavailable, "Database not connected")
		return
	}

	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		utils.WriteError(w, r, http.StatusBadRequest, "Invalid ID")
		return
	}

	if err := database.SetActiveVersion(database.DBConn, id); err != nil {
		utils.WriteError(w, r, http.StatusInternalServerError, "Failed to set active version")
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]string{"message": "Version activated"})
}

func HandleDeleteVersion(w http.ResponseWriter, r *http.Request) {
	if database.DBConn == nil {
		utils.WriteError(w, r, http.StatusServiceUnavailable, "Database not connected")
		return
	}

	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		utils.WriteError(w, r, http.StatusBadRequest, "Invalid ID")
		return
	}

	// Prevent deleting the active version?
	active, _ := database.GetActiveServerVersion(database.DBConn)
	if active != nil && active.ID == id {
		utils.WriteError(w, r, http.StatusBadRequest, "Cannot delete the active version. Activate another version first.")
		return
	}

	filename, err := database.DeleteServerVersion(database.DBConn, id)
	if err != nil {
		utils.WriteError(w, r, http.StatusInternalServerError, "Failed to delete version from DB")
		return
	}

	// Remove file from disk
	if filename != "" {
		os.Remove(filepath.Join("files", filename))
	}

	utils.WriteJSON(w, http.StatusOK, map[string]string{"message": "Version deleted"})
}
