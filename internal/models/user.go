package models

import (
	"time"
)

// User represents a user/participant in the system
type User struct {
	ID        string    `json:"id"`
	Name      string    `json:"name" validate:"required"`
	Email     string    `json:"email" validate:"required,email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
