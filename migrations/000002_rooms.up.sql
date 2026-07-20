SET search_path TO voco;

CREATE TABLE IF NOT EXISTS rooms (
    id               UUID PRIMARY KEY,
    title            TEXT        NOT NULL DEFAULT '',
    owner_id         TEXT,
    status           TEXT        NOT NULL DEFAULT 'active'
        CHECK (status IN ('active', 'closed')),
    join_policy      TEXT        NOT NULL DEFAULT 'open_by_link',
    max_participants INT         NOT NULL DEFAULT 10 CHECK (max_participants > 0),
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at       TIMESTAMPTZ,
    closed_at        TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_rooms_owner_id ON rooms (owner_id) WHERE owner_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_rooms_expires_at ON rooms (expires_at) WHERE expires_at IS NOT NULL;
