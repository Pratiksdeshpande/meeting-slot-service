package service

import (
	"context"
	"meeting-slot-service/internal/models"
	"meeting-slot-service/internal/utils"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRecommendationService_AllParticipantsAvailable(t *testing.T) {
	mockEventRepo := new(MockEventRepository)
	mockAvailRepo := new(MockAvailabilityRepository)
	mockPartRepo := new(MockParticipantRepository)

	service := NewRecommendationService(mockEventRepo, mockAvailRepo, mockPartRepo)

	ctx := context.Background()
	eventID := "evt_123"

	// Setup test data
	event := &models.Event{
		ID:              eventID,
		DurationMinutes: 60,
		ProposedSlots: []models.ProposedSlot{
			{
				StartTime: time.Date(2025, 1, 12, 14, 0, 0, 0, time.UTC),
				EndTime:   time.Date(2025, 1, 12, 16, 0, 0, 0, time.UTC),
				Timezone:  "UTC",
			},
		},
	}

	participants := []models.EventParticipant{
		{UserID: "user1"},
		{UserID: "user2"},
	}

	availabilitySlots := []models.AvailabilitySlot{
		{
			UserID:    "user1",
			StartTime: time.Date(2025, 1, 12, 14, 0, 0, 0, time.UTC),
			EndTime:   time.Date(2025, 1, 12, 16, 0, 0, 0, time.UTC),
		},
		{
			UserID:    "user2",
			StartTime: time.Date(2025, 1, 12, 14, 0, 0, 0, time.UTC),
			EndTime:   time.Date(2025, 1, 12, 16, 0, 0, 0, time.UTC),
		},
	}

	// Setup mocks
	mockEventRepo.On("GetByID", ctx, eventID).Return(event, nil)
	mockPartRepo.On("GetEventParticipants", ctx, eventID).Return(participants, nil)
	mockAvailRepo.On("GetByEvent", ctx, eventID).Return(availabilitySlots, nil)

	// Execute
	result, err := service.GetRecommendations(ctx, eventID)

	// Verify
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, eventID, result.EventID)
	assert.Equal(t, 60, result.DurationMinutes)
	assert.Equal(t, 2, result.TotalParticipants)
	assert.NotNil(t, result.BestRecommendation)

	// Check that the best recommendation has 100% availability
	assert.Equal(t, 1.0, result.BestRecommendation.AvailabilityRate)
	assert.Equal(t, 2, result.BestRecommendation.AvailableParticipants)
	assert.Contains(t, result.Message, "Perfect match")

	mockEventRepo.AssertExpectations(t)
	mockPartRepo.AssertExpectations(t)
	mockAvailRepo.AssertExpectations(t)
}

func TestRecommendationService_PartialAvailability(t *testing.T) {
	mockEventRepo := new(MockEventRepository)
	mockAvailRepo := new(MockAvailabilityRepository)
	mockPartRepo := new(MockParticipantRepository)

	service := NewRecommendationService(mockEventRepo, mockAvailRepo, mockPartRepo)

	ctx := context.Background()
	eventID := "evt_123"

	event := &models.Event{
		ID:              eventID,
		DurationMinutes: 60,
		ProposedSlots: []models.ProposedSlot{
			{
				StartTime: time.Date(2025, 1, 12, 14, 0, 0, 0, time.UTC),
				EndTime:   time.Date(2025, 1, 12, 16, 0, 0, 0, time.UTC),
				Timezone:  "UTC",
			},
		},
	}

	participants := []models.EventParticipant{
		{UserID: "user1"},
		{UserID: "user2"},
		{UserID: "user3"},
	}

	// User1 and User2 available 14:00-15:00, User3 available 15:00-16:00
	availabilitySlots := []models.AvailabilitySlot{
		{
			UserID:    "user1",
			StartTime: time.Date(2025, 1, 12, 14, 0, 0, 0, time.UTC),
			EndTime:   time.Date(2025, 1, 12, 15, 0, 0, 0, time.UTC),
		},
		{
			UserID:    "user2",
			StartTime: time.Date(2025, 1, 12, 14, 0, 0, 0, time.UTC),
			EndTime:   time.Date(2025, 1, 12, 15, 0, 0, 0, time.UTC),
		},
		{
			UserID:    "user3",
			StartTime: time.Date(2025, 1, 12, 15, 0, 0, 0, time.UTC),
			EndTime:   time.Date(2025, 1, 12, 16, 0, 0, 0, time.UTC),
		},
	}

	mockEventRepo.On("GetByID", ctx, eventID).Return(event, nil)
	mockPartRepo.On("GetEventParticipants", ctx, eventID).Return(participants, nil)
	mockAvailRepo.On("GetByEvent", ctx, eventID).Return(availabilitySlots, nil)

	result, err := service.GetRecommendations(ctx, eventID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 3, result.TotalParticipants)
	assert.NotNil(t, result.BestRecommendation)

	// No perfect match (no time when all 3 are available)
	// Best slot should have 2 participants available (user1 and user2 at 14:00-15:00)
	assert.Less(t, result.BestRecommendation.AvailabilityRate, 1.0)
	assert.Equal(t, 2, result.BestRecommendation.AvailableParticipants)
	assert.Contains(t, result.Message, "2 out of 3")

	mockEventRepo.AssertExpectations(t)
	mockPartRepo.AssertExpectations(t)
	mockAvailRepo.AssertExpectations(t)
}

func TestCheckCandidateSlot(t *testing.T) {
	service := &RecommendationService{}

	candidate := utils.TimeSlot{
		Start: time.Date(2025, 1, 12, 14, 0, 0, 0, time.UTC),
		End:   time.Date(2025, 1, 12, 15, 0, 0, 0, time.UTC),
	}

	participantIDs := []string{"user1", "user2", "user3"}

	userAvailability := map[string][]utils.TimeSlot{
		"user1": {
			{
				Start: time.Date(2025, 1, 12, 14, 0, 0, 0, time.UTC),
				End:   time.Date(2025, 1, 12, 16, 0, 0, 0, time.UTC),
			},
		},
		"user2": {
			{
				Start: time.Date(2025, 1, 12, 14, 0, 0, 0, time.UTC),
				End:   time.Date(2025, 1, 12, 16, 0, 0, 0, time.UTC),
			},
		},
		// user3 has no availability
	}

	result := service.checkCandidateSlot(candidate, participantIDs, userAvailability, "UTC")

	assert.Equal(t, 2, result.AvailableParticipants)
	assert.Equal(t, 2.0/3.0, result.AvailabilityRate)
	assert.Contains(t, result.AvailableUsers, "user1")
	assert.Contains(t, result.AvailableUsers, "user2")
	assert.Contains(t, result.UnavailableUsers, "user3")
}
