package utils

import (
	"time"
)

// TimeSlot represents a time interval
type TimeSlot struct {
	Start time.Time
	End   time.Time
}

// Contains checks if this slot completely contains another slot
func (ts TimeSlot) Contains(other TimeSlot) bool {
	return (ts.Start.Before(other.Start) || ts.Start.Equal(other.Start)) &&
		(ts.End.After(other.End) || ts.End.Equal(other.End))
}

// Overlaps checks if two time slots overlap at all
func (ts TimeSlot) Overlaps(other TimeSlot) bool {
	return ts.Start.Before(other.End) && ts.End.After(other.Start)
}

// Duration returns the duration of the time slot
func (ts TimeSlot) Duration() time.Duration {
	return ts.End.Sub(ts.Start)
}

// NormalizeToUTC converts a time to UTC timezone
func NormalizeToUTC(t time.Time) time.Time {
	return t.UTC()
}

// GenerateCandidateSlots generates candidate time slots within a window
// using a sliding window approach with the specified interval
func GenerateCandidateSlots(window TimeSlot, durationMinutes int, intervalMinutes int) []TimeSlot {
	var candidates []TimeSlot
	duration := time.Duration(durationMinutes) * time.Minute
	interval := time.Duration(intervalMinutes) * time.Minute

	currentStart := window.Start
	for {
		candidateEnd := currentStart.Add(duration)

		// Check if candidate slot fits within the window
		if candidateEnd.After(window.End) {
			break
		}

		candidates = append(candidates, TimeSlot{
			Start: currentStart,
			End:   candidateEnd,
		})

		// Move to next interval
		currentStart = currentStart.Add(interval)

		// Prevent infinite loop
		if currentStart.After(window.End) || currentStart.Equal(window.End) {
			break
		}
	}

	return candidates
}
