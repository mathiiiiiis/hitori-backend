package auth

import (
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var (
	GoogleConfig *oauth2.Config
	DiscordConfig *oauth2.Config
)

func InitOAuth() {
	baseURL := os.Getenv("BASE_URL")

	GoogleConfig = &oauth2.Config{
		ClientID:		os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret:	os.Getenv("GOOGLE_CLIENT_SECRET"),
		RedirectURL: 	baseURL + "/auth/google/callback",
		Scopes:			[]string{"openid", "email", "profile"},
		Endpoint:		google.Endpoint,
	}

	DiscordConfig = &oauth2.Config{
		ClientID:		os.Getenv("DISCORD_CLIENT_ID"),
		ClientSecret:	os.Getenv("DISCORD_CLIENT_SECRET"),
		RedirectURL: 	baseURL + "/auth/discord/callback",
		Scopes:			[]string{"identify", "email"},
		Endpoint:		oauth2.Endpoint{
			AuthURL: "https://discord.com/api/oauth2/authorize",
			TokenURL: "https://discord.com/api/oauth2/token",
		},
	}
}
