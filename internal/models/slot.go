package models

import (
	"time"
)

// ProposedSlot represents a time slot proposed by the organizer
type ProposedSlot struct {
	ID        uint      `json:"-"`
	EventID   string    `json:"-"`
	StartTime time.Time `json:"start_time" validate:"required"`
	EndTime   time.Time `json:"end_time" validate:"required,gtfield=StartTime"`
	Timezone  string    `json:"timezone" validate:"required"`
	CreatedAt time.Time `json:"-"`
}

// AvailabilitySlot represents a participant's available time slot
type AvailabilitySlot struct {
	ID        uint      `json:"id"`
	EventID   string    `json:"event_id"`
	UserID    string    `json:"user_id"`
	StartTime time.Time `json:"start_time" validate:"required"`
	EndTime   time.Time `json:"end_time" validate:"required,gtfield=StartTime"`
	Timezone  string    `json:"timezone" validate:"required"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
