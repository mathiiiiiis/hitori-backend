package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/mathiiiiiis/hitori-backend/internal/auth"
	"github.com/mathiiiiiis/hitori-backend/internal/db"
	"github.com/mathiiiiiis/hitori-backend/internal/handler"
	"github.com/mathiiiiiis/hitori-backend/internal/middleware"
)

func main() {
	godotenv.Load()

	ctx := context.Background()
	if err := db.Connect(ctx); err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	auth.InitJWT()
	auth.InitOAuth()

	mux := http.NewServeMux()

	// auth routes
	mux.HandleFunc("GET /auth/google", handler.GoogleLogin)
	mux.HandleFunc("GET /auth/google/callback", handler.GoogleCallback)
	mux.HandleFunc("GET /auth/discord", handler.DiscordLogin)
	mux.HandleFunc("GET /auth/discord/callback", handler.DiscordCallback)

	// CLI auth flow
	mux.HandleFunc("GET /auth/cli/init", handler.CLIAuthInit)
	mux.HandleFunc("GET /auth/cli/poll", handler.CLIAuthPoll)

	// protected routes
	mux.Handle("GET /me", middleware.Auth(http.HandlerFunc(handler.GetMe)))
	mux.Handle("GET /save", middleware.Auth(http.HandlerFunc(handler.GetSave)))
	mux.Handle("PUT /save", middleware.Auth(http.HandlerFunc(handler.PutSave)))
	mux.Handle("POST /events", middleware.Auth(http.HandlerFunc(handler.PostEvent)))
	mux.Handle("GET /events", middleware.Auth(http.HandlerFunc(handler.GetEvents)))

	// health check
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})

	wrapped := corsMiddleware(mux)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, wrapped))
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := os.Getenv("FRONTEND_URL")
		if origin == "" {
			origin = "https://playhitori.de"
			origin = "http://localhost:5173"
		}

		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, PUT, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
