package chat_test

import (
	"context"
	"errors"
	"testing"

	blobmem "voco/internal/adapters/blob/memory"
	"voco/internal/domain"
	"voco/internal/usecase/chat"

	"github.com/google/uuid"
)

type noopRT struct{}

func (noopRT) PublishToUsers([]domain.UserID, string, any) {}

type roomStub struct{ n int }

func (r *roomStub) CreateRoom(context.Context, string, *domain.UserID) (domain.Room, error) {
	r.n++
	return domain.Room{ID: domain.NewRoomID(), Title: "Call"}, nil
}

func (r *roomStub) CloseRoom(context.Context, domain.RoomID) error { return nil }

type userStub struct{}

func (userStub) GetByID(_ context.Context, id domain.UserID) (domain.User, error) {
	return domain.User{ID: id, Nickname: "nick", DisplayName: "Name"}, nil
}

func TestDirectAcceptBlockAndDeleteModes(t *testing.T) {
	ctx := context.Background()
	store := chat.NewMemoryStore()
	uc := chat.New(store, blobmem.NewBlobStore(), &roomStub{}, userStub{}, noopRT{}, chat.Config{})

	a, b := uuid.New(), uuid.New()
	c, req, err := uc.GetOrCreateDirect(ctx, a, b)
	if err != nil || req.Status != domain.MessageRequestPending {
		t.Fatalf("create: %v %+v", err, req)
	}
	c2, _, err := uc.GetOrCreateDirect(ctx, a, b)
	if err != nil || c2.ID != c.ID {
		t.Fatalf("idempotent dm")
	}

	msg, err := uc.SendMessage(ctx, a, c.ID, "hi", nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := uc.SendMessage(ctx, b, c.ID, "nope", nil); err != domain.ErrMessageRequestPending {
		t.Fatalf("recipient before accept: %v", err)
	}
	if err := uc.AcceptRequest(ctx, b, c.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := uc.SendMessage(ctx, b, c.ID, "yo", nil); err != nil {
		t.Fatal(err)
	}

	// new peer block path
	x, y := uuid.New(), uuid.New()
	c3, _, _ := uc.GetOrCreateDirect(ctx, x, y)
	_, _ = uc.SendMessage(ctx, x, c3.ID, "spam", nil)
	if err := uc.BlockRequest(ctx, y, c3.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := uc.SendMessage(ctx, x, c3.ID, "again", nil); err != domain.ErrBlocked {
		t.Fatalf("want blocked got %v", err)
	}

	if err := uc.DeleteMessage(ctx, a, msg.ID, domain.DeleteForMe); err != nil {
		t.Fatal(err)
	}
	list, _ := uc.ListMessages(ctx, a, c.ID, 50)
	for _, m := range list {
		if m.ID == msg.ID {
			t.Fatal("hidden for me")
		}
	}
}

func TestGroupTitleAdminLeaveReact(t *testing.T) {
	ctx := context.Background()
	store := chat.NewMemoryStore()
	uc := chat.New(store, blobmem.NewBlobStore(), &roomStub{}, userStub{}, noopRT{}, chat.Config{})
	admin, m1, m2 := uuid.New(), uuid.New(), uuid.New()

	if _, err := uc.CreateGroup(ctx, admin, "", []domain.UserID{m1}, nil); !errors.Is(err, domain.ErrValidation) {
		t.Fatal("title required")
	}
	g, err := uc.CreateGroup(ctx, admin, "Team", []domain.UserID{m1, m2}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := uc.PromoteAdmin(ctx, admin, m1, g.ID); err != nil {
		t.Fatal(err)
	}
	if err := uc.Leave(ctx, m2, g.ID); err != nil {
		t.Fatal(err)
	}
	msg, err := uc.SendMessage(ctx, admin, g.ID, "hello", nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := uc.React(ctx, m1, msg.ID, "🔥", true); err != nil {
		t.Fatal(err)
	}
	if err := uc.DeleteMessage(ctx, admin, msg.ID, domain.DeleteForAll); err != nil {
		t.Fatal(err)
	}
	got, _ := store.GetMessage(ctx, msg.ID)
	if got.DeletedForAllAt == nil {
		t.Fatal("soft delete")
	}
}
