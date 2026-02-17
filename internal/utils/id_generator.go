package utils

import (
	"fmt"
	"github.com/google/uuid"
	"strings"
)

// GenerateEventID generates a unique event ID with 'evt_' prefix
func GenerateEventID() string {
	id := uuid.New().String()
	// Take first 12 characters for shorter IDs
	shortID := strings.ReplaceAll(id[:13], "-", "")
	return fmt.Sprintf("evt_%s", shortID)
}

// GenerateUserID generates a unique user ID with 'usr_' prefix
func GenerateUserID() string {
	id := uuid.New().String()
	shortID := strings.ReplaceAll(id[:13], "-", "")
	return fmt.Sprintf("usr_%s", shortID)
}

// GenerateUUID generates a standard UUID
func GenerateUUID() string {
	return uuid.New().String()
}
