package domain

import (
	"time"

	"github.com/google/uuid"
)

type EventID = uuid.UUID

type EventStatus string

const (
	EventScheduled EventStatus = "scheduled"
	EventCancelled EventStatus = "cancelled"
)

type RSVPStatus string

const (
	RSVPPending   RSVPStatus = "pending"
	RSVPAccepted  RSVPStatus = "accepted"
	RSVPTentative RSVPStatus = "tentative"
	RSVPDeclined  RSVPStatus = "declined"
)

type CalendarEvent struct {
	ID          EventID
	OrganizerID UserID
	Title       string
	Description string
	StartsAt    time.Time
	EndsAt      time.Time
	Timezone    string
	Status      EventStatus
	RoomID      *RoomID
	RRule       string // e.g. FREQ=WEEKLY;COUNT=10 or FREQ=DAILY;UNTIL=...
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Attendees   []EventAttendee
	Reminders   []int // minutes before
}

type EventAttendee struct {
	EventID EventID
	UserID  UserID
	Status  RSVPStatus
}

type BusyInterval struct {
	Start time.Time
	End   time.Time
}
