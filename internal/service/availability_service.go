package service

import (
	"context"
	"fmt"
	"meeting-slot-service/internal/models"
	"meeting-slot-service/internal/repository"
)

// AvailabilityService handles participant availability business logic
type AvailabilityService struct {
	availabilityRepo repository.AvailabilityRepository
	eventRepo        repository.EventRepository
	participantRepo  repository.ParticipantRepository
}

// NewAvailabilityService creates a new availability service
func NewAvailabilityService(
	availabilityRepo repository.AvailabilityRepository,
	eventRepo repository.EventRepository,
	participantRepo repository.ParticipantRepository,
) *AvailabilityService {
	return &AvailabilityService{
		availabilityRepo: availabilityRepo,
		eventRepo:        eventRepo,
		participantRepo:  participantRepo,
	}
}

// SubmitAvailability submits a participant's availability for an event
func (s *AvailabilityService) SubmitAvailability(ctx context.Context, eventID, userID string, slots []models.AvailabilitySlot) error {
	// Check if event exists
	_, err := s.eventRepo.GetByID(ctx, eventID)
	if err != nil {
		return fmt.Errorf("event not found: %w", err)
	}

	// Validate slots
	for i, slot := range slots {
		if slot.EndTime.Before(slot.StartTime) || slot.EndTime.Equal(slot.StartTime) {
			return fmt.Errorf("invalid time slot %d: end time must be after start time", i)
		}
		// Set event and user IDs
		slots[i].EventID = eventID
		slots[i].UserID = userID
	}

	// Create slots
	err = s.availabilityRepo.CreateSlots(ctx, slots)
	if err != nil {
		return err
	}

	// Update participant status to responded
	return s.participantRepo.UpdateParticipantStatus(ctx, eventID, userID, models.ParticipantStatusResponded)
}

// UpdateAvailability updates a participant's availability
func (s *AvailabilityService) UpdateAvailability(ctx context.Context, eventID, userID string, slots []models.AvailabilitySlot) error {
	// Check if event exists
	_, err := s.eventRepo.GetByID(ctx, eventID)
	if err != nil {
		return fmt.Errorf("event not found: %w", err)
	}

	// Validate and set IDs
	for i := range slots {
		if slots[i].EndTime.Before(slots[i].StartTime) || slots[i].EndTime.Equal(slots[i].StartTime) {
			return fmt.Errorf("invalid time slot %d: end time must be after start time", i)
		}
		slots[i].EventID = eventID
		slots[i].UserID = userID
	}

	// Update slots (delete old, insert new)
	return s.availabilityRepo.UpdateUserSlots(ctx, eventID, userID, slots)
}

// GetAvailability retrieves a participant's availability
func (s *AvailabilityService) GetAvailability(ctx context.Context, eventID, userID string) ([]models.AvailabilitySlot, error) {
	return s.availabilityRepo.GetByEventAndUser(ctx, eventID, userID)
}

// GetEventAvailability retrieves all availability for an event
func (s *AvailabilityService) GetEventAvailability(ctx context.Context, eventID string) ([]models.AvailabilitySlot, error) {
	return s.availabilityRepo.GetByEvent(ctx, eventID)
}
