package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
)

func UserHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		handleGetUser(w, r)
	case http.MethodPost:
		handlePostUser(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleGetUser(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		sendJSONError(w, "invalid id", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		sendJSONError(w, "invalid id", http.StatusBadRequest)
		return
	}

	response := map[string]int{"user_id": id}
	sendJSON(w, response, http.StatusOK)
}

func handlePostUser(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name string `json:"name"`
	}

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		sendJSONError(w, "invalid name", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		sendJSONError(w, "invalid name", http.StatusBadRequest)
		return
	}

	response := map[string]string{"created": req.Name}
	sendJSON(w, response, http.StatusCreated)
}

func sendJSON(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

func sendJSONError(w http.ResponseWriter, message string, statusCode int) {
	response := map[string]string{"error": message}
	sendJSON(w, response, statusCode)
}
