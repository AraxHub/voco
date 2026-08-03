package calendar

import (
	"context"
	"fmt"
	"strings"
	"time"

	"voco/internal/domain"

	"github.com/google/uuid"
)

type RoomCreator interface {
	CreateRoom(ctx context.Context, title string, owner *domain.UserID) (domain.Room, error)
}

type Realtime interface {
	PublishToUsers(userIDs []domain.UserID, event string, payload any)
}

type Store interface {
	InsertEvent(ctx context.Context, e domain.CalendarEvent) error
	UpdateEvent(ctx context.Context, e domain.CalendarEvent) error
	GetEvent(ctx context.Context, id domain.EventID) (domain.CalendarEvent, error)
	ListEventsForUser(ctx context.Context, userID domain.UserID, from, to time.Time) ([]domain.CalendarEvent, error)
	SetRSVP(ctx context.Context, eventID domain.EventID, userID domain.UserID, status domain.RSVPStatus) error
	ListBusy(ctx context.Context, userIDs []domain.UserID, from, to time.Time) ([]domain.BusyInterval, error)
	EnqueueReminders(ctx context.Context, eventID domain.EventID, userIDs []domain.UserID, startsAt time.Time, minutes []int) error
	DueReminders(ctx context.Context, now time.Time, limit int) ([]ReminderDue, error)
	MarkReminderSent(ctx context.Context, id uuid.UUID, at time.Time) error
}

type ReminderDue struct {
	ID      uuid.UUID
	EventID domain.EventID
	UserID  domain.UserID
	FireAt  time.Time
	Title   string
}

type Usecase struct {
	store Store
	rooms RoomCreator
	rt    Realtime
}

func New(store Store, rooms RoomCreator, rt Realtime) *Usecase {
	return &Usecase{store: store, rooms: rooms, rt: rt}
}

func (uc *Usecase) Create(ctx context.Context, organizer domain.UserID, title, description, tz, rrule string, starts, ends time.Time, attendees []domain.UserID, remind []int) (domain.CalendarEvent, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		return domain.CalendarEvent{}, fmt.Errorf("%w: укажите название встречи", domain.ErrValidation)
	}
	if !ends.After(starts) {
		return domain.CalendarEvent{}, fmt.Errorf("%w: время окончания должно быть позже начала", domain.ErrValidation)
	}
	if tz == "" {
		tz = "UTC"
	}
	if rrule != "" && !validRRule(rrule) {
		return domain.CalendarEvent{}, domain.ErrValidation
	}
	owner := organizer
	room, err := uc.rooms.CreateRoom(ctx, title, &owner)
	if err != nil {
		return domain.CalendarEvent{}, err
	}
	rid := room.ID
	now := time.Now().UTC()
	ev := domain.CalendarEvent{
		ID: uuid.New(), OrganizerID: organizer, Title: title, Description: description,
		StartsAt: starts.UTC(), EndsAt: ends.UTC(), Timezone: tz, Status: domain.EventScheduled,
		RoomID: &rid, RRule: rrule, CreatedAt: now, UpdatedAt: now, Reminders: remind,
	}
	seen := map[domain.UserID]struct{}{organizer: {}}
	ev.Attendees = append(ev.Attendees, domain.EventAttendee{EventID: ev.ID, UserID: organizer, Status: domain.RSVPAccepted})
	for _, id := range attendees {
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		ev.Attendees = append(ev.Attendees, domain.EventAttendee{EventID: ev.ID, UserID: id, Status: domain.RSVPPending})
	}
	if err := uc.store.InsertEvent(ctx, ev); err != nil {
		return domain.CalendarEvent{}, err
	}
	uids := make([]domain.UserID, 0, len(seen))
	for id := range seen {
		uids = append(uids, id)
	}
	_ = uc.store.EnqueueReminders(ctx, ev.ID, uids, ev.StartsAt, remind)
	if uc.rt != nil {
		uc.rt.PublishToUsers(uids, "calendar.updated", ev)
	}
	return ev, nil
}

func validRRule(r string) bool {
	r = strings.ToUpper(r)
	return strings.Contains(r, "FREQ=DAILY") || strings.Contains(r, "FREQ=WEEKLY") || strings.Contains(r, "FREQ=MONTHLY")
}

func (uc *Usecase) Cancel(ctx context.Context, me domain.UserID, id domain.EventID) (domain.CalendarEvent, error) {
	ev, err := uc.store.GetEvent(ctx, id)
	if err != nil {
		return domain.CalendarEvent{}, err
	}
	if ev.OrganizerID != me {
		return domain.CalendarEvent{}, domain.ErrForbidden
	}
	ev.Status = domain.EventCancelled
	ev.UpdatedAt = time.Now().UTC()
	if err := uc.store.UpdateEvent(ctx, ev); err != nil {
		return domain.CalendarEvent{}, err
	}
	uc.publishAttendees(ev, "calendar.updated", ev)
	return ev, nil
}

func (uc *Usecase) Reschedule(ctx context.Context, me domain.UserID, id domain.EventID, starts, ends time.Time) (domain.CalendarEvent, error) {
	if !ends.After(starts) {
		return domain.CalendarEvent{}, domain.ErrValidation
	}
	ev, err := uc.store.GetEvent(ctx, id)
	if err != nil {
		return domain.CalendarEvent{}, err
	}
	if ev.OrganizerID != me {
		return domain.CalendarEvent{}, domain.ErrForbidden
	}
	if ev.Status == domain.EventCancelled {
		return domain.CalendarEvent{}, fmt.Errorf("%w: отменённую встречу нельзя переносить", domain.ErrValidation)
	}
	roomID := ev.RoomID
	ev.StartsAt = starts.UTC()
	ev.EndsAt = ends.UTC()
	ev.UpdatedAt = time.Now().UTC()
	ev.RoomID = roomID
	if err := uc.store.UpdateEvent(ctx, ev); err != nil {
		return domain.CalendarEvent{}, err
	}
	uids := attendeeIDs(ev)
	_ = uc.store.EnqueueReminders(ctx, ev.ID, uids, ev.StartsAt, ev.Reminders)
	uc.publishAttendees(ev, "calendar.updated", ev)
	return ev, nil
}

func (uc *Usecase) RSVP(ctx context.Context, me domain.UserID, id domain.EventID, status domain.RSVPStatus) error {
	switch status {
	case domain.RSVPAccepted, domain.RSVPTentative, domain.RSVPDeclined, domain.RSVPPending:
	default:
		return domain.ErrValidation
	}
	ev, err := uc.store.GetEvent(ctx, id)
	if err != nil {
		return err
	}
	ok := false
	for _, a := range ev.Attendees {
		if a.UserID == me {
			ok = true
			break
		}
	}
	if !ok {
		return domain.ErrForbidden
	}
	if err := uc.store.SetRSVP(ctx, id, me, status); err != nil {
		return err
	}
	uc.publishAttendees(ev, "calendar.updated", map[string]any{"id": id, "userId": me, "status": status})
	return nil
}

func (uc *Usecase) List(ctx context.Context, me domain.UserID, from, to time.Time) ([]domain.CalendarEvent, error) {
	events, err := uc.store.ListEventsForUser(ctx, me, from, to)
	if err != nil {
		return nil, err
	}
	var out []domain.CalendarEvent
	for _, ev := range events {
		out = append(out, ExpandOccurrences(ev, from, to)...)
	}
	return out, nil
}

func (uc *Usecase) FreeBusy(ctx context.Context, userIDs []domain.UserID, from, to time.Time) ([]domain.BusyInterval, error) {
	raw, err := uc.store.ListBusy(ctx, userIDs, from, to)
	if err != nil {
		return nil, err
	}
	return QuantizeBusy(raw, 30*time.Minute), nil
}

func (uc *Usecase) publishAttendees(ev domain.CalendarEvent, event string, payload any) {
	if uc.rt == nil {
		return
	}
	uc.rt.PublishToUsers(attendeeIDs(ev), event, payload)
}

func attendeeIDs(ev domain.CalendarEvent) []domain.UserID {
	ids := make([]domain.UserID, 0, len(ev.Attendees)+1)
	seen := map[domain.UserID]struct{}{}
	add := func(id domain.UserID) {
		if _, ok := seen[id]; ok {
			return
		}
		seen[id] = struct{}{}
		ids = append(ids, id)
	}
	add(ev.OrganizerID)
	for _, a := range ev.Attendees {
		add(a.UserID)
	}
	return ids
}

// ExpandOccurrences expands simple DAILY/WEEKLY/MONTHLY rrules into concrete instances in [from,to].
func ExpandOccurrences(ev domain.CalendarEvent, from, to time.Time) []domain.CalendarEvent {
	if ev.RRule == "" {
		if ev.EndsAt.Before(from) || ev.StartsAt.After(to) {
			return nil
		}
		return []domain.CalendarEvent{ev}
	}
	freq := parseFreq(ev.RRule)
	count := parseCount(ev.RRule)
	until := parseUntil(ev.RRule)
	dur := ev.EndsAt.Sub(ev.StartsAt)
	var out []domain.CalendarEvent
	cur := ev.StartsAt
	n := 0
	for !cur.After(to) {
		end := cur.Add(dur)
		if !end.Before(from) && !cur.After(to) {
			inst := ev
			inst.StartsAt = cur
			inst.EndsAt = end
			out = append(out, inst)
		}
		n++
		if count > 0 && n >= count {
			break
		}
		if until != nil && cur.After(*until) {
			break
		}
		switch freq {
		case "DAILY":
			cur = cur.AddDate(0, 0, 1)
		case "WEEKLY":
			cur = cur.AddDate(0, 0, 7)
		case "MONTHLY":
			cur = cur.AddDate(0, 1, 0)
		default:
			return out
		}
		if n > 366 {
			break
		}
	}
	return out
}

func parseFreq(r string) string {
	u := strings.ToUpper(r)
	for _, f := range []string{"DAILY", "WEEKLY", "MONTHLY"} {
		if strings.Contains(u, "FREQ="+f) {
			return f
		}
	}
	return ""
}

func parseCount(r string) int {
	u := strings.ToUpper(r)
	i := strings.Index(u, "COUNT=")
	if i < 0 {
		return 0
	}
	n := 0
	for _, c := range u[i+6:] {
		if c < '0' || c > '9' {
			break
		}
		n = n*10 + int(c-'0')
	}
	return n
}

func parseUntil(r string) *time.Time {
	u := strings.ToUpper(r)
	i := strings.Index(u, "UNTIL=")
	if i < 0 {
		return nil
	}
	raw := strings.SplitN(u[i+6:], ";", 2)[0]
	t, err := time.Parse("20060102T150405Z", raw)
	if err != nil {
		t, err = time.Parse("20060102", raw)
		if err != nil {
			return nil
		}
	}
	return &t
}

// QuantizeBusy merges and snaps busy intervals to slotSize grid.
func QuantizeBusy(in []domain.BusyInterval, slotSize time.Duration) []domain.BusyInterval {
	if slotSize <= 0 {
		slotSize = 30 * time.Minute
	}
	var out []domain.BusyInterval
	for _, b := range in {
		start := b.Start.UTC().Truncate(slotSize)
		end := b.End.UTC()
		if !end.Equal(end.Truncate(slotSize)) {
			end = end.Truncate(slotSize).Add(slotSize)
		}
		if !end.After(start) {
			end = start.Add(slotSize)
		}
		out = append(out, domain.BusyInterval{Start: start, End: end})
	}
	// merge overlaps
	if len(out) == 0 {
		return out
	}
	for i := 0; i < len(out); i++ {
		for j := i + 1; j < len(out); j++ {
			if out[j].Start.Before(out[i].Start) {
				out[i], out[j] = out[j], out[i]
			}
		}
	}
	merged := []domain.BusyInterval{out[0]}
	for _, b := range out[1:] {
		last := &merged[len(merged)-1]
		if !b.Start.After(last.End) {
			if b.End.After(last.End) {
				last.End = b.End
			}
			continue
		}
		merged = append(merged, b)
	}
	return merged
}
