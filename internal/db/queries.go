package db

import (
	"context"
	"encoding/json"
	"time"
)

type User struct {
	ID			string		`json:"id"`
	Provider	string		`json:"provider"`
	ProviderID	string		`json:"provider_id"`
	Email		*string		`json:"email,omitempty"`
	DisplayName	*string		`json:"display_name,omitempty"`
	AvatarURL	*string		`json:"avatar_url,omitempty"`
	CreatedAt	time.Time	`json:"created_at"`
	UpdatedAt	time.Time	`json:"updated_at"`
}

type Save struct {
	UserID		string				`json:"user_id"`
	Data		json.RawMessage		`json:"data"`
	UpdatedAt	time.Time			`json:"updated_at"`
}

/// UpsertUser creates or updates a user by provider+provider_id
/// Returns the users UUID
func UpsertUser(ctx context.Context, provider, providerID, email, displayName, avatarURL string) (string, error) {
	var id string
	err := Pool.QueryRow(ctx, `
		INSERT INTO users (provider, provider_id, email, display_name, avatar_url)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (provider, provider_id)
		DO UPDATE SET
			email = EXCLUDED.email,
			display_name = EXCLUDED.display_name,
			avatar_url = EXCLUDED.avatar_url,
			updated_at = now()
		RETURNING id
	`, provider, providerID, email, displayName, avatarURL).Scan(&id)
	return id, err
}

func GetSave(ctx context.Context, userID string) (*Save, error) {
	s := &Save{}
	err := Pool.QueryRow(ctx, `
		SELECT user_id, data, updated_at FROM saves WHERE user_id = $1
	`, userID).Scan(&s.UserID, &s.Data, &s.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func PutSave(ctx context.Context, userID string, data json.RawMessage) (*Save, error) {
	s := &Save{}
	err := Pool.QueryRow(ctx, `
		INSERT INTO saves (user_id, data, updated_at)
		VALUES ($1, $2, now())
		ON CONFLICT (user_id)
		DO UPDATE SET data = EXCLUDED.data, updated_at = now()
		RETURNING user_id, data, updated_at
	`, userID, data).Scan(&s.UserID, &s.Data, &s.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return s, nil
}
