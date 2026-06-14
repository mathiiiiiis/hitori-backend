package db

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

type User struct {
	ID          string    `json:"id"`
	Provider    string    `json:"provider"`
	ProviderID  string    `json:"provider_id"`
	Email       *string   `json:"email,omitempty"`
	DisplayName *string   `json:"display_name,omitempty"`
	AvatarURL   *string   `json:"avatar_url,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Save struct {
	UserID    string          `json:"user_id"`
	Data      json.RawMessage `json:"data"`
	UpdatedAt time.Time       `json:"updated_at"`
}

// UpsertUser creates or updates a user by provider+provider_id
// Returns users UUID
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

// ==== User info ====

type UserInfo struct {
	ID          string    `json:"id"`
	Provider    string    `json:"provider"`
	DisplayName *string   `json:"display_name,omitempty"`
	AvatarURL   *string   `json:"avatar_url,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

func GetUser(ctx context.Context, userID string) (*UserInfo, error) {
	u := &UserInfo{}
	err := Pool.QueryRow(ctx, `
		SELECT id, provider, display_name, avatar_url, created_at
		FROM users WHERE id = $1
	`, userID).Scan(&u.ID, &u.Provider, &u.DisplayName, &u.AvatarURL, &u.CreatedAt)
	if err != nil {
		return nil, err
	}
	return u, nil
}

// ==== Events ====

type Event struct {
	ID        string          `json:"id"`
	UserID    string          `json:"user_id"`
	Type      string          `json:"type"`
	Metadata  json.RawMessage `json:"metadata,omitempty"`
	CreatedAt time.Time       `json:"created_at"`
}

func LogEvent(ctx context.Context, userID, eventType string, metadata json.RawMessage) (*Event, error) {
	e := &Event{}
	err := Pool.QueryRow(ctx, `
		INSERT INTO events (user_id, type, metadata)
		VALUES ($1, $2, $3)
		RETURNING id, user_id, type, metadata, created_at
	`, userID, eventType, metadata).Scan(&e.ID, &e.UserID, &e.Type, &e.Metadata, &e.CreatedAt)
	return e, err
}

func GetRecentEvents(ctx context.Context, userID string, limit int) ([]Event, error) {
	rows, err := Pool.Query(ctx, `
		SELECT id, user_id, type, metadata, created_at
		FROM events
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []Event
	for rows.Next() {
		var e Event
		if err := rows.Scan(&e.ID, &e.UserID, &e.Type, &e.Metadata, &e.CreatedAt); err != nil {
			return nil, err
		}
		events = append(events, e)
	}
	return events, rows.Err()
}

// ==== CLI Sessions ====

func CreateCLISession(ctx context.Context) (string, error) {
	var id string
	err := Pool.QueryRow(ctx, `
		INSERT INTO cli_sessions DEFAULT VALUES RETURNING id
	`).Scan(&id)
	return id, err
}

func CompleteCLISession(ctx context.Context, sessionID, token string) error {
	tag, err := Pool.Exec(ctx, `
		UPDATE cli_sessions SET token = $1
		WHERE id = $2 AND expires_at > now() AND token IS NULL
	`, token, sessionID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("session not found or already completed")
	}
	return nil
}

// PollCLISession returns ("", nil) while pending, (token, nil) when ready
func PollCLISession(ctx context.Context, sessionID string) (string, error) {
	var token *string
	err := Pool.QueryRow(ctx, `
		SELECT token FROM cli_sessions WHERE id = $1 AND expires_at > now()
	`, sessionID).Scan(&token)
	if err != nil {
		return "", err
	}
	if token == nil {
		return "", nil
	}
	//consume
	Pool.Exec(ctx, `DELETE FROM cli_sessions WHERE id = $1`, sessionID)
	return *token, nil
}
