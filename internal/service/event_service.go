package service

import (
	"context"
	"fmt"
	"meeting-slot-service/internal/models"
	"meeting-slot-service/internal/repository"
	"meeting-slot-service/internal/utils"
)

// EventService handles event business logic
type EventService struct {
	eventRepo       repository.EventRepository
	userRepo        repository.UserRepository
	participantRepo repository.ParticipantRepository
}

// NewEventService creates a new event service
func NewEventService(
	eventRepo repository.EventRepository,
	userRepo repository.UserRepository,
	participantRepo repository.ParticipantRepository,
) *EventService {
	return &EventService{
		eventRepo:       eventRepo,
		userRepo:        userRepo,
		participantRepo: participantRepo,
	}
}

// CreateEvent creates a new event with proposed slots
func (s *EventService) CreateEvent(ctx context.Context, event *models.Event) error {
	// Generate event ID
	if event.ID == "" {
		event.ID = utils.GenerateEventID()
	}

	// Validate organizer exists
	_, err := s.userRepo.GetByID(ctx, event.OrganizerID)
	if err != nil {
		return fmt.Errorf("organizer not found: %w", err)
	}

	// Validate duration
	if event.DurationMinutes <= 0 {
		return fmt.Errorf("invalid duration: must be greater than 0")
	}

	// Validate proposed slots
	if len(event.ProposedSlots) == 0 {
		return fmt.Errorf("at least one proposed slot is required")
	}

	for i, slot := range event.ProposedSlots {
		if slot.EndTime.Before(slot.StartTime) || slot.EndTime.Equal(slot.StartTime) {
			return fmt.Errorf("invalid time slot %d: end time must be after start time", i)
		}
	}

	// Set default status
	if event.Status == "" {
		event.Status = models.EventStatusPending
	}

	// Create event
	return s.eventRepo.Create(ctx, event)
}

// GetEvent retrieves an event by ID
func (s *EventService) GetEvent(ctx context.Context, eventID string) (*models.Event, error) {
	return s.eventRepo.GetByID(ctx, eventID)
}

// UpdateEvent updates an existing event
func (s *EventService) UpdateEvent(ctx context.Context, event *models.Event) error {
	// Check if event exists
	existing, err := s.eventRepo.GetByID(ctx, event.ID)
	if err != nil {
		return err
	}

	// Preserve certain fields
	event.CreatedAt = existing.CreatedAt
	event.OrganizerID = existing.OrganizerID

	return s.eventRepo.Update(ctx, event)
}

// DeleteEvent deletes an event
func (s *EventService) DeleteEvent(ctx context.Context, eventID string) error {
	// Check if event exists
	_, err := s.eventRepo.GetByID(ctx, eventID)
	if err != nil {
		return err
	}

	return s.eventRepo.Delete(ctx, eventID)
}

// ListEvents retrieves events with filters
func (s *EventService) ListEvents(ctx context.Context, filter models.EventFilter) ([]*models.Event, int, error) {
	// Set default pagination
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.Limit <= 0 {
		filter.Limit = 20
	}
	if filter.Limit > 100 {
		filter.Limit = 100
	}

	return s.eventRepo.List(ctx, filter)
}

// AddParticipant adds a participant to an event
func (s *EventService) AddParticipant(ctx context.Context, eventID, userID string) error {
	// Check if event exists
	_, err := s.eventRepo.GetByID(ctx, eventID)
	if err != nil {
		return err
	}

	// Check if user exists
	_, err = s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	// Add participant
	participant := &models.EventParticipant{
		EventID: eventID,
		UserID:  userID,
		Status:  models.ParticipantStatusInvited,
	}

	return s.participantRepo.AddParticipant(ctx, participant)
}

// RemoveParticipant removes a participant from an event
func (s *EventService) RemoveParticipant(ctx context.Context, eventID, userID string) error {
	return s.participantRepo.RemoveParticipant(ctx, eventID, userID)
}

// GetEventParticipants retrieves all participants of an event
func (s *EventService) GetEventParticipants(ctx context.Context, eventID string) ([]models.EventParticipant, error) {
	return s.participantRepo.GetEventParticipants(ctx, eventID)
}
