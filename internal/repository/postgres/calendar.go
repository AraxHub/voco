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

const (
	CalendarEventTable = "calendar_events"

	CalendarEventColID          = "id"
	CalendarEventColOrganizerID = "organizer_id"
	CalendarEventColTitle       = "title"
	CalendarEventColDescription = "description"
	CalendarEventColStartsAt    = "starts_at"
	CalendarEventColEndsAt      = "ends_at"
	CalendarEventColTimezone    = "timezone"
	CalendarEventColStatus      = "status"
	CalendarEventColRoomID      = "room_id"
	CalendarEventColRRule       = "rrule"
	CalendarEventColCreatedAt   = "created_at"
	CalendarEventColUpdatedAt   = "updated_at"
)

func CalendarEventColumns() []string {
	return []string{
		CalendarEventColID,
		CalendarEventColOrganizerID,
		CalendarEventColTitle,
		CalendarEventColDescription,
		CalendarEventColStartsAt,
		CalendarEventColEndsAt,
		CalendarEventColTimezone,
		CalendarEventColStatus,
		CalendarEventColRoomID,
		CalendarEventColRRule,
		CalendarEventColCreatedAt,
		CalendarEventColUpdatedAt,
	}
}

func CalendarEventSelect(alias string) string {
	return selectList(alias, CalendarEventColumns())
}

const (
	EventAttendeeTable = "event_attendees"

	EventAttendeeColEventID = "event_id"
	EventAttendeeColUserID  = "user_id"
	EventAttendeeColStatus  = "status"
)

func EventAttendeeColumns() []string {
	return []string{
		EventAttendeeColEventID,
		EventAttendeeColUserID,
		EventAttendeeColStatus,
	}
}

func EventAttendeeSelect(alias string) string {
	return selectList(alias, EventAttendeeColumns())
}

const (
	EventReminderTable = "event_reminders"

	EventReminderColID                  = "id"
	EventReminderColEventID             = "event_id"
	EventReminderColRemindBeforeMinutes = "remind_before_minutes"
)

func EventReminderColumns() []string {
	return []string{
		EventReminderColID,
		EventReminderColEventID,
		EventReminderColRemindBeforeMinutes,
	}
}

func EventReminderSelect(alias string) string {
	return selectList(alias, EventReminderColumns())
}

const (
	ReminderDeliveryTable = "reminder_deliveries"

	ReminderDeliveryColID         = "id"
	ReminderDeliveryColReminderID = "reminder_id"
	ReminderDeliveryColUserID     = "user_id"
	ReminderDeliveryColFireAt     = "fire_at"
	ReminderDeliveryColSentAt     = "sent_at"
)

func ReminderDeliveryColumns() []string {
	return []string{
		ReminderDeliveryColID,
		ReminderDeliveryColReminderID,
		ReminderDeliveryColUserID,
		ReminderDeliveryColFireAt,
		ReminderDeliveryColSentAt,
	}
}

func ReminderDeliverySelect(alias string) string {
	return selectList(alias, ReminderDeliveryColumns())
}

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
	cols := CalendarEventColumns()
	_, err = tx.Exec(ctx,
		"INSERT INTO "+CalendarEventTable+" ("+CalendarEventSelect("")+") VALUES ("+placeholders(len(cols))+")",
		e.ID, e.OrganizerID, e.Title, e.Description, e.StartsAt, e.EndsAt, e.Timezone, string(e.Status),
		roomID, rrule, e.CreatedAt, e.UpdatedAt)
	if err != nil {
		return err
	}
	attendeeCols := EventAttendeeColumns()
	for _, a := range e.Attendees {
		_, err = tx.Exec(ctx,
			"INSERT INTO "+EventAttendeeTable+" ("+EventAttendeeSelect("")+") VALUES ("+placeholders(len(attendeeCols))+")",
			e.ID, a.UserID, string(a.Status))
		if err != nil {
			return err
		}
	}
	reminderCols := EventReminderColumns()
	for _, m := range e.Reminders {
		_, err = tx.Exec(ctx,
			"INSERT INTO "+EventReminderTable+" ("+EventReminderSelect("")+") VALUES ("+placeholders(len(reminderCols))+")",
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
	tag, err := r.db.Exec(ctx,
		"UPDATE "+CalendarEventTable+" SET "+
			CalendarEventColTitle+" = $2, "+
			CalendarEventColDescription+" = $3, "+
			CalendarEventColStartsAt+" = $4, "+
			CalendarEventColEndsAt+" = $5, "+
			CalendarEventColTimezone+" = $6, "+
			CalendarEventColStatus+" = $7, "+
			CalendarEventColRoomID+" = $8, "+
			CalendarEventColRRule+" = $9, "+
			CalendarEventColUpdatedAt+" = $10"+
			" WHERE "+CalendarEventColID+" = $1",
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
	err := r.db.QueryRow(ctx,
		"SELECT "+CalendarEventSelect("")+" FROM "+CalendarEventTable+" WHERE "+CalendarEventColID+" = $1", id).
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
	rows, err := r.db.Query(ctx,
		"SELECT "+EventAttendeeSelect("")+" FROM "+EventAttendeeTable+" WHERE "+EventAttendeeColEventID+" = $1", id)
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
	rows, err := r.db.Query(ctx,
		"SELECT "+EventReminderColRemindBeforeMinutes+" FROM "+EventReminderTable+
			" WHERE "+EventReminderColEventID+" = $1", id)
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
	rows, err := r.db.Query(ctx,
		"SELECT DISTINCT e."+CalendarEventColID+" FROM "+CalendarEventTable+" e"+
			" LEFT JOIN "+EventAttendeeTable+" a ON a."+EventAttendeeColEventID+" = e."+CalendarEventColID+
			" WHERE (e."+CalendarEventColOrganizerID+" = $1 OR a."+EventAttendeeColUserID+" = $1)"+
			" AND e."+CalendarEventColStartsAt+" < $3 AND e."+CalendarEventColEndsAt+" > $2",
		userID, from, to)
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
	tag, err := r.db.Exec(ctx,
		"UPDATE "+EventAttendeeTable+" SET "+EventAttendeeColStatus+" = $3"+
			" WHERE "+EventAttendeeColEventID+" = $1 AND "+EventAttendeeColUserID+" = $2",
		eventID, userID, string(status))
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrForbidden
	}
	return nil
}

func (r *CalendarRepo) ListBusy(ctx context.Context, userIDs []domain.UserID, from, to time.Time) ([]domain.BusyInterval, error) {
	rows, err := r.db.Query(ctx,
		"SELECT e."+CalendarEventColStartsAt+", e."+CalendarEventColEndsAt+
			" FROM "+CalendarEventTable+" e"+
			" LEFT JOIN "+EventAttendeeTable+" a ON a."+EventAttendeeColEventID+" = e."+CalendarEventColID+
			" WHERE e."+CalendarEventColStatus+" = 'scheduled'"+
			" AND e."+CalendarEventColStartsAt+" < $2 AND e."+CalendarEventColEndsAt+" > $1"+
			" AND (e."+CalendarEventColOrganizerID+" = ANY($3) OR (a."+EventAttendeeColUserID+" = ANY($3) AND a."+EventAttendeeColStatus+" <> 'declined'))",
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
	reminderCols := EventReminderColumns()
	deliveryInsertCols := []string{
		ReminderDeliveryColID, ReminderDeliveryColReminderID, ReminderDeliveryColUserID, ReminderDeliveryColFireAt,
	}
	for _, m := range minutes {
		var remID uuid.UUID
		err := r.db.QueryRow(ctx,
			"INSERT INTO "+EventReminderTable+" ("+EventReminderSelect("")+") VALUES ("+placeholders(len(reminderCols))+")"+
				" ON CONFLICT ("+EventReminderColEventID+", "+EventReminderColRemindBeforeMinutes+") DO UPDATE SET "+
				EventReminderColRemindBeforeMinutes+" = EXCLUDED."+EventReminderColRemindBeforeMinutes+
				" RETURNING "+EventReminderColID,
			uuid.New(), eventID, m).Scan(&remID)
		if err != nil {
			_ = r.db.QueryRow(ctx,
				"SELECT "+EventReminderColID+" FROM "+EventReminderTable+
					" WHERE "+EventReminderColEventID+" = $1 AND "+EventReminderColRemindBeforeMinutes+" = $2",
				eventID, m).Scan(&remID)
		}
		fire := startsAt.Add(-time.Duration(m) * time.Minute)
		for _, uid := range userIDs {
			_, _ = r.db.Exec(ctx,
				"INSERT INTO "+ReminderDeliveryTable+" ("+selectList("", deliveryInsertCols)+")"+
					" VALUES ("+placeholders(len(deliveryInsertCols))+")"+
					" ON CONFLICT DO NOTHING",
				uuid.New(), remID, uid, fire)
		}
	}
	return nil
}

func (r *CalendarRepo) DueReminders(ctx context.Context, now time.Time, limit int) ([]calendar.ReminderDue, error) {
	rows, err := r.db.Query(ctx,
		"SELECT d."+ReminderDeliveryColID+", r."+EventReminderColEventID+", d."+ReminderDeliveryColUserID+
			", d."+ReminderDeliveryColFireAt+", e."+CalendarEventColTitle+
			" FROM "+ReminderDeliveryTable+" d"+
			" JOIN "+EventReminderTable+" r ON r."+EventReminderColID+" = d."+ReminderDeliveryColReminderID+
			" JOIN "+CalendarEventTable+" e ON e."+CalendarEventColID+" = r."+EventReminderColEventID+
			" WHERE d."+ReminderDeliveryColSentAt+" IS NULL AND d."+ReminderDeliveryColFireAt+" <= $1"+
			" ORDER BY d."+ReminderDeliveryColFireAt+
			" LIMIT $2",
		now, limit)
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
	_, err := r.db.Exec(ctx,
		"UPDATE "+ReminderDeliveryTable+" SET "+ReminderDeliveryColSentAt+" = $2 WHERE "+ReminderDeliveryColID+" = $1",
		id, at)
	return err
}
