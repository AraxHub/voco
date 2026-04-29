package domain

import (
	"time"

	"github.com/google/uuid"
)

type RoomID uuid.UUID

func NewRoomID() RoomID {
	return RoomID(uuid.New())
}

func (id RoomID) String() string {
	return uuid.UUID(id).String()
}

type RoomStatus string

const (
	RoomStatusActive RoomStatus = "active"
	RoomStatusClosed RoomStatus = "closed"
)

type JoinPolicy string

const (
	JoinPolicyOpenByLink JoinPolicy = "open_by_link"
)

type Room struct {
	ID RoomID `json:"id"`

	Title  string `json:"title,omitempty"`
	Owner  string `json:"owner,omitempty"`
	Status RoomStatus

	JoinPolicy      JoinPolicy `json:"joinPolicy"`
	MaxParticipants int        `json:"maxParticipants,omitempty"`

	CreatedAt time.Time `json:"createdAt"`
	ExpiresAt time.Time `json:"expiresAt,omitempty"`
	ClosedAt  time.Time `json:"closedAt,omitempty"`
}

func NewOpenRoomByLink(title string, now time.Time, maxParticipants int) Room {
	if title == "" {
		title = "default"
	}
	if maxParticipants <= 0 {
		maxParticipants = 10
	}

	return Room{
		ID:              NewRoomID(),
		Title:           title,
		Status:          RoomStatusActive,
		JoinPolicy:      JoinPolicyOpenByLink,
		MaxParticipants: maxParticipants,
		CreatedAt:       now,
	}
}
