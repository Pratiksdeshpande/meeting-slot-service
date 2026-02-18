package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateEventID(t *testing.T) {
	id := GenerateEventID()

	// Should start with evt_
	assert.Contains(t, id, EventIDPrefix)
	assert.Greater(t, len(id), 10)
}

func TestGenerateUserID(t *testing.T) {
	id := GenerateUserID()

	// Should start with usr_
	assert.Contains(t, id, UserIDPrefix)
	assert.Greater(t, len(id), 10)
}

func TestGenerateUUID(t *testing.T) {
	id := generateUUID()

	// Should be a valid UUID format
	assert.Equal(t, 36, len(id)) // UUIDs are 36 characters with dashes
	assert.Contains(t, id, "-")
}

func TestUniqueIDs(t *testing.T) {
	// Generate multiple IDs and ensure they're unique
	ids := make(map[string]bool)

	for i := 0; i < 100; i++ {
		id := GenerateEventID()
		assert.False(t, ids[id], "Duplicate ID generated")
		ids[id] = true
	}
}
