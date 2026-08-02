SET search_path TO voco;

CREATE TABLE IF NOT EXISTS calendar_events (
    id           UUID PRIMARY KEY,
    organizer_id UUID NOT NULL REFERENCES users(id),
    title        TEXT NOT NULL,
    description  TEXT NOT NULL DEFAULT '',
    starts_at    TIMESTAMPTZ NOT NULL,
    ends_at      TIMESTAMPTZ NOT NULL,
    timezone     TEXT NOT NULL DEFAULT 'UTC',
    status       TEXT NOT NULL DEFAULT 'scheduled'
        CHECK (status IN ('scheduled', 'cancelled')),
    room_id      UUID REFERENCES rooms(id) ON DELETE SET NULL,
    rrule        TEXT,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (ends_at > starts_at)
);

CREATE INDEX IF NOT EXISTS idx_calendar_events_range ON calendar_events (starts_at, ends_at);
CREATE INDEX IF NOT EXISTS idx_calendar_events_organizer ON calendar_events (organizer_id);

CREATE TABLE IF NOT EXISTS event_attendees (
    event_id UUID NOT NULL REFERENCES calendar_events(id) ON DELETE CASCADE,
    user_id  UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status   TEXT NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'accepted', 'tentative', 'declined')),
    PRIMARY KEY (event_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_event_attendees_user ON event_attendees (user_id);

CREATE TABLE IF NOT EXISTS event_reminders (
    id                    UUID PRIMARY KEY,
    event_id              UUID NOT NULL REFERENCES calendar_events(id) ON DELETE CASCADE,
    remind_before_minutes INT  NOT NULL CHECK (remind_before_minutes > 0),
    UNIQUE (event_id, remind_before_minutes)
);

CREATE TABLE IF NOT EXISTS reminder_deliveries (
    id          UUID PRIMARY KEY,
    reminder_id UUID NOT NULL REFERENCES event_reminders(id) ON DELETE CASCADE,
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    fire_at     TIMESTAMPTZ NOT NULL,
    sent_at     TIMESTAMPTZ,
    UNIQUE (reminder_id, user_id, fire_at)
);

CREATE INDEX IF NOT EXISTS idx_reminder_deliveries_pending
    ON reminder_deliveries (fire_at) WHERE sent_at IS NULL;
