package utils

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
)

const (
	EventIDPrefix = "evt_"
	UserIDPrefix  = "usr_"
)

// GenerateEventID generates a unique event ID with 'evt_' prefix
func GenerateEventID() string {
	id := generateUUID()
	// Take first 12 characters for shorter IDs
	shortID := strings.ReplaceAll(id[:13], "-", "")
	return fmt.Sprintf("%s%s", EventIDPrefix, shortID)
}

// GenerateUserID generates a unique user ID with 'usr_' prefix
func GenerateUserID() string {
	id := generateUUID()
	shortID := strings.ReplaceAll(id[:13], "-", "")
	return fmt.Sprintf("%s%s", UserIDPrefix, shortID)
}

// generateUUID generates a standard UUID
func generateUUID() string {
	return uuid.New().String()
}
