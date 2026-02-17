package service

import (
	"context"
	"fmt"
	"meeting-slot-service/internal/models"
	"meeting-slot-service/internal/repository"
	"meeting-slot-service/internal/utils"
)

// UserService handles user business logic
type UserService struct {
	userRepo repository.UserRepository
}

// NewUserService creates a new user service
func NewUserService(userRepo repository.UserRepository) *UserService {
	return &UserService{
		userRepo: userRepo,
	}
}

// CreateUser creates a new user
func (s *UserService) CreateUser(ctx context.Context, user *models.User) error {
	// Generate user ID if not provided
	if user.ID == "" {
		user.ID = utils.GenerateUserID()
	}

	// Validate email
	if user.Email == "" {
		return fmt.Errorf("email is required")
	}

	// Check if email already exists
	existingUser, err := s.userRepo.GetByEmail(ctx, user.Email)
	if err == nil && existingUser != nil {
		return fmt.Errorf("email already exists")
	}

	// Create user
	return s.userRepo.Create(ctx, user)
}

// GetUser retrieves a user by ID
func (s *UserService) GetUser(ctx context.Context, userID string) (*models.User, error) {
	return s.userRepo.GetByID(ctx, userID)
}

// GetUserByEmail retrieves a user by email
func (s *UserService) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	return s.userRepo.GetByEmail(ctx, email)
}

// UpdateUser updates an existing user
func (s *UserService) UpdateUser(ctx context.Context, user *models.User) error {
	// Check if user exists
	_, err := s.userRepo.GetByID(ctx, user.ID)
	if err != nil {
		return err
	}

	return s.userRepo.Update(ctx, user)
}

// DeleteUser deletes a user
func (s *UserService) DeleteUser(ctx context.Context, userID string) error {
	return s.userRepo.Delete(ctx, userID)
}

// ListUsers retrieves users with pagination
func (s *UserService) ListUsers(ctx context.Context, page, limit int) ([]*models.User, error) {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	offset := (page - 1) * limit
	return s.userRepo.List(ctx, limit, offset)
}
