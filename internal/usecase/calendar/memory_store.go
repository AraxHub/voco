package calendar

import (
	"context"
	"sync"
	"time"

	"voco/internal/domain"

	"github.com/google/uuid"
)

type MemoryStore struct {
	mu        sync.RWMutex
	events    map[uuid.UUID]domain.CalendarEvent
	reminders []ReminderDue
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{events: map[uuid.UUID]domain.CalendarEvent{}}
}

func (s *MemoryStore) InsertEvent(_ context.Context, e domain.CalendarEvent) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.events[e.ID] = e
	return nil
}

func (s *MemoryStore) UpdateEvent(_ context.Context, e domain.CalendarEvent) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.events[e.ID]; !ok {
		return domain.ErrEventNotFound
	}
	s.events[e.ID] = e
	return nil
}

func (s *MemoryStore) GetEvent(_ context.Context, id domain.EventID) (domain.CalendarEvent, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	e, ok := s.events[id]
	if !ok {
		return domain.CalendarEvent{}, domain.ErrEventNotFound
	}
	return e, nil
}

func (s *MemoryStore) ListEventsForUser(_ context.Context, userID domain.UserID, from, to time.Time) ([]domain.CalendarEvent, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []domain.CalendarEvent
	for _, e := range s.events {
		if e.OrganizerID == userID {
			out = append(out, e)
			continue
		}
		for _, a := range e.Attendees {
			if a.UserID == userID {
				out = append(out, e)
				break
			}
		}
	}
	_ = from
	_ = to
	return out, nil
}

func (s *MemoryStore) SetRSVP(_ context.Context, eventID domain.EventID, userID domain.UserID, status domain.RSVPStatus) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	e, ok := s.events[eventID]
	if !ok {
		return domain.ErrEventNotFound
	}
	for i := range e.Attendees {
		if e.Attendees[i].UserID == userID {
			e.Attendees[i].Status = status
			s.events[eventID] = e
			return nil
		}
	}
	return domain.ErrForbidden
}

func (s *MemoryStore) ListBusy(_ context.Context, userIDs []domain.UserID, from, to time.Time) ([]domain.BusyInterval, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	set := map[uuid.UUID]struct{}{}
	for _, id := range userIDs {
		set[id] = struct{}{}
	}
	var out []domain.BusyInterval
	for _, e := range s.events {
		if e.Status == domain.EventCancelled {
			continue
		}
		involved := false
		if _, ok := set[e.OrganizerID]; ok {
			involved = true
		}
		for _, a := range e.Attendees {
			if _, ok := set[a.UserID]; ok && a.Status != domain.RSVPDeclined {
				involved = true
			}
		}
		if !involved {
			continue
		}
		if e.EndsAt.Before(from) || e.StartsAt.After(to) {
			continue
		}
		out = append(out, domain.BusyInterval{Start: e.StartsAt, End: e.EndsAt})
	}
	return out, nil
}

func (s *MemoryStore) EnqueueReminders(_ context.Context, eventID domain.EventID, userIDs []domain.UserID, startsAt time.Time, minutes []int) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	title := ""
	if e, ok := s.events[eventID]; ok {
		title = e.Title
	}
	for _, m := range minutes {
		fire := startsAt.Add(-time.Duration(m) * time.Minute)
		for _, uid := range userIDs {
			s.reminders = append(s.reminders, ReminderDue{
				ID: uuid.New(), EventID: eventID, UserID: uid, FireAt: fire, Title: title,
			})
		}
	}
	return nil
}

func (s *MemoryStore) DueReminders(_ context.Context, now time.Time, limit int) ([]ReminderDue, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []ReminderDue
	for _, r := range s.reminders {
		if !r.FireAt.After(now) {
			out = append(out, r)
			if len(out) >= limit {
				break
			}
		}
	}
	return out, nil
}

func (s *MemoryStore) MarkReminderSent(_ context.Context, id uuid.UUID, _ time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := s.reminders[:0]
	for _, r := range s.reminders {
		if r.ID == id {
			continue
		}
		out = append(out, r)
	}
	s.reminders = out
	return nil
}
