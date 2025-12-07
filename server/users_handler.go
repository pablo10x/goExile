package main

import (
	"net/http"
)

// UsersPage serves the HTML page for users.
func UsersPage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	http.ServeFile(w, r, "webpage/users.html")
}

