package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/mathiiiiiis/hitori-backend/internal/db"
	"github.com/mathiiiiiis/hitori-backend/internal/middleware"
)

var validEventTypes = map[string]bool{
	"commit":        true,
	"build_success": true,
	"build_fail":    true,
	"idle":          true,
	"active":        true,
	"feed":          true,
	"pet":           true,
}

// POST /events
func PostEvent(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	body, err := io.ReadAll(io.LimitReader(r.Body, 64*1024))
	if err != nil {
		http.Error(w, "read error", http.StatusBadRequest)
		return
	}

	var req struct {
		Type     string          `json:"type"`
		Metadata json.RawMessage `json:"metadata,omitempty"`
	}
	if err := json.Unmarshal(body, &req); err != nil || req.Type == "" {
		http.Error(w, "invalid request: missing type", http.StatusBadRequest)
		return
	}
	if !validEventTypes[req.Type] {
		http.Error(w, "unknown event type", http.StatusBadRequest)
		return
	}

	event, err := db.LogEvent(r.Context(), userID, req.Type, req.Metadata)
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(event)
}

// GET /events?limit=20
func GetEvents(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	limit := 20
	if l := r.URL.Query().Get("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 && n <= 100 {
			limit = n
		}
	}

	events, err := db.GetRecentEvents(r.Context(), userID, limit)
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}
	if events == nil {
		events = []db.Event{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(events)
}
