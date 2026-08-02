package postgres

import (
	"context"
	"errors"
	"time"

	pgadapter "voco/internal/adapters/postgres"
	"voco/internal/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type RoomRepo struct {
	db *pgadapter.Client
}

func NewRoomRepo(db *pgadapter.Client) *RoomRepo {
	return &RoomRepo{db: db}
}

func (r *RoomRepo) Get(ctx context.Context, id domain.RoomID) (domain.Room, bool, error) {
	var room domain.Room
	var rid uuid.UUID
	var owner *uuid.UUID
	var expires, closed *time.Time
	var status, policy string
	err := r.db.QueryRow(ctx, `
		SELECT id, title, owner_id, status, join_policy, max_participants, created_at, expires_at, closed_at
		FROM rooms WHERE id = $1
	`, id.UUID()).Scan(
		&rid, &room.Title, &owner, &status, &policy, &room.MaxParticipants,
		&room.CreatedAt, &expires, &closed,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Room{}, false, nil
	}
	if err != nil {
		return domain.Room{}, false, err
	}
	room.ID = domain.RoomID(rid)
	room.Status = domain.RoomStatus(status)
	room.JoinPolicy = domain.JoinPolicy(policy)
	if owner != nil {
		uid := domain.UserID(*owner)
		room.Owner = &uid
	}
	if expires != nil {
		room.ExpiresAt = *expires
	}
	if closed != nil {
		room.ClosedAt = *closed
	}
	return room, true, nil
}

func (r *RoomRepo) Upsert(ctx context.Context, room domain.Room, _ time.Duration) error {
	var owner any
	if room.Owner != nil {
		owner = *room.Owner
	}
	var expires, closed any
	if !room.ExpiresAt.IsZero() {
		expires = room.ExpiresAt
	}
	if !room.ClosedAt.IsZero() {
		closed = room.ClosedAt
	}
	_, err := r.db.Exec(ctx, `
		INSERT INTO rooms (id, title, owner_id, status, join_policy, max_participants, created_at, expires_at, closed_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
		ON CONFLICT (id) DO UPDATE SET
			title = EXCLUDED.title,
			owner_id = EXCLUDED.owner_id,
			status = EXCLUDED.status,
			join_policy = EXCLUDED.join_policy,
			max_participants = EXCLUDED.max_participants,
			expires_at = EXCLUDED.expires_at,
			closed_at = EXCLUDED.closed_at
	`, room.ID.UUID(), room.Title, owner, string(room.Status), string(room.JoinPolicy),
		room.MaxParticipants, room.CreatedAt, expires, closed)
	return err
}

func (r *RoomRepo) Delete(ctx context.Context, id domain.RoomID) error {
	_, err := r.db.Exec(ctx, `DELETE FROM rooms WHERE id = $1`, id.UUID())
	return err
}
