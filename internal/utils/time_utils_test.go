package utils

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTimeSlotContains(t *testing.T) {
	tests := []struct {
		name     string
		outer    TimeSlot
		inner    TimeSlot
		expected bool
	}{
		{
			name: "completely contains",
			outer: TimeSlot{
				Start: time.Date(2025, 1, 12, 14, 0, 0, 0, time.UTC),
				End:   time.Date(2025, 1, 12, 16, 0, 0, 0, time.UTC),
			},
			inner: TimeSlot{
				Start: time.Date(2025, 1, 12, 14, 30, 0, 0, time.UTC),
				End:   time.Date(2025, 1, 12, 15, 30, 0, 0, time.UTC),
			},
			expected: true,
		},
		{
			name: "exact match",
			outer: TimeSlot{
				Start: time.Date(2025, 1, 12, 14, 0, 0, 0, time.UTC),
				End:   time.Date(2025, 1, 12, 16, 0, 0, 0, time.UTC),
			},
			inner: TimeSlot{
				Start: time.Date(2025, 1, 12, 14, 0, 0, 0, time.UTC),
				End:   time.Date(2025, 1, 12, 16, 0, 0, 0, time.UTC),
			},
			expected: true,
		},
		{
			name: "does not contain - overlaps start",
			outer: TimeSlot{
				Start: time.Date(2025, 1, 12, 14, 0, 0, 0, time.UTC),
				End:   time.Date(2025, 1, 12, 16, 0, 0, 0, time.UTC),
			},
			inner: TimeSlot{
				Start: time.Date(2025, 1, 12, 13, 30, 0, 0, time.UTC),
				End:   time.Date(2025, 1, 12, 15, 0, 0, 0, time.UTC),
			},
			expected: false,
		},
		{
			name: "does not contain - overlaps end",
			outer: TimeSlot{
				Start: time.Date(2025, 1, 12, 14, 0, 0, 0, time.UTC),
				End:   time.Date(2025, 1, 12, 16, 0, 0, 0, time.UTC),
			},
			inner: TimeSlot{
				Start: time.Date(2025, 1, 12, 15, 0, 0, 0, time.UTC),
				End:   time.Date(2025, 1, 12, 17, 0, 0, 0, time.UTC),
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.outer.Contains(tt.inner)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTimeSlotOverlaps(t *testing.T) {
	tests := []struct {
		name     string
		slot1    TimeSlot
		slot2    TimeSlot
		expected bool
	}{
		{
			name: "overlaps in middle",
			slot1: TimeSlot{
				Start: time.Date(2025, 1, 12, 14, 0, 0, 0, time.UTC),
				End:   time.Date(2025, 1, 12, 16, 0, 0, 0, time.UTC),
			},
			slot2: TimeSlot{
				Start: time.Date(2025, 1, 12, 15, 0, 0, 0, time.UTC),
				End:   time.Date(2025, 1, 12, 17, 0, 0, 0, time.UTC),
			},
			expected: true,
		},
		{
			name: "no overlap - completely before",
			slot1: TimeSlot{
				Start: time.Date(2025, 1, 12, 14, 0, 0, 0, time.UTC),
				End:   time.Date(2025, 1, 12, 16, 0, 0, 0, time.UTC),
			},
			slot2: TimeSlot{
				Start: time.Date(2025, 1, 12, 16, 0, 0, 0, time.UTC),
				End:   time.Date(2025, 1, 12, 18, 0, 0, 0, time.UTC),
			},
			expected: false,
		},
		{
			name: "no overlap - completely after",
			slot1: TimeSlot{
				Start: time.Date(2025, 1, 12, 16, 0, 0, 0, time.UTC),
				End:   time.Date(2025, 1, 12, 18, 0, 0, 0, time.UTC),
			},
			slot2: TimeSlot{
				Start: time.Date(2025, 1, 12, 14, 0, 0, 0, time.UTC),
				End:   time.Date(2025, 1, 12, 16, 0, 0, 0, time.UTC),
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.slot1.Overlaps(tt.slot2)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGenerateCandidateSlots(t *testing.T) {
	window := TimeSlot{
		Start: time.Date(2025, 1, 12, 14, 0, 0, 0, time.UTC),
		End:   time.Date(2025, 1, 12, 16, 0, 0, 0, time.UTC),
	}

	// 60 minute meeting, 15 minute intervals
	candidates := GenerateCandidateSlots(window, 60, 15)

	// Should generate: 14:00-15:00, 14:15-15:15, 14:30-15:30, 14:45-15:45, 15:00-16:00
	assert.Equal(t, 5, len(candidates))

	// Check first candidate
	assert.Equal(t, time.Date(2025, 1, 12, 14, 0, 0, 0, time.UTC), candidates[0].Start)
	assert.Equal(t, time.Date(2025, 1, 12, 15, 0, 0, 0, time.UTC), candidates[0].End)

	// Check last candidate
	assert.Equal(t, time.Date(2025, 1, 12, 15, 0, 0, 0, time.UTC), candidates[4].Start)
	assert.Equal(t, time.Date(2025, 1, 12, 16, 0, 0, 0, time.UTC), candidates[4].End)
}

func TestNormalizeToUTC(t *testing.T) {
	// Create a time in EST
	est, _ := time.LoadLocation("America/New_York")
	localTime := time.Date(2025, 1, 12, 14, 0, 0, 0, est)

	utcTime := NormalizeToUTC(localTime)

	// Verify it's in UTC
	assert.Equal(t, "UTC", utcTime.Location().String())

	// Verify the time is correct (EST is UTC-5)
	assert.Equal(t, 19, utcTime.Hour()) // 14 + 5 = 19
}
