package handlers

import (
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"exile/server/database"
	"exile/server/models"
	"exile/server/utils"

	"github.com/gorilla/mux"
)

// -- Notes Handlers --

func ListNotesHandler(w http.ResponseWriter, r *http.Request) {
	if database.DBConn == nil {
		utils.WriteJSON(w, http.StatusOK, []models.Note{})
		return
	}
	notes, err := database.GetNotes(database.DBConn)
	if err != nil {
		utils.WriteError(w, r, http.StatusInternalServerError, fmt.Sprintf("failed to list notes: %v", err))
		return
	}
	utils.WriteJSON(w, http.StatusOK, notes)
}

func CreateNoteHandler(w http.ResponseWriter, r *http.Request) {
	if database.DBConn == nil {
		utils.WriteError(w, r, http.StatusServiceUnavailable, "database not connected")
		return
	}
	var n models.Note
	if err := utils.DecodeJSON(r, &n); err != nil {
		utils.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	// Input Validation
	if len(n.Title) > 255 {
		utils.WriteError(w, r, http.StatusBadRequest, "title too long (max 255 chars)")
		return
	}
	if len(n.Content) > 10000 {
		utils.WriteError(w, r, http.StatusBadRequest, "content too long (max 10000 chars)")
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

	id, err := database.SaveNote(database.DBConn, &n)
	if err != nil {
		utils.WriteError(w, r, http.StatusInternalServerError, fmt.Sprintf("failed to save note: %v", err))
		return
	}
	n.ID = id
	utils.WriteJSON(w, http.StatusCreated, n)
}

func UpdateNoteHandler(w http.ResponseWriter, r *http.Request) {
	if database.DBConn == nil {
		utils.WriteError(w, r, http.StatusServiceUnavailable, "database not connected")
		return
	}
	vars := mux.Vars(r)
	id, err := utils.ParseID(vars["id"])
	if err != nil {
		utils.WriteError(w, r, http.StatusBadRequest, "invalid id")
		return
	}

	var n models.Note
	if err := utils.DecodeJSON(r, &n); err != nil {
		utils.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	// Input Validation
	if len(n.Title) > 255 {
		utils.WriteError(w, r, http.StatusBadRequest, "title too long (max 255 chars)")
		return
	}
	if len(n.Content) > 10000 {
		utils.WriteError(w, r, http.StatusBadRequest, "content too long (max 10000 chars)")
		return
	}

	n.ID = id
	// Set UpdatedAt on the server side
	n.UpdatedAt = time.Now().UTC()

	if _, err := database.SaveNote(database.DBConn, &n); err != nil {
		utils.WriteError(w, r, http.StatusInternalServerError, fmt.Sprintf("failed to update note: %v", err))
		return
	}
	utils.WriteJSON(w, http.StatusOK, n)
}

func DeleteNoteHandler(w http.ResponseWriter, r *http.Request) {
	if database.DBConn == nil {
		utils.WriteError(w, r, http.StatusServiceUnavailable, "database not connected")
		return
	}
	vars := mux.Vars(r)
	id, err := utils.ParseID(vars["id"])
	if err != nil {
		utils.WriteError(w, r, http.StatusBadRequest, "invalid id")
		return
	}

	if err := database.DeleteNote(database.DBConn, id); err != nil {
		utils.WriteError(w, r, http.StatusInternalServerError, fmt.Sprintf("failed to delete note: %v", err))
		return
	}
	utils.WriteJSON(w, http.StatusOK, map[string]string{"message": "note deleted"})
}

// -- models.Todo Handlers --

func ListTodosHandler(w http.ResponseWriter, r *http.Request) {
	if database.DBConn == nil {
		utils.WriteJSON(w, http.StatusOK, []models.Todo{})
		return
	}
	todos, err := database.GetTodos(database.DBConn)
	if err != nil {
		utils.WriteError(w, r, http.StatusInternalServerError, fmt.Sprintf("failed to list todos: %v", err))
		return
	}
	utils.WriteJSON(w, http.StatusOK, todos)
}

func CreateTodoHandler(w http.ResponseWriter, r *http.Request) {
	if database.DBConn == nil {
		utils.WriteError(w, r, http.StatusServiceUnavailable, "database not connected")
		return
	}
	var t models.Todo
	if err := utils.DecodeJSON(r, &t); err != nil {
		utils.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	id, err := database.SaveTodo(database.DBConn, &t)
	if err != nil {
		utils.WriteError(w, r, http.StatusInternalServerError, fmt.Sprintf("failed to save todo: %v", err))
		return
	}
	t.ID = id
	t.CreatedAt = time.Now()
	utils.WriteJSON(w, http.StatusCreated, t)
}

func UpdateTodoHandler(w http.ResponseWriter, r *http.Request) {
	if database.DBConn == nil {
		utils.WriteError(w, r, http.StatusServiceUnavailable, "database not connected")
		return
	}
	vars := mux.Vars(r)
	id, err := utils.ParseID(vars["id"])
	if err != nil {
		utils.WriteError(w, r, http.StatusBadRequest, "invalid id")
		return
	}

	var t models.Todo
	if err := utils.DecodeJSON(r, &t); err != nil {
		utils.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	t.ID = id

	if _, err := database.SaveTodo(database.DBConn, &t); err != nil {
		utils.WriteError(w, r, http.StatusInternalServerError, fmt.Sprintf("failed to update todo: %v", err))
		return
	}
	utils.WriteJSON(w, http.StatusOK, t)
}

func DeleteTodoHandler(w http.ResponseWriter, r *http.Request) {
	if database.DBConn == nil {
		utils.WriteError(w, r, http.StatusServiceUnavailable, "database not connected")
		return
	}
	vars := mux.Vars(r)
	id, err := utils.ParseID(vars["id"])
	if err != nil {
		utils.WriteError(w, r, http.StatusBadRequest, "invalid id")
		return
	}

	if err := database.DeleteTodo(database.DBConn, id); err != nil {
		utils.WriteError(w, r, http.StatusInternalServerError, fmt.Sprintf("failed to delete todo: %v", err))
		return
	}
	utils.WriteJSON(w, http.StatusOK, map[string]string{"message": "todo deleted"})
}

func CreateTodoCommentHandler(w http.ResponseWriter, r *http.Request) {
	if database.DBConn == nil {
		utils.WriteError(w, r, http.StatusServiceUnavailable, "database not connected")
		return
	}
	vars := mux.Vars(r)
	todoID, err := utils.ParseID(vars["id"])
	if err != nil {
		utils.WriteError(w, r, http.StatusBadRequest, "invalid todo id")
		return
	}

	var c models.TodoComment
	if err := utils.DecodeJSON(r, &c); err != nil {
		utils.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	c.TodoID = todoID
	c.CreatedAt = time.Now().UTC()

	id, err := database.SaveTodoComment(database.DBConn, &c)
	if err != nil {
		utils.WriteError(w, r, http.StatusInternalServerError, fmt.Sprintf("failed to save comment: %v", err))
		return
	}
	c.ID = id
	utils.WriteJSON(w, http.StatusCreated, c)
}

func DeleteTodoCommentHandler(w http.ResponseWriter, r *http.Request) {
	if database.DBConn == nil {
		utils.WriteError(w, r, http.StatusServiceUnavailable, "database not connected")
		return
	}
	vars := mux.Vars(r)
	id, err := utils.ParseID(vars["comment_id"])
	if err != nil {
		utils.WriteError(w, r, http.StatusBadRequest, "invalid comment id")
		return
	}

	if err := database.DeleteTodoComment(database.DBConn, id); err != nil {
		utils.WriteError(w, r, http.StatusInternalServerError, fmt.Sprintf("failed to delete comment: %v", err))
		return
	}
	utils.WriteJSON(w, http.StatusOK, map[string]string{"message": "comment deleted"})
}

// -- AI Bot --

func AIChatHandler(w http.ResponseWriter, r *http.Request) {
	var req models.AIChatRequest
	if err := utils.DecodeJSON(r, &req); err != nil {
		utils.WriteError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	// In a real implementation, we would call an LLM API here.
	// We would pass the context (notes content, todos) to the LLM.

	// Mock Logic for prototype:
	// If message contains "todo", suggest a todo.
	// Otherwise, just chat.

	response := models.AIChatResponse{
		Response: "I see you're working hard! I've analyzed your notes.",
	}

	if database.DBConn != nil && (req.Context == "notes" || req.Context == "general") {
		// Read notes to "simulate" reading
		notes, _ := database.GetNotes(database.DBConn)
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

	utils.WriteJSON(w, http.StatusOK, response)
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
