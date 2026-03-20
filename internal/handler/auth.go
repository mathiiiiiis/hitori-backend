package handler

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/mathiiiiiis/hitori-backend/internal/auth"
	"github.com/mathiiiiiis/hitori-backend/internal/db"
)

var frontendURL string

func init() {
	frontendURL = os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "https://hitori.mathiiis.de"
	}
}

func randomState() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

/// GET /auth/google
func GoogleLogin(w http.ResponseWriter, r *http.Request) {
	state := randomState()
	// TODO: store state in a short lived cookie for CSRF validation
	http.Redirect(w, r, auth.GoogleConfig.AuthCodeURL(state), http.StatusTemporaryRedirect)
}

/// GET /auth/google/callback
func GoogleCallback(w http.ResponseWriter, r *http.Request) {
	// TODO: validate state param against cookie
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "missing code", http.StatusBadRequest)
		return
	}

	token, err := auth.GoogleConfig.Exchange(r.Context(), code)
	if err != nil {
		http.Error(w, "token exchange failed", http.StatusInternalServerError)
		return
	}

	client := auth.GoogleConfig.Client(r.Context(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v3/userinfo")
	if err != nil {
		http.Error(w, "failed to get user info", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	var info struct {
		Sub     string `json:"sub"`
		Email   string `json:"email"`
		Name    string `json:"name"`
		Picture string `json:"picture"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		http.Error(w, "failed to decode user info", http.StatusInternalServerError)
		return
	}

	userID, err := db.UpsertUser(r.Context(), "google", info.Sub, info.Email, info.Name, info.Picture)
	if err != nil {
		log.Printf("google upsert error: %v", err)
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}

	jwt, err := auth.IssueToken(userID)
	if err != nil {
		http.Error(w, "token error", http.StatusInternalServerError)
		return
	}

	//redirect to frontend with token as fragment (never hits server)
	http.Redirect(w, r, fmt.Sprintf("%s/auth/callback#token=%s", frontendURL, jwt), http.StatusTemporaryRedirect)
}

/// GET /auth/discord
func DiscordLogin(w http.ResponseWriter, r *http.Request) {
	state := randomState()
	http.Redirect(w, r, auth.DiscordConfig.AuthCodeURL(state), http.StatusTemporaryRedirect)
}

/// GET /auth/discord/callback
func DiscordCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "missing code", http.StatusBadRequest)
		return
	}

	token, err := auth.DiscordConfig.Exchange(r.Context(), code)
	if err != nil {
		http.Error(w, "token exchange failed", http.StatusInternalServerError)
		return
	}

	req, _ := http.NewRequestWithContext(r.Context(), "GET", "https://discord.com/api/v10/users/@me", nil)
	req.Header.Set("Authorization", "Bearer "+token.AccessToken)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		http.Error(w, "failed to get user info", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var info struct {
		ID       string `json:"id"`
		Email    string `json:"email"`
		Username string `json:"username"`
		Avatar   string `json:"avatar"`
	}
	if err := json.Unmarshal(body, &info); err != nil {
		http.Error(w, "failed to decode user info", http.StatusInternalServerError)
		return
	}

	avatarURL := ""
	if info.Avatar != "" {
		avatarURL = fmt.Sprintf("https://cdn.discordapp.com/avatars/%s/%s.png", info.ID, info.Avatar)
	}

	userID, err := db.UpsertUser(r.Context(), "discord", info.ID, info.Email, info.Username, avatarURL)
	if err != nil {
		log.Printf("discord upsert error: %v", err)
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}

	jwt, err := auth.IssueToken(userID)
	if err != nil {
		http.Error(w, "token error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("%s/auth/callback#token=%s", frontendURL, jwt), http.StatusTemporaryRedirect)
}
