package repository

import (
	"context"
	"meeting-slot-service/internal/models"
)

// UserRepository defines the interface for user data operations
type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
	GetByID(ctx context.Context, id string) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	Update(ctx context.Context, user *models.User) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, limit, offset int) ([]*models.User, error)
}

// EventRepository defines the interface for event data operations
type EventRepository interface {
	Create(ctx context.Context, event *models.Event) error
	GetByID(ctx context.Context, id string) (*models.Event, error)
	Update(ctx context.Context, event *models.Event) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, filter models.EventFilter) ([]*models.Event, int, error)
}

// AvailabilityRepository defines the interface for availability data operations
type AvailabilityRepository interface {
	CreateSlots(ctx context.Context, slots []models.AvailabilitySlot) error
	GetByEventAndUser(ctx context.Context, eventID, userID string) ([]models.AvailabilitySlot, error)
	GetByEvent(ctx context.Context, eventID string) ([]models.AvailabilitySlot, error)
	UpdateUserSlots(ctx context.Context, eventID, userID string, slots []models.AvailabilitySlot) error
	DeleteUserSlots(ctx context.Context, eventID, userID string) error
}

// ParticipantRepository defines the interface for participant data operations
type ParticipantRepository interface {
	AddParticipant(ctx context.Context, participant *models.EventParticipant) error
	GetEventParticipants(ctx context.Context, eventID string) ([]models.EventParticipant, error)
	RemoveParticipant(ctx context.Context, eventID, userID string) error
	UpdateParticipantStatus(ctx context.Context, eventID, userID, status string) error
}
