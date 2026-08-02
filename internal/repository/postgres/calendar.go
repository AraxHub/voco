package postgres

import (
	"context"
	"errors"
	"time"

	pgadapter "voco/internal/adapters/postgres"
	"voco/internal/domain"
	"voco/internal/usecase/calendar"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type CalendarRepo struct {
	db *pgadapter.Client
}

func NewCalendarRepo(db *pgadapter.Client) *CalendarRepo {
	return &CalendarRepo{db: db}
}

func (r *CalendarRepo) InsertEvent(ctx context.Context, e domain.CalendarEvent) error {
	tx, err := r.db.Pool().Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	var roomID any
	if e.RoomID != nil {
		roomID = e.RoomID.UUID()
	}
	var rrule any
	if e.RRule != "" {
		rrule = e.RRule
	}
	_, err = tx.Exec(ctx, `
		INSERT INTO calendar_events (id, organizer_id, title, description, starts_at, ends_at, timezone, status, room_id, rrule, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`,
		e.ID, e.OrganizerID, e.Title, e.Description, e.StartsAt, e.EndsAt, e.Timezone, string(e.Status),
		roomID, rrule, e.CreatedAt, e.UpdatedAt)
	if err != nil {
		return err
	}
	for _, a := range e.Attendees {
		_, err = tx.Exec(ctx, `INSERT INTO event_attendees (event_id, user_id, status) VALUES ($1,$2,$3)`,
			e.ID, a.UserID, string(a.Status))
		if err != nil {
			return err
		}
	}
	for _, m := range e.Reminders {
		_, err = tx.Exec(ctx, `INSERT INTO event_reminders (id, event_id, remind_before_minutes) VALUES ($1,$2,$3)`,
			uuid.New(), e.ID, m)
		if err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func (r *CalendarRepo) UpdateEvent(ctx context.Context, e domain.CalendarEvent) error {
	var roomID any
	if e.RoomID != nil {
		roomID = e.RoomID.UUID()
	}
	var rrule any
	if e.RRule != "" {
		rrule = e.RRule
	}
	tag, err := r.db.Exec(ctx, `
		UPDATE calendar_events SET title=$2, description=$3, starts_at=$4, ends_at=$5, timezone=$6,
			status=$7, room_id=$8, rrule=$9, updated_at=$10 WHERE id=$1`,
		e.ID, e.Title, e.Description, e.StartsAt, e.EndsAt, e.Timezone, string(e.Status), roomID, rrule, e.UpdatedAt)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrEventNotFound
	}
	return nil
}

func (r *CalendarRepo) GetEvent(ctx context.Context, id domain.EventID) (domain.CalendarEvent, error) {
	var e domain.CalendarEvent
	var status string
	var roomID *uuid.UUID
	var rrule *string
	err := r.db.QueryRow(ctx, `
		SELECT id, organizer_id, title, description, starts_at, ends_at, timezone, status, room_id, rrule, created_at, updated_at
		FROM calendar_events WHERE id=$1`, id).
		Scan(&e.ID, &e.OrganizerID, &e.Title, &e.Description, &e.StartsAt, &e.EndsAt, &e.Timezone, &status, &roomID, &rrule, &e.CreatedAt, &e.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.CalendarEvent{}, domain.ErrEventNotFound
	}
	if err != nil {
		return domain.CalendarEvent{}, err
	}
	e.Status = domain.EventStatus(status)
	if roomID != nil {
		rid := domain.RoomID(*roomID)
		e.RoomID = &rid
	}
	if rrule != nil {
		e.RRule = *rrule
	}
	e.Attendees, _ = r.loadAttendees(ctx, id)
	e.Reminders, _ = r.loadRemindMinutes(ctx, id)
	return e, nil
}

func (r *CalendarRepo) loadAttendees(ctx context.Context, id domain.EventID) ([]domain.EventAttendee, error) {
	rows, err := r.db.Query(ctx, `SELECT event_id, user_id, status FROM event_attendees WHERE event_id=$1`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.EventAttendee
	for rows.Next() {
		var a domain.EventAttendee
		var st string
		if err := rows.Scan(&a.EventID, &a.UserID, &st); err != nil {
			return nil, err
		}
		a.Status = domain.RSVPStatus(st)
		out = append(out, a)
	}
	return out, rows.Err()
}

func (r *CalendarRepo) loadRemindMinutes(ctx context.Context, id domain.EventID) ([]int, error) {
	rows, err := r.db.Query(ctx, `SELECT remind_before_minutes FROM event_reminders WHERE event_id=$1`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []int
	for rows.Next() {
		var m int
		if err := rows.Scan(&m); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

func (r *CalendarRepo) ListEventsForUser(ctx context.Context, userID domain.UserID, from, to time.Time) ([]domain.CalendarEvent, error) {
	rows, err := r.db.Query(ctx, `
		SELECT DISTINCT e.id FROM calendar_events e
		LEFT JOIN event_attendees a ON a.event_id = e.id
		WHERE (e.organizer_id = $1 OR a.user_id = $1)
		  AND e.starts_at < $3 AND e.ends_at > $2`, userID, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ids []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	out := make([]domain.CalendarEvent, 0, len(ids))
	for _, id := range ids {
		e, err := r.GetEvent(ctx, id)
		if err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, nil
}

func (r *CalendarRepo) SetRSVP(ctx context.Context, eventID domain.EventID, userID domain.UserID, status domain.RSVPStatus) error {
	tag, err := r.db.Exec(ctx, `UPDATE event_attendees SET status=$3 WHERE event_id=$1 AND user_id=$2`, eventID, userID, string(status))
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrForbidden
	}
	return nil
}

func (r *CalendarRepo) ListBusy(ctx context.Context, userIDs []domain.UserID, from, to time.Time) ([]domain.BusyInterval, error) {
	rows, err := r.db.Query(ctx, `
		SELECT e.starts_at, e.ends_at
		FROM calendar_events e
		LEFT JOIN event_attendees a ON a.event_id = e.id
		WHERE e.status = 'scheduled'
		  AND e.starts_at < $2 AND e.ends_at > $1
		  AND (e.organizer_id = ANY($3) OR (a.user_id = ANY($3) AND a.status <> 'declined'))`,
		from, to, userIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.BusyInterval
	for rows.Next() {
		var b domain.BusyInterval
		if err := rows.Scan(&b.Start, &b.End); err != nil {
			return nil, err
		}
		out = append(out, b)
	}
	return out, rows.Err()
}

func (r *CalendarRepo) EnqueueReminders(ctx context.Context, eventID domain.EventID, userIDs []domain.UserID, startsAt time.Time, minutes []int) error {
	for _, m := range minutes {
		var remID uuid.UUID
		err := r.db.QueryRow(ctx, `
			INSERT INTO event_reminders (id, event_id, remind_before_minutes) VALUES ($1,$2,$3)
			ON CONFLICT (event_id, remind_before_minutes) DO UPDATE SET remind_before_minutes = EXCLUDED.remind_before_minutes
			RETURNING id`, uuid.New(), eventID, m).Scan(&remID)
		if err != nil {
			// try select existing
			_ = r.db.QueryRow(ctx, `SELECT id FROM event_reminders WHERE event_id=$1 AND remind_before_minutes=$2`, eventID, m).Scan(&remID)
		}
		fire := startsAt.Add(-time.Duration(m) * time.Minute)
		for _, uid := range userIDs {
			_, _ = r.db.Exec(ctx, `
				INSERT INTO reminder_deliveries (id, reminder_id, user_id, fire_at)
				VALUES ($1,$2,$3,$4)
				ON CONFLICT DO NOTHING`, uuid.New(), remID, uid, fire)
		}
	}
	return nil
}

func (r *CalendarRepo) DueReminders(ctx context.Context, now time.Time, limit int) ([]calendar.ReminderDue, error) {
	rows, err := r.db.Query(ctx, `
		SELECT d.id, r.event_id, d.user_id, d.fire_at, e.title
		FROM reminder_deliveries d
		JOIN event_reminders r ON r.id = d.reminder_id
		JOIN calendar_events e ON e.id = r.event_id
		WHERE d.sent_at IS NULL AND d.fire_at <= $1
		ORDER BY d.fire_at
		LIMIT $2`, now, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []calendar.ReminderDue
	for rows.Next() {
		var d calendar.ReminderDue
		if err := rows.Scan(&d.ID, &d.EventID, &d.UserID, &d.FireAt, &d.Title); err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, rows.Err()
}

func (r *CalendarRepo) MarkReminderSent(ctx context.Context, id uuid.UUID, at time.Time) error {
	_, err := r.db.Exec(ctx, `UPDATE reminder_deliveries SET sent_at=$2 WHERE id=$1`, id, at)
	return err
}
