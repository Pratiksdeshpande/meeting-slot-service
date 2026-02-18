package models

import (
	"time"
)

// EventParticipant represents a participant in an event
type EventParticipant struct {
	ID        uint      `json:"-"`
	EventID   string    `json:"-"`
	UserID    string    `json:"-"`
	Status    string    `json:"status"`
	User      *User     `json:"user,omitempty"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}

// ParticipantStatus constants
const (
	ParticipantStatusInvited   = "invited"
	ParticipantStatusResponded = "responded"
)
