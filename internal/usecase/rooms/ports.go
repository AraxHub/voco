package rooms

import (
	"context"
	"time"

	"voco/internal/domain"
)

type Store interface {
	Get(ctx context.Context, id domain.RoomID) (domain.Room, bool, error)
	Upsert(ctx context.Context, room domain.Room, ttl time.Duration) error
	Delete(ctx context.Context, id domain.RoomID) error
}
