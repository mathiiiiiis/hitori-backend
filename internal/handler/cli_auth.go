package handler

import (
	"encoding/json"
	"net/http"

	"github.com/mathiiiiiis/hitori-backend/internal/auth"
	"github.com/mathiiiiiis/hitori-backend/internal/db"
)

// GET /auth/cli/init?provider=discord|google
// Creates pending CLI session and returns the OAuth URL to open in browser
func CLIAuthInit(w http.ResponseWriter, r *http.Request) {
	provider := r.URL.Query().Get("provider")
	if provider != "discord" && provider != "google" {
		provider = "discord"
	}

	sessionID, err := db.CreateCLISession(r.Context())
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}

	// embed session ID in OAuth state: "cli:<session_id>"
	state := "cli:" + sessionID

	var authURL string
	switch provider {
	case "google":
		authURL = auth.GoogleConfig.AuthCodeURL(state)
	default:
		authURL = auth.DiscordConfig.AuthCodeURL(state)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"session_id": sessionID,
		"auth_url":   authURL,
	})
}

// GET /auth/cli/poll?session_id=X
// Returns {"status":"pending"} while waiting, {"status":"ready","token":"..."} when done
func CLIAuthPoll(w http.ResponseWriter, r *http.Request) {
	sessionID := r.URL.Query().Get("session_id")
	if sessionID == "" {
		http.Error(w, "missing session_id", http.StatusBadRequest)
		return
	}

	token, err := db.PollCLISession(r.Context(), sessionID)
	if err != nil {
		http.Error(w, "session not found or expired", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if token == "" {
		json.NewEncoder(w).Encode(map[string]string{"status": "pending"})
	} else {
		json.NewEncoder(w).Encode(map[string]string{"status": "ready", "token": token})
	}
}

// cliSuccessPage is shown after user authenticates via a CLI-initiated flow
var cliSuccessPage = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>Hitori – Authenticated</title>
  <style>
    body { font-family: system-ui, sans-serif; background: #0e0e14; color: #fff;
           display: flex; flex-direction: column; align-items: center;
           justify-content: center; height: 100vh; margin: 0; }
    h1   { font-size: 2rem; color: #688BFF; margin-bottom: .5rem; }
    p    { color: #aaa; }
  </style>
</head>
<body>
	<h1>Authenticated :D</h1>
  <p>You can close this tab and return to your terminal.</p>
</body>
</html>`
