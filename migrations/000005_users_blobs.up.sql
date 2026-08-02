-- +goose не используем; golang-migrate.
SET search_path TO voco;

CREATE TABLE IF NOT EXISTS blobs (
    id            UUID PRIMARY KEY,
    owner_user_id UUID,
    content_type  TEXT        NOT NULL,
    byte_size     BIGINT      NOT NULL CHECK (byte_size >= 0),
    sha256        TEXT        NOT NULL DEFAULT '',
    data          BYTEA       NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS users (
    id              UUID PRIMARY KEY,
    keycloak_sub    TEXT        NOT NULL UNIQUE,
    nickname        TEXT        NOT NULL DEFAULT '',
    email           TEXT        NOT NULL DEFAULT '',
    display_name    TEXT        NOT NULL DEFAULT '',
    avatar_blob_id  UUID REFERENCES blobs(id) ON DELETE SET NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    last_seen_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_users_nickname_lower
    ON users (lower(nickname))
    WHERE nickname <> '';

CREATE INDEX IF NOT EXISTS idx_users_email ON users (email);

ALTER TABLE blobs
    ADD CONSTRAINT blobs_owner_user_id_fkey
    FOREIGN KEY (owner_user_id) REFERENCES users(id) ON DELETE SET NULL;

ALTER TABLE rooms
    ALTER COLUMN owner_id TYPE UUID USING NULL;

COMMENT ON COLUMN rooms.owner_id IS 'voco.users.id when known';
