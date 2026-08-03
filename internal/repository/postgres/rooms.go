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

const (
	RoomTable = "rooms"

	RoomColID              = "id"
	RoomColTitle           = "title"
	RoomColOwnerID         = "owner_id"
	RoomColStatus          = "status"
	RoomColJoinPolicy      = "join_policy"
	RoomColMaxParticipants = "max_participants"
	RoomColCreatedAt       = "created_at"
	RoomColExpiresAt       = "expires_at"
	RoomColClosedAt        = "closed_at"
)

func RoomColumns() []string {
	return []string{
		RoomColID,
		RoomColTitle,
		RoomColOwnerID,
		RoomColStatus,
		RoomColJoinPolicy,
		RoomColMaxParticipants,
		RoomColCreatedAt,
		RoomColExpiresAt,
		RoomColClosedAt,
	}
}

func RoomSelect(alias string) string {
	return selectList(alias, RoomColumns())
}

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
	err := r.db.QueryRow(ctx,
		"SELECT "+RoomSelect("")+" FROM "+RoomTable+" WHERE "+RoomColID+" = $1",
		id.UUID(),
	).Scan(
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
	cols := RoomColumns()
	_, err := r.db.Exec(ctx,
		"INSERT INTO "+RoomTable+" ("+RoomSelect("")+") VALUES ("+placeholders(len(cols))+")"+
			" ON CONFLICT ("+RoomColID+") DO UPDATE SET "+
			RoomColTitle+" = EXCLUDED."+RoomColTitle+", "+
			RoomColOwnerID+" = EXCLUDED."+RoomColOwnerID+", "+
			RoomColStatus+" = EXCLUDED."+RoomColStatus+", "+
			RoomColJoinPolicy+" = EXCLUDED."+RoomColJoinPolicy+", "+
			RoomColMaxParticipants+" = EXCLUDED."+RoomColMaxParticipants+", "+
			RoomColExpiresAt+" = EXCLUDED."+RoomColExpiresAt+", "+
			RoomColClosedAt+" = EXCLUDED."+RoomColClosedAt,
		room.ID.UUID(), room.Title, owner, string(room.Status), string(room.JoinPolicy),
		room.MaxParticipants, room.CreatedAt, expires, closed)
	return err
}

func (r *RoomRepo) Delete(ctx context.Context, id domain.RoomID) error {
	_, err := r.db.Exec(ctx, "DELETE FROM "+RoomTable+" WHERE "+RoomColID+" = $1", id.UUID())
	return err
}
