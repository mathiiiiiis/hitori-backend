package handler

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/mathiiiiiis/hitori-backend/internal/db"
	"github.com/mathiiiiiis/hitori-backend/internal/middleware"
	"github.com/jackc/pgx/v5"
)

/// GET /save
func GetSave(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	save, err := db.GetSave(r.Context(), userID)
	if err == pgx.ErrNoRows {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"data":{},"updated_at":null}`))
		return
	}
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(save)
}

/// PUT /save
func PutSave(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	body, err := io.ReadAll(io.LimitReader(r.Body, 5<<20))
	if err != nil {
		http.Error(w, "read error", http.StatusBadRequest)
		return
	}

	if !json.Valid(body) {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	save, err := db.PutSave(r.Context(), userID, body)
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(save)
}
