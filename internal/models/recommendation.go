package models

import (
	"time"
)

// Recommendation represents a recommended time slot with availability information
type Recommendation struct {
	Slot                  TimeSlot `json:"slot"`
	AvailableParticipants int      `json:"available_participants"`
	AvailabilityRate      float64  `json:"availability_rate"`
	AvailableUsers        []string `json:"available_users"`
	UnavailableUsers      []string `json:"unavailable_users"`
}

// TimeSlot represents a time interval for recommendations
type TimeSlot struct {
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Timezone  string    `json:"timezone"`
}

// RecommendationResponse represents the API response for recommendations
type RecommendationResponse struct {
	EventID            string          `json:"event_id"`
	DurationMinutes    int             `json:"duration_minutes"`
	TotalParticipants  int             `json:"total_participants"`
	BestRecommendation *Recommendation `json:"best_recommendation"`
	Message            string          `json:"message"`
}
