package rooms_test

import (
	"context"
	"testing"
	"time"

	"voco/internal/adapters/livekit"
	"voco/internal/domain"
	"voco/internal/usecase/rooms"

	"github.com/google/uuid"
)

type mem struct {
	m map[domain.RoomID]domain.Room
}

func (s *mem) Get(_ context.Context, id domain.RoomID) (domain.Room, bool, error) {
	r, ok := s.m[id]
	return r, ok, nil
}
func (s *mem) Upsert(_ context.Context, room domain.Room, _ time.Duration) error {
	if s.m == nil {
		s.m = map[domain.RoomID]domain.Room{}
	}
	s.m[room.ID] = room
	return nil
}
func (s *mem) Delete(_ context.Context, id domain.RoomID) error {
	delete(s.m, id)
	return nil
}

func TestCreateRoomPersistsOwner(t *testing.T) {
	uc := rooms.New(&mem{}, livekit.Cfg{
		LiveKitAPIKey: "k", LiveKitAPISecret: "s", LiveKitURL: "ws://x", TokenTTL: time.Hour,
	})
	owner := uuid.New()
	room, err := uc.CreateRoom(context.Background(), "t", &owner)
	if err != nil || room.Owner == nil || *room.Owner != owner {
		t.Fatalf("%v %+v", err, room)
	}
	tok, url, err := uc.IssueToken(context.Background(), room.ID, "n", owner.String())
	if err != nil || tok == "" || url == "" {
		t.Fatalf("%v", err)
	}
}
