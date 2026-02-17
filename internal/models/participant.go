package models

import (
	"time"
)

// EventParticipant represents a participant in an event
type EventParticipant struct {
	ID        uint      `json:"id"`
	EventID   string    `json:"event_id"`
	UserID    string    `json:"user_id"`
	Status    string    `json:"status"`
	User      *User     `json:"user,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ParticipantStatus constants
const (
	ParticipantStatusInvited   = "invited"
	ParticipantStatusResponded = "responded"
	ParticipantStatusDeclined  = "declined"
)
