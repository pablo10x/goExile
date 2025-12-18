package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

// -- Notes Handlers --

func ListNotesHandler(w http.ResponseWriter, r *http.Request) {
	if dbConn == nil {
		writeJSON(w, http.StatusOK, []Note{})
		return
	}
	notes, err := GetNotes(dbConn)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, fmt.Sprintf("failed to list notes: %v", err))
		return
	}
	writeJSON(w, http.StatusOK, notes)
}

func CreateNoteHandler(w http.ResponseWriter, r *http.Request) {
	if dbConn == nil {
		writeError(w, r, http.StatusServiceUnavailable, "database not connected")
		return
	}
	var n Note
	if err := decodeJSON(r, &n); err != nil {
		writeError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	// Defaults/Validation
	if n.Color == "" {
		n.Color = "yellow"
	}
	if n.Status == "" {
		n.Status = "normal"
	}
	if n.Rotation == 0 { // Assign a small random rotation for new notes if not provided
		n.Rotation = (rand.Float64() * 4) - 2
	}
	// Set creation/update timestamps on the server side
	n.CreatedAt = time.Now().UTC()
	n.UpdatedAt = time.Now().UTC()

	id, err := SaveNote(dbConn, &n)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, fmt.Sprintf("failed to save note: %v", err))
		return
	}
	n.ID = id
	writeJSON(w, http.StatusCreated, n)
}

func UpdateNoteHandler(w http.ResponseWriter, r *http.Request) {
	if dbConn == nil {
		writeError(w, r, http.StatusServiceUnavailable, "database not connected")
		return
	}
	vars := mux.Vars(r)
	id, err := parseID(vars["id"])
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid id")
		return
	}

	var n Note
	if err := decodeJSON(r, &n); err != nil {
		writeError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	n.ID = id
	// Set UpdatedAt on the server side
	n.UpdatedAt = time.Now().UTC()

	if _, err := SaveNote(dbConn, &n); err != nil {
		writeError(w, r, http.StatusInternalServerError, fmt.Sprintf("failed to update note: %v", err))
		return
	}
	writeJSON(w, http.StatusOK, n)
}

func DeleteNoteHandler(w http.ResponseWriter, r *http.Request) {
	if dbConn == nil {
		writeError(w, r, http.StatusServiceUnavailable, "database not connected")
		return
	}
	vars := mux.Vars(r)
	id, err := parseID(vars["id"])
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid id")
		return
	}

	if err := DeleteNote(dbConn, id); err != nil {
		writeError(w, r, http.StatusInternalServerError, fmt.Sprintf("failed to delete note: %v", err))
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "note deleted"})
}

// -- Todo Handlers --

func ListTodosHandler(w http.ResponseWriter, r *http.Request) {
	if dbConn == nil {
		writeJSON(w, http.StatusOK, []Todo{})
		return
	}
	todos, err := GetTodos(dbConn)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, fmt.Sprintf("failed to list todos: %v", err))
		return
	}
	writeJSON(w, http.StatusOK, todos)
}

func CreateTodoHandler(w http.ResponseWriter, r *http.Request) {
	if dbConn == nil {
		writeError(w, r, http.StatusServiceUnavailable, "database not connected")
		return
	}
	var t Todo
	if err := decodeJSON(r, &t); err != nil {
		writeError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	id, err := SaveTodo(dbConn, &t)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, fmt.Sprintf("failed to save todo: %v", err))
		return
	}
	t.ID = id
	t.CreatedAt = time.Now()
	writeJSON(w, http.StatusCreated, t)
}

func UpdateTodoHandler(w http.ResponseWriter, r *http.Request) {
	if dbConn == nil {
		writeError(w, r, http.StatusServiceUnavailable, "database not connected")
		return
	}
	vars := mux.Vars(r)
	id, err := parseID(vars["id"])
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid id")
		return
	}

	var t Todo
	if err := decodeJSON(r, &t); err != nil {
		writeError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	t.ID = id

	if _, err := SaveTodo(dbConn, &t); err != nil {
		writeError(w, r, http.StatusInternalServerError, fmt.Sprintf("failed to update todo: %v", err))
		return
	}
	writeJSON(w, http.StatusOK, t)
}

func DeleteTodoHandler(w http.ResponseWriter, r *http.Request) {
	if dbConn == nil {
		writeError(w, r, http.StatusServiceUnavailable, "database not connected")
		return
	}
	vars := mux.Vars(r)
	id, err := parseID(vars["id"])
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid id")
		return
	}

	if err := DeleteTodo(dbConn, id); err != nil {
		writeError(w, r, http.StatusInternalServerError, fmt.Sprintf("failed to delete todo: %v", err))
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "todo deleted"})
}

// -- AI Bot --

func AIChatHandler(w http.ResponseWriter, r *http.Request) {
	var req AIChatRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	// In a real implementation, we would call an LLM API here.
	// We would pass the context (notes content, todos) to the LLM.
	
	// Mock Logic for prototype:
	// If message contains "todo", suggest a todo.
	// Otherwise, just chat.

	response := AIChatResponse{
		Response: "I see you're working hard! I've analyzed your notes.",
	}

	if dbConn != nil && (req.Context == "notes" || req.Context == "general") {
		// Read notes to "simulate" reading
		notes, _ := GetNotes(dbConn)
		count := len(notes)
		response.Response += fmt.Sprintf(" You have %d active notes.", count)
	}

	// Simple heuristic for demo
	msgLower := req.Message
	if contains(msgLower, "remind", "todo", "task", "plan") {
		response.Response = "I've added a suggestion to your todo list based on your request."
		response.SuggestedTodo = "Review server logs for errors" // Static for now, could be dynamic
	} else if contains(msgLower, "hello", "hi") {
		response.Response = "Hello! I am your AI assistant. I can help manage your notes and tasks."
	}

	writeJSON(w, http.StatusOK, response)
}

func contains(s string, substrs ...string) bool {
	s = strings.ToLower(s)
	for _, sub := range substrs {
		if strings.Contains(s, sub) {
			return true
		}
	}
	return false
}
