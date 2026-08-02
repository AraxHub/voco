package calendar_test

import (
	"context"
	"testing"
	"time"

	"voco/internal/domain"
	"voco/internal/usecase/calendar"

	"github.com/google/uuid"
)

type roomStub struct{}

func (roomStub) CreateRoom(context.Context, string, *domain.UserID) (domain.Room, error) {
	id := domain.NewRoomID()
	return domain.Room{ID: id}, nil
}

func TestCreateCancelRSVPFreeBusyReschedule(t *testing.T) {
	ctx := context.Background()
	store := calendar.NewMemoryStore()
	uc := calendar.New(store, roomStub{}, nil)
	org, peer := uuid.New(), uuid.New()
	start := time.Now().UTC().Truncate(time.Hour).Add(2 * time.Hour)
	end := start.Add(time.Hour)

	ev, err := uc.Create(ctx, org, "Sync", "", "Europe/Moscow", "", start, end, []domain.UserID{peer}, []int{15})
	if err != nil || ev.RoomID == nil {
		t.Fatalf("create: %v %+v", err, ev)
	}
	roomKeep := *ev.RoomID

	if err := uc.RSVP(ctx, peer, ev.ID, domain.RSVPAccepted); err != nil {
		t.Fatal(err)
	}
	busy, err := uc.FreeBusy(ctx, []domain.UserID{peer}, start.Add(-time.Hour), end.Add(time.Hour))
	if err != nil || len(busy) == 0 {
		t.Fatalf("busy: %v %#v", err, busy)
	}
	for _, b := range busy {
		if b.Start.Minute()%30 != 0 {
			t.Fatalf("not 30m quantized: %v", b.Start)
		}
	}

	moved, err := uc.Reschedule(ctx, org, ev.ID, start.Add(time.Hour), end.Add(time.Hour))
	if err != nil || moved.RoomID == nil || *moved.RoomID != roomKeep {
		t.Fatalf("reschedule room: %v %+v", err, moved)
	}

	cancelled, err := uc.Cancel(ctx, org, ev.ID)
	if err != nil || cancelled.Status != domain.EventCancelled {
		t.Fatalf("cancel: %v", err)
	}
}

func TestRecurrenceExpandAndOverlapAllowed(t *testing.T) {
	ctx := context.Background()
	store := calendar.NewMemoryStore()
	uc := calendar.New(store, roomStub{}, nil)
	org := uuid.New()
	start := time.Date(2026, 8, 3, 10, 0, 0, 0, time.UTC)
	end := start.Add(30 * time.Minute)
	_, err := uc.Create(ctx, org, "A", "", "UTC", "FREQ=WEEKLY;COUNT=3", start, end, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	// overlap same time different event — allowed
	_, err = uc.Create(ctx, org, "B", "", "UTC", "", start, end, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	list, err := uc.List(ctx, org, start.Add(-time.Hour), start.AddDate(0, 0, 21))
	if err != nil {
		t.Fatal(err)
	}
	weekly := 0
	for _, e := range list {
		if e.Title == "A" {
			weekly++
		}
	}
	if weekly != 3 {
		t.Fatalf("want 3 weekly instances, got %d (total %d)", weekly, len(list))
	}
}

func TestQuantizeBusy(t *testing.T) {
	in := []domain.BusyInterval{{
		Start: time.Date(2026, 1, 1, 10, 5, 0, 0, time.UTC),
		End:   time.Date(2026, 1, 1, 10, 40, 0, 0, time.UTC),
	}}
	out := calendar.QuantizeBusy(in, 30*time.Minute)
	if len(out) != 1 || out[0].Start.Minute() != 0 || out[0].End.Minute() != 0 {
		t.Fatalf("%+v", out)
	}
}
