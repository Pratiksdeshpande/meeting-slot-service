package service

import (
	"context"
	"fmt"
	"meeting-slot-service/internal/models"
	"meeting-slot-service/internal/repository"
	"meeting-slot-service/internal/utils"
	"sort"
	"time"
)

// RecommendationService handles slot recommendation logic
type RecommendationService struct {
	eventRepo        repository.EventRepository
	availabilityRepo repository.AvailabilityRepository
	participantRepo  repository.ParticipantRepository
}

// NewRecommendationService creates a new recommendation service
func NewRecommendationService(
	eventRepo repository.EventRepository,
	availabilityRepo repository.AvailabilityRepository,
	participantRepo repository.ParticipantRepository,
) *RecommendationService {
	return &RecommendationService{
		eventRepo:        eventRepo,
		availabilityRepo: availabilityRepo,
		participantRepo:  participantRepo,
	}
}

// GetRecommendations finds optimal meeting slots based on participant availability
func (s *RecommendationService) GetRecommendations(ctx context.Context, eventID string) (*models.RecommendationResponse, error) {
	// Get event
	event, err := s.eventRepo.GetByID(ctx, eventID)
	if err != nil {
		return nil, fmt.Errorf("event not found: %w", err)
	}

	// Get participants
	participants, err := s.participantRepo.GetEventParticipants(ctx, eventID)
	if err != nil {
		return nil, fmt.Errorf("failed to get participants: %w", err)
	}

	if len(participants) == 0 {
		return &models.RecommendationResponse{
			EventID:            eventID,
			DurationMinutes:    event.DurationMinutes,
			TotalParticipants:  0,
			BestRecommendation: nil,
			Message:            "No participants found for this event",
		}, nil
	}

	// Get all availability slots
	availabilitySlots, err := s.availabilityRepo.GetByEvent(ctx, eventID)
	if err != nil {
		return nil, fmt.Errorf("failed to get availability: %w", err)
	}

	// Build user availability map
	userAvailability := make(map[string][]utils.TimeSlot)
	for _, slot := range availabilitySlots {
		startUTC := utils.NormalizeToUTC(slot.StartTime)
		endUTC := utils.NormalizeToUTC(slot.EndTime)
		userAvailability[slot.UserID] = append(userAvailability[slot.UserID], utils.TimeSlot{
			Start: startUTC,
			End:   endUTC,
		})
	}

	// Get all participant user IDs
	participantIDs := make([]string, 0, len(participants))
	for _, p := range participants {
		participantIDs = append(participantIDs, p.UserID)
	}

	// Find best recommendation
	bestRecommendation, message := s.findBestSlot(
		event.ProposedSlots,
		event.DurationMinutes,
		participantIDs,
		userAvailability,
	)

	return &models.RecommendationResponse{
		EventID:            eventID,
		DurationMinutes:    event.DurationMinutes,
		TotalParticipants:  len(participants),
		BestRecommendation: bestRecommendation,
		Message:            message,
	}, nil
}

// findBestSlot finds the single best meeting slot based on maximum availability at earliest time
func (s *RecommendationService) findBestSlot(
	proposedSlots []models.ProposedSlot,
	durationMinutes int,
	participantIDs []string,
	userAvailability map[string][]utils.TimeSlot,
) (*models.Recommendation, string) {
	var allCandidates []models.Recommendation

	// Iterate through each proposed slot
	for _, proposedSlot := range proposedSlots {
		// Normalize proposed slot to UTC
		startUTC := utils.NormalizeToUTC(proposedSlot.StartTime)
		endUTC := utils.NormalizeToUTC(proposedSlot.EndTime)

		proposedWindow := utils.TimeSlot{
			Start: startUTC,
			End:   endUTC,
		}

		// Generate candidate slots with 15-minute intervals
		candidateSlots := utils.GenerateCandidateSlots(proposedWindow, durationMinutes, 15)

		// Check each candidate slot
		for _, candidate := range candidateSlots {
			recommendation := s.checkCandidateSlot(
				candidate,
				participantIDs,
				userAvailability,
				proposedSlot.Timezone,
			)

			allCandidates = append(allCandidates, recommendation)
		}
	}

	// No candidates found
	if len(allCandidates) == 0 {
		return nil, "No available time slots found within the proposed time windows"
	}

	// Sort: max availability first, then earliest time
	sort.Slice(allCandidates, func(i, j int) bool {
		// Primary: available participants count (descending)
		if allCandidates[i].AvailableParticipants != allCandidates[j].AvailableParticipants {
			return allCandidates[i].AvailableParticipants > allCandidates[j].AvailableParticipants
		}
		// Secondary: start time (ascending) - earliest slot wins
		return allCandidates[i].Slot.StartTime.Before(allCandidates[j].Slot.StartTime)
	})

	// Get the best recommendation
	best := allCandidates[0]

	// Generate appropriate message
	var message string
	if best.AvailabilityRate == 1.0 {
		message = fmt.Sprintf("Perfect match! All %d participants are available for this time slot", best.AvailableParticipants)
	} else if best.AvailableParticipants == 0 {
		message = "No common availability found. Consider expanding the proposed time window or collecting more availability data"
	} else {
		message = fmt.Sprintf("Best available slot with %d out of %d participants (%d%% availability)",
			best.AvailableParticipants,
			len(participantIDs),
			int(best.AvailabilityRate*100))
	}

	return &best, message
}

// checkCandidateSlot checks how many participants are available for a slot
func (s *RecommendationService) checkCandidateSlot(
	candidate utils.TimeSlot,
	participantIDs []string,
	userAvailability map[string][]utils.TimeSlot,
	timezone string,
) models.Recommendation {
	availableUsers := []string{}
	unavailableUsers := []string{}

	// Check each participant
	for _, userID := range participantIDs {
		availableSlots, exists := userAvailability[userID]

		if !exists || len(availableSlots) == 0 {
			// User hasn't submitted availability
			unavailableUsers = append(unavailableUsers, userID)
			continue
		}

		// Check if candidate slot is fully contained in any user availability slot
		isAvailable := false
		for _, availSlot := range availableSlots {
			if availSlot.Contains(candidate) {
				isAvailable = true
				break
			}
		}

		if isAvailable {
			availableUsers = append(availableUsers, userID)
		} else {
			unavailableUsers = append(unavailableUsers, userID)
		}
	}

	// Calculate availability rate
	availabilityRate := 0.0
	if len(participantIDs) > 0 {
		availabilityRate = float64(len(availableUsers)) / float64(len(participantIDs))
	}

	// Convert times back to original timezone for response
	loc, _ := time.LoadLocation(timezone)
	startInTZ := candidate.Start.In(loc)
	endInTZ := candidate.End.In(loc)

	return models.Recommendation{
		Slot: models.TimeSlot{
			StartTime: startInTZ,
			EndTime:   endInTZ,
			Timezone:  timezone,
		},
		AvailableParticipants: len(availableUsers),
		AvailabilityRate:      availabilityRate,
		AvailableUsers:        availableUsers,
		UnavailableUsers:      unavailableUsers,
	}
}
