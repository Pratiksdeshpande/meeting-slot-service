package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"meeting-slot-service/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAvailabilityService_SubmitAvailability_Success(t *testing.T) {
	svc, availRepo, eventRepo, partRepo, userRepo := setupAvailabilitySvc()
	ctx := context.Background()
	slots := validAvailabilitySlots()

	eventRepo.On("GetByID", ctx, "e1").Return(&models.Event{ID: "e1"}, nil)
	userRepo.On("GetByID", ctx, "u1").Return(&models.User{ID: "u1"}, nil)
	partRepo.On("GetParticipant", ctx, "e1", "u1").Return(&models.EventParticipant{}, nil)
	availRepo.On("CreateSlots", ctx, mock.AnythingOfType("[]models.AvailabilitySlot")).Return(nil)
	partRepo.On("UpdateParticipantStatus", ctx, "e1", "u1", models.ParticipantStatusResponded).Return(nil)

	err := svc.SubmitAvailability(ctx, "e1", "u1", slots)

	assert.NoError(t, err)
	// IDs should be injected into each slot
	assert.Equal(t, "e1", slots[0].EventID)
	assert.Equal(t, "u1", slots[0].UserID)
	eventRepo.AssertExpectations(t)
	userRepo.AssertExpectations(t)
	partRepo.AssertExpectations(t)
	availRepo.AssertExpectations(t)
}

func TestAvailabilityService_SubmitAvailability_EventNotFound(t *testing.T) {
	svc, _, eventRepo, _, _ := setupAvailabilitySvc()
	ctx := context.Background()

	eventRepo.On("GetByID", ctx, "ghost").Return(nil, errors.New("not found"))

	err := svc.SubmitAvailability(ctx, "ghost", "u1", validAvailabilitySlots())

	assert.EqualError(t, err, "event not found")
	eventRepo.AssertExpectations(t)
}

func TestAvailabilityService_SubmitAvailability_UserNotFound(t *testing.T) {
	svc, _, eventRepo, _, userRepo := setupAvailabilitySvc()
	ctx := context.Background()

	eventRepo.On("GetByID", ctx, "e1").Return(&models.Event{ID: "e1"}, nil)
	userRepo.On("GetByID", ctx, "ghost").Return(nil, errors.New("not found"))

	err := svc.SubmitAvailability(ctx, "e1", "ghost", validAvailabilitySlots())

	assert.EqualError(t, err, "user not found")
	eventRepo.AssertExpectations(t)
	userRepo.AssertExpectations(t)
}

func TestAvailabilityService_SubmitAvailability_ParticipantNotFound(t *testing.T) {
	svc, _, eventRepo, partRepo, userRepo := setupAvailabilitySvc()
	ctx := context.Background()

	eventRepo.On("GetByID", ctx, "e1").Return(&models.Event{ID: "e1"}, nil)
	userRepo.On("GetByID", ctx, "u1").Return(&models.User{ID: "u1"}, nil)
	partRepo.On("GetParticipant", ctx, "e1", "u1").Return(nil, errors.New("not found"))

	err := svc.SubmitAvailability(ctx, "e1", "u1", validAvailabilitySlots())

	assert.EqualError(t, err, "participant not found")
	partRepo.AssertExpectations(t)
}

func TestAvailabilityService_SubmitAvailability_InvalidSlotTimes(t *testing.T) {
	svc, _, eventRepo, partRepo, userRepo := setupAvailabilitySvc()
	ctx := context.Background()

	eventRepo.On("GetByID", ctx, "e1").Return(&models.Event{ID: "e1"}, nil)
	userRepo.On("GetByID", ctx, "u1").Return(&models.User{ID: "u1"}, nil)
	partRepo.On("GetParticipant", ctx, "e1", "u1").Return(&models.EventParticipant{}, nil)

	badSlots := []models.AvailabilitySlot{
		{
			StartTime: time.Date(2025, 1, 12, 10, 0, 0, 0, time.UTC),
			EndTime:   time.Date(2025, 1, 12, 9, 0, 0, 0, time.UTC), // end before start
		},
	}

	err := svc.SubmitAvailability(ctx, "e1", "u1", badSlots)

	assert.ErrorContains(t, err, "invalid time slot 0")
}

func TestAvailabilityService_SubmitAvailability_RepoError(t *testing.T) {
	svc, availRepo, eventRepo, partRepo, userRepo := setupAvailabilitySvc()
	ctx := context.Background()

	eventRepo.On("GetByID", ctx, "e1").Return(&models.Event{ID: "e1"}, nil)
	userRepo.On("GetByID", ctx, "u1").Return(&models.User{ID: "u1"}, nil)
	partRepo.On("GetParticipant", ctx, "e1", "u1").Return(&models.EventParticipant{}, nil)
	availRepo.On("CreateSlots", ctx, mock.AnythingOfType("[]models.AvailabilitySlot")).Return(errors.New("db error"))

	err := svc.SubmitAvailability(ctx, "e1", "u1", validAvailabilitySlots())

	assert.EqualError(t, err, "db error")
	availRepo.AssertExpectations(t)
}

func TestAvailabilityService_UpdateAvailability_Success(t *testing.T) {
	svc, availRepo, eventRepo, partRepo, userRepo := setupAvailabilitySvc()
	ctx := context.Background()
	slots := validAvailabilitySlots()

	eventRepo.On("GetByID", ctx, "e1").Return(&models.Event{ID: "e1"}, nil)
	userRepo.On("GetByID", ctx, "u1").Return(&models.User{ID: "u1"}, nil)
	partRepo.On("GetParticipant", ctx, "e1", "u1").Return(&models.EventParticipant{}, nil)
	availRepo.On("UpdateUserSlots", ctx, "e1", "u1", mock.AnythingOfType("[]models.AvailabilitySlot")).Return(nil)

	err := svc.UpdateAvailability(ctx, "e1", "u1", slots)

	assert.NoError(t, err)
	assert.Equal(t, "e1", slots[0].EventID)
	assert.Equal(t, "u1", slots[0].UserID)
	availRepo.AssertExpectations(t)
}

func TestAvailabilityService_UpdateAvailability_EventNotFound(t *testing.T) {
	svc, _, eventRepo, _, _ := setupAvailabilitySvc()
	ctx := context.Background()

	eventRepo.On("GetByID", ctx, "ghost").Return(nil, errors.New("not found"))

	err := svc.UpdateAvailability(ctx, "ghost", "u1", validAvailabilitySlots())

	assert.EqualError(t, err, "event not found")
	eventRepo.AssertExpectations(t)
}

func TestAvailabilityService_UpdateAvailability_UserNotFound(t *testing.T) {
	svc, _, eventRepo, _, userRepo := setupAvailabilitySvc()
	ctx := context.Background()

	eventRepo.On("GetByID", ctx, "e1").Return(&models.Event{ID: "e1"}, nil)
	userRepo.On("GetByID", ctx, "ghost").Return(nil, errors.New("not found"))

	err := svc.UpdateAvailability(ctx, "e1", "ghost", validAvailabilitySlots())

	assert.EqualError(t, err, "user not found")
}

func TestAvailabilityService_UpdateAvailability_ParticipantNotFound(t *testing.T) {
	svc, _, eventRepo, partRepo, userRepo := setupAvailabilitySvc()
	ctx := context.Background()

	eventRepo.On("GetByID", ctx, "e1").Return(&models.Event{ID: "e1"}, nil)
	userRepo.On("GetByID", ctx, "u1").Return(&models.User{ID: "u1"}, nil)
	partRepo.On("GetParticipant", ctx, "e1", "u1").Return(nil, errors.New("not found"))

	err := svc.UpdateAvailability(ctx, "e1", "u1", validAvailabilitySlots())

	assert.EqualError(t, err, "participant not found")
}

func TestAvailabilityService_UpdateAvailability_InvalidSlotTimes(t *testing.T) {
	svc, _, eventRepo, partRepo, userRepo := setupAvailabilitySvc()
	ctx := context.Background()

	eventRepo.On("GetByID", ctx, "e1").Return(&models.Event{ID: "e1"}, nil)
	userRepo.On("GetByID", ctx, "u1").Return(&models.User{ID: "u1"}, nil)
	partRepo.On("GetParticipant", ctx, "e1", "u1").Return(&models.EventParticipant{}, nil)

	badSlots := []models.AvailabilitySlot{
		{
			StartTime: time.Date(2025, 1, 12, 11, 0, 0, 0, time.UTC),
			EndTime:   time.Date(2025, 1, 12, 10, 0, 0, 0, time.UTC),
		},
	}

	err := svc.UpdateAvailability(ctx, "e1", "u1", badSlots)

	assert.ErrorContains(t, err, "invalid time slot 0")
}

func TestAvailabilityService_GetAvailability_Success(t *testing.T) {
	svc, availRepo, _, _, _ := setupAvailabilitySvc()
	ctx := context.Background()

	expected := []models.AvailabilitySlot{{EventID: "e1", UserID: "u1"}}
	availRepo.On("GetByEventAndUser", ctx, "e1", "u1").Return(expected, nil)

	result, err := svc.GetAvailability(ctx, "e1", "u1")

	assert.NoError(t, err)
	assert.Equal(t, expected, result)
	availRepo.AssertExpectations(t)
}

func TestAvailabilityService_GetAvailability_RepoError(t *testing.T) {
	svc, availRepo, _, _, _ := setupAvailabilitySvc()
	ctx := context.Background()

	availRepo.On("GetByEventAndUser", ctx, "e1", "u1").Return([]models.AvailabilitySlot{}, errors.New("db error"))

	result, err := svc.GetAvailability(ctx, "e1", "u1")

	assert.EqualError(t, err, "db error")
	assert.Empty(t, result)
	availRepo.AssertExpectations(t)
}

func TestAvailabilityService_GetEventAvailability_Success(t *testing.T) {
	svc, availRepo, _, _, _ := setupAvailabilitySvc()
	ctx := context.Background()

	expected := []models.AvailabilitySlot{
		{EventID: "e1", UserID: "u1"},
		{EventID: "e1", UserID: "u2"},
	}
	availRepo.On("GetByEvent", ctx, "e1").Return(expected, nil)

	result, err := svc.GetEventAvailability(ctx, "e1")

	assert.NoError(t, err)
	assert.Len(t, result, 2)
	availRepo.AssertExpectations(t)
}

func validAvailabilitySlots() []models.AvailabilitySlot {
	return []models.AvailabilitySlot{
		{
			StartTime: time.Date(2025, 1, 12, 9, 0, 0, 0, time.UTC),
			EndTime:   time.Date(2025, 1, 12, 10, 0, 0, 0, time.UTC),
			Timezone:  "UTC",
		},
	}
}

func setupAvailabilitySvc() (
	*AvailabilityService,
	*MockAvailabilityRepository,
	*MockEventRepository,
	*MockParticipantRepository,
	*MockUserRepository,
) {
	avail := new(MockAvailabilityRepository)
	event := new(MockEventRepository)
	part := new(MockParticipantRepository)
	user := new(MockUserRepository)
	svc := NewAvailabilityService(avail, event, part, user)
	return svc, avail, event, part, user
}
