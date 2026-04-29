package rooms

import (
	"context"
	"time"
	"voco/internal/repository/liveKit"

	"voco/internal/domain"
)

type RoomStore interface {
	Get(ctx context.Context, id domain.RoomID) (domain.Room, bool, error)
	Upsert(ctx context.Context, room domain.Room, ttl time.Duration) error
	Delete(ctx context.Context, id domain.RoomID) error
}

type RoomUsecase struct {
	store      RoomStore
	LiveKitCfg liveKit.Cfg
}

func New(store RoomStore, liveKitCfg liveKit.Cfg) *RoomUsecase {
	return &RoomUsecase{
		store:      store,
		LiveKitCfg: liveKitCfg,
	}
}
