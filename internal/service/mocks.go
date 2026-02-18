package service

import (
	"context"
	"meeting-slot-service/internal/models"

	"github.com/stretchr/testify/mock"
)

// Mock repositories

type MockEventRepository struct {
	mock.Mock
}
type MockAvailabilityRepository struct {
	mock.Mock
}
type MockParticipantRepository struct {
	mock.Mock
}
type MockUserRepository struct {
	mock.Mock
}

func (m *MockEventRepository) Create(ctx context.Context, event *models.Event) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockEventRepository) GetByID(ctx context.Context, id string) (*models.Event, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Event), args.Error(1)
}

func (m *MockEventRepository) Update(ctx context.Context, event *models.Event) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockEventRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockEventRepository) List(ctx context.Context, filter models.EventFilter) ([]*models.Event, int, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]*models.Event), args.Int(1), args.Error(2)
}

func (m *MockAvailabilityRepository) CreateSlots(ctx context.Context, slots []models.AvailabilitySlot) error {
	args := m.Called(ctx, slots)
	return args.Error(0)
}

func (m *MockAvailabilityRepository) GetByEventAndUser(ctx context.Context, eventID, userID string) ([]models.AvailabilitySlot, error) {
	args := m.Called(ctx, eventID, userID)
	return args.Get(0).([]models.AvailabilitySlot), args.Error(1)
}

func (m *MockAvailabilityRepository) GetByEvent(ctx context.Context, eventID string) ([]models.AvailabilitySlot, error) {
	args := m.Called(ctx, eventID)
	return args.Get(0).([]models.AvailabilitySlot), args.Error(1)
}

func (m *MockAvailabilityRepository) UpdateUserSlots(ctx context.Context, eventID, userID string, slots []models.AvailabilitySlot) error {
	args := m.Called(ctx, eventID, userID, slots)
	return args.Error(0)
}

func (m *MockAvailabilityRepository) DeleteUserSlots(ctx context.Context, eventID, userID string) error {
	args := m.Called(ctx, eventID, userID)
	return args.Error(0)
}

func (m *MockParticipantRepository) AddParticipant(ctx context.Context, participant *models.EventParticipant) error {
	args := m.Called(ctx, participant)
	return args.Error(0)
}

func (m *MockParticipantRepository) GetEventParticipants(ctx context.Context, eventID string) ([]models.EventParticipant, error) {
	args := m.Called(ctx, eventID)
	return args.Get(0).([]models.EventParticipant), args.Error(1)
}

func (m *MockParticipantRepository) GetParticipant(ctx context.Context, eventID, userID string) (*models.EventParticipant, error) {
	args := m.Called(ctx, eventID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.EventParticipant), args.Error(1)
}

func (m *MockParticipantRepository) RemoveParticipant(ctx context.Context, eventID, userID string) error {
	args := m.Called(ctx, eventID, userID)
	return args.Error(0)
}

func (m *MockParticipantRepository) UpdateParticipantStatus(ctx context.Context, eventID, userID, status string) error {
	args := m.Called(ctx, eventID, userID, status)
	return args.Error(0)
}

func (m *MockUserRepository) Create(ctx context.Context, user *models.User) error {
	return m.Called(ctx, user).Error(0)
}

func (m *MockUserRepository) GetByID(ctx context.Context, id string) (*models.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) Update(ctx context.Context, user *models.User) error {
	return m.Called(ctx, user).Error(0)
}

func (m *MockUserRepository) Delete(ctx context.Context, id string) error {
	return m.Called(ctx, id).Error(0)
}

func (m *MockUserRepository) List(ctx context.Context, limit, offset int) ([]*models.User, error) {
	args := m.Called(ctx, limit, offset)
	return args.Get(0).([]*models.User), args.Error(1)
}
