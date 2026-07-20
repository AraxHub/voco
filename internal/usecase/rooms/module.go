package rooms

import (
	"voco/internal/adapters/livekit"
)

type RoomUsecase struct {
	store      Store
	LiveKitCfg livekit.Cfg
}

func New(store Store, liveKitCfg livekit.Cfg) *RoomUsecase {
	return &RoomUsecase{
		store:      store,
		LiveKitCfg: liveKitCfg,
	}
}
