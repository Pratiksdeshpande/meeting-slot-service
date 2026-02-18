package models

import (
	"database/sql"
	"time"
)

// Event represents a meeting event
type Event struct {
	ID              string             `json:"id"`
	Title           string             `json:"title" validate:"required"`
	Description     string             `json:"description"`
	OrganizerID     string             `json:"organizer_id" validate:"required"`
	DurationMinutes int                `json:"duration_minutes" validate:"required,gt=0"`
	Status          string             `json:"status"`
	ProposedSlots   []ProposedSlot     `json:"proposed_slots,omitempty"`
	Participants    []EventParticipant `json:"participants,omitempty"`
	CreatedAt       time.Time          `json:"created_at"`
	UpdatedAt       time.Time          `json:"updated_at"`
	DeletedAt       sql.NullTime       `json:"-"`
}

// EventStatusPending EventStatus constants
const (
	EventStatusPending = "pending"
)

// EventFilter represents filters for querying events
type EventFilter struct {
	OrganizerID string
	Status      string
	Page        int
	Limit       int
}
