CREATE TABLE IF NOT EXISTS voco.ntfy_users (
    id UUID PRIMARY KEY,
    ntfy_topic TEXT NOT NULL UNIQUE,
    display_name TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_ntfy_users_topic ON voco.ntfy_users (ntfy_topic);
