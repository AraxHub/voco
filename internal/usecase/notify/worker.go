package notify

import (
	"context"
	"log/slog"
	"time"

	"voco/internal/domain"
	"voco/internal/usecase/calendar"
)

type PushSender interface {
	Send(ctx context.Context, userID domain.UserID, title, body string) error
}

type Realtime interface {
	PublishToUsers(userIDs []domain.UserID, event string, payload any)
}

// Worker polls due calendar reminders and fans out toast + push.
type Worker struct {
	store calendar.Store
	rt    Realtime
	push  PushSender
	log   *slog.Logger
}

func NewWorker(store calendar.Store, rt Realtime, push PushSender, log *slog.Logger) *Worker {
	return &Worker{store: store, rt: rt, push: push, log: log}
}

func (w *Worker) Start(ctx context.Context, every time.Duration) {
	if every <= 0 {
		every = 30 * time.Second
	}
	go func() {
		t := time.NewTicker(every)
		defer t.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				w.tick(ctx)
			}
		}
	}()
}

func (w *Worker) tick(ctx context.Context) {
	due, err := w.store.DueReminders(ctx, time.Now().UTC(), 100)
	if err != nil {
		if w.log != nil {
			w.log.Error("reminders", "error", err)
		}
		return
	}
	for _, d := range due {
		if w.rt != nil {
			w.rt.PublishToUsers([]domain.UserID{d.UserID}, "notification", map[string]any{
				"type": "reminder", "title": d.Title, "eventId": d.EventID, "fireAt": d.FireAt,
			})
		}
		if w.push != nil {
			_ = w.push.Send(ctx, d.UserID, "Напоминание", d.Title)
		}
		_ = w.store.MarkReminderSent(ctx, d.ID, time.Now().UTC())
	}
}
