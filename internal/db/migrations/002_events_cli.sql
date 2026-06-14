-- 002_events_cli.sql

CREATE TABLE IF NOT EXISTS events (
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type       TEXT        NOT NULL,
    metadata   JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS events_user_created ON events(user_id, created_at DESC);

-- short-lived sessions for CLI OAuth flow (device-code style)
CREATE TABLE IF NOT EXISTS cli_sessions (
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    token      TEXT,                          -- null until auth completes
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at TIMESTAMPTZ NOT NULL DEFAULT now() + INTERVAL '10 minutes'
);
