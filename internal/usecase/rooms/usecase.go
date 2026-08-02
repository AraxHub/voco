package rooms

import (
	"context"
	"time"

	"voco/internal/adapters/livekit"
	"voco/internal/domain"

	"github.com/google/uuid"
	lkauth "github.com/livekit/protocol/auth"
)

const defaultMaxParticipants = 10

type Store interface {
	Get(ctx context.Context, id domain.RoomID) (domain.Room, bool, error)
	Upsert(ctx context.Context, room domain.Room, ttl time.Duration) error
	Delete(ctx context.Context, id domain.RoomID) error
}

type RoomUsecase struct {
	store      Store
	LiveKitCfg livekit.Cfg
	TTL        time.Duration
}

func New(store Store, cfg livekit.Cfg) *RoomUsecase {
	return &RoomUsecase{store: store, LiveKitCfg: cfg, TTL: 24 * time.Hour}
}

func (uc *RoomUsecase) CreateRoom(ctx context.Context, title string, owner *domain.UserID) (domain.Room, error) {
	now := time.Now().UTC()
	ttl := uc.TTL
	if ttl <= 0 {
		ttl = 24 * time.Hour
	}

	room := domain.NewOpenRoomByLink(title, now, defaultMaxParticipants)
	room.ExpiresAt = now.Add(ttl)
	room.Owner = owner

	if err := uc.store.Upsert(ctx, room, ttl); err != nil {
		return domain.Room{}, err
	}
	return room, nil
}

func (uc *RoomUsecase) IssueToken(ctx context.Context, roomID domain.RoomID, participantName string, identity string) (string, string, error) {
	room, ok, err := uc.store.Get(ctx, roomID)
	if err != nil {
		return "", "", err
	}
	if !ok || room.Status == domain.RoomStatusClosed {
		return "", "", domain.ErrRoomNotFound
	}

	now := time.Now().UTC()
	ttl := uc.TTL
	if ttl <= 0 {
		ttl = 24 * time.Hour
	}
	room.ExpiresAt = now.Add(ttl)
	if err := uc.store.Upsert(ctx, room, ttl); err != nil {
		return "", "", err
	}

	if identity == "" {
		identity = uuid.NewString()
	}
	name := participantName
	if name == "" {
		name = "guest"
	}

	canPublish := true
	canSubscribe := true
	canPublishData := true

	at := lkauth.NewAccessToken(uc.LiveKitCfg.LiveKitAPIKey, uc.LiveKitCfg.LiveKitAPISecret).
		SetIdentity(identity).
		SetName(name).
		SetValidFor(uc.LiveKitCfg.TokenTTL)
	at.AddGrant(&lkauth.VideoGrant{
		RoomJoin:       true,
		Room:           roomID.String(),
		CanPublish:     &canPublish,
		CanSubscribe:   &canSubscribe,
		CanPublishData: &canPublishData,
	})

	jwt, err := at.ToJWT()
	if err != nil {
		return "", "", err
	}
	return jwt, uc.LiveKitCfg.LiveKitURL, nil
}
