package rooms

import (
	"context"
	"time"

	"voco/internal/domain"

	"github.com/google/uuid"
	lkauth "github.com/livekit/protocol/auth"
)

const defaultMaxParticipants = 10

func (uc *RoomUsecase) CreateRoom(ctx context.Context, title string) (domain.Room, error) {
	now := time.Now().UTC()
	ttl := 24 * time.Hour

	room := domain.NewOpenRoomByLink(title, now, defaultMaxParticipants)
	room.ExpiresAt = now.Add(ttl)

	if err := uc.store.Upsert(ctx, room, ttl); err != nil {
		return domain.Room{}, err
	}
	return room, nil
}

func (uc *RoomUsecase) IssueToken(ctx context.Context, roomID domain.RoomID, participantName string) (string, string, error) {
	room, ok, err := uc.store.Get(ctx, roomID)
	if err != nil {
		return "", "", err
	}
	if !ok || room.Status == domain.RoomStatusClosed {
		return "", "", domain.ErrRoomNotFound
	}

	now := time.Now().UTC()
	ttl := 24 * time.Hour
	room.ExpiresAt = now.Add(ttl)
	if err := uc.store.Upsert(ctx, room, ttl); err != nil {
		return "", "", err
	}

	identity := uuid.NewString()
	name := participantName
	if name == "" {
		name = "guest"
	}

	at := lkauth.NewAccessToken(uc.LiveKitCfg.LiveKitAPIKey, uc.LiveKitCfg.LiveKitAPISecret).
		SetIdentity(identity).
		SetName(name).
		SetValidFor(uc.LiveKitCfg.TokenTTL)
	at.AddGrant(&lkauth.VideoGrant{
		RoomJoin: true,
		Room:     roomID.String(),
	})

	jwt, err := at.ToJWT()
	if err != nil {
		return "", "", err
	}
	return jwt, uc.LiveKitCfg.LiveKitURL, nil
}
