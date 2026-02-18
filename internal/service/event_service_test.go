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

func TestEventService_CreateEvent_Success(t *testing.T) {
	eventRepo := new(MockEventRepository)
	userRepo := new(MockUserRepository)
	partRepo := new(MockParticipantRepository)
	svc := NewEventService(eventRepo, userRepo, partRepo)
	ctx := context.Background()

	event := baseEvent()
	userRepo.On("GetByID", ctx, "u1").Return(&models.User{ID: "u1"}, nil)
	eventRepo.On("Create", ctx, mock.AnythingOfType("*models.Event")).Return(nil)

	err := svc.CreateEvent(ctx, event)

	assert.NoError(t, err)
	assert.NotEmpty(t, event.ID)
	assert.Equal(t, models.EventStatusPending, event.Status)
	eventRepo.AssertExpectations(t)
	userRepo.AssertExpectations(t)
}

func TestEventService_CreateEvent_OrganizerNotFound(t *testing.T) {
	eventRepo := new(MockEventRepository)
	userRepo := new(MockUserRepository)
	svc := NewEventService(eventRepo, userRepo, new(MockParticipantRepository))
	ctx := context.Background()

	userRepo.On("GetByID", ctx, "u1").Return(nil, errors.New("not found"))

	err := svc.CreateEvent(ctx, baseEvent())

	assert.ErrorContains(t, err, "organizer not found")
	userRepo.AssertExpectations(t)
}

func TestEventService_CreateEvent_InvalidDuration(t *testing.T) {
	userRepo := new(MockUserRepository)
	svc := NewEventService(new(MockEventRepository), userRepo, new(MockParticipantRepository))
	ctx := context.Background()

	userRepo.On("GetByID", ctx, "u1").Return(&models.User{ID: "u1"}, nil)

	event := baseEvent()
	event.DurationMinutes = 0

	err := svc.CreateEvent(ctx, event)

	assert.EqualError(t, err, "invalid duration: must be greater than 0")
}

func TestEventService_CreateEvent_NoProposedSlots(t *testing.T) {
	userRepo := new(MockUserRepository)
	svc := NewEventService(new(MockEventRepository), userRepo, new(MockParticipantRepository))
	ctx := context.Background()

	userRepo.On("GetByID", ctx, "u1").Return(&models.User{ID: "u1"}, nil)

	event := baseEvent()
	event.ProposedSlots = nil

	err := svc.CreateEvent(ctx, event)

	assert.EqualError(t, err, "at least one proposed slot is required")
}

func TestEventService_CreateEvent_InvalidSlotTimes(t *testing.T) {
	userRepo := new(MockUserRepository)
	svc := NewEventService(new(MockEventRepository), userRepo, new(MockParticipantRepository))
	ctx := context.Background()

	userRepo.On("GetByID", ctx, "u1").Return(&models.User{ID: "u1"}, nil)

	event := baseEvent()
	// end before start
	event.ProposedSlots = []models.ProposedSlot{
		{
			StartTime: time.Date(2025, 1, 12, 10, 0, 0, 0, time.UTC),
			EndTime:   time.Date(2025, 1, 12, 9, 0, 0, 0, time.UTC),
		},
	}

	err := svc.CreateEvent(ctx, event)

	assert.ErrorContains(t, err, "invalid time slot 0")
}

func TestEventService_GetEvent_Success(t *testing.T) {
	eventRepo := new(MockEventRepository)
	svc := NewEventService(eventRepo, new(MockUserRepository), new(MockParticipantRepository))
	ctx := context.Background()

	expected := &models.Event{ID: "e1", Title: "Planning"}
	eventRepo.On("GetByID", ctx, "e1").Return(expected, nil)

	result, err := svc.GetEvent(ctx, "e1")

	assert.NoError(t, err)
	assert.Equal(t, expected, result)
	eventRepo.AssertExpectations(t)
}

func TestEventService_GetEvent_NotFound(t *testing.T) {
	eventRepo := new(MockEventRepository)
	svc := NewEventService(eventRepo, new(MockUserRepository), new(MockParticipantRepository))
	ctx := context.Background()

	eventRepo.On("GetByID", ctx, "ghost").Return(nil, errors.New("not found"))

	result, err := svc.GetEvent(ctx, "ghost")

	assert.Error(t, err)
	assert.Nil(t, result)
	eventRepo.AssertExpectations(t)
}

func TestEventService_UpdateEvent_Success(t *testing.T) {
	eventRepo := new(MockEventRepository)
	svc := NewEventService(eventRepo, new(MockUserRepository), new(MockParticipantRepository))
	ctx := context.Background()

	existing := &models.Event{ID: "e1", OrganizerID: "u1", CreatedAt: time.Now()}
	updated := &models.Event{ID: "e1", Title: "Updated"}

	eventRepo.On("GetByID", ctx, "e1").Return(existing, nil)
	eventRepo.On("Update", ctx, updated).Return(nil)

	err := svc.UpdateEvent(ctx, updated)

	assert.NoError(t, err)
	// OrganizerID and CreatedAt must be preserved from existing
	assert.Equal(t, existing.OrganizerID, updated.OrganizerID)
	assert.Equal(t, existing.CreatedAt, updated.CreatedAt)
	eventRepo.AssertExpectations(t)
}

func TestEventService_UpdateEvent_NotFound(t *testing.T) {
	eventRepo := new(MockEventRepository)
	svc := NewEventService(eventRepo, new(MockUserRepository), new(MockParticipantRepository))
	ctx := context.Background()

	eventRepo.On("GetByID", ctx, "ghost").Return(nil, errors.New("not found"))

	err := svc.UpdateEvent(ctx, &models.Event{ID: "ghost"})

	assert.Error(t, err)
	eventRepo.AssertExpectations(t)
}

func TestEventService_DeleteEvent_Success(t *testing.T) {
	eventRepo := new(MockEventRepository)
	svc := NewEventService(eventRepo, new(MockUserRepository), new(MockParticipantRepository))
	ctx := context.Background()

	eventRepo.On("GetByID", ctx, "e1").Return(&models.Event{ID: "e1"}, nil)
	eventRepo.On("Delete", ctx, "e1").Return(nil)

	err := svc.DeleteEvent(ctx, "e1")

	assert.NoError(t, err)
	eventRepo.AssertExpectations(t)
}

func TestEventService_DeleteEvent_NotFound(t *testing.T) {
	eventRepo := new(MockEventRepository)
	svc := NewEventService(eventRepo, new(MockUserRepository), new(MockParticipantRepository))
	ctx := context.Background()

	eventRepo.On("GetByID", ctx, "ghost").Return(nil, errors.New("not found"))

	err := svc.DeleteEvent(ctx, "ghost")

	assert.Error(t, err)
	eventRepo.AssertExpectations(t)
}

func TestEventService_ListEvents_DefaultPagination(t *testing.T) {
	eventRepo := new(MockEventRepository)
	svc := NewEventService(eventRepo, new(MockUserRepository), new(MockParticipantRepository))
	ctx := context.Background()

	events := []*models.Event{{ID: "e1"}}
	// page=0 → 1, limit=0 → 20
	eventRepo.On("List", ctx, models.EventFilter{Page: 1, Limit: 20}).Return(events, 1, nil)

	result, total, err := svc.ListEvents(ctx, models.EventFilter{})

	assert.NoError(t, err)
	assert.Equal(t, events, result)
	assert.Equal(t, 1, total)
	eventRepo.AssertExpectations(t)
}

func TestEventService_ListEvents_LimitCappedAt100(t *testing.T) {
	eventRepo := new(MockEventRepository)
	svc := NewEventService(eventRepo, new(MockUserRepository), new(MockParticipantRepository))
	ctx := context.Background()

	eventRepo.On("List", ctx, models.EventFilter{Page: 1, Limit: 100}).Return([]*models.Event{}, 0, nil)

	_, _, err := svc.ListEvents(ctx, models.EventFilter{Page: 1, Limit: 500})

	assert.NoError(t, err)
	eventRepo.AssertExpectations(t)
}

func TestEventService_AddParticipant_Success(t *testing.T) {
	eventRepo := new(MockEventRepository)
	userRepo := new(MockUserRepository)
	partRepo := new(MockParticipantRepository)
	svc := NewEventService(eventRepo, userRepo, partRepo)
	ctx := context.Background()

	eventRepo.On("GetByID", ctx, "e1").Return(&models.Event{ID: "e1"}, nil)
	userRepo.On("GetByID", ctx, "u1").Return(&models.User{ID: "u1"}, nil)
	partRepo.On("AddParticipant", ctx, mock.MatchedBy(func(p *models.EventParticipant) bool {
		return p.EventID == "e1" && p.UserID == "u1" && p.Status == models.ParticipantStatusInvited
	})).Return(nil)

	err := svc.AddParticipant(ctx, "e1", "u1")

	assert.NoError(t, err)
	eventRepo.AssertExpectations(t)
	userRepo.AssertExpectations(t)
	partRepo.AssertExpectations(t)
}

func TestEventService_AddParticipant_EventNotFound(t *testing.T) {
	eventRepo := new(MockEventRepository)
	svc := NewEventService(eventRepo, new(MockUserRepository), new(MockParticipantRepository))
	ctx := context.Background()

	eventRepo.On("GetByID", ctx, "ghost").Return(nil, errors.New("not found"))

	err := svc.AddParticipant(ctx, "ghost", "u1")

	assert.Error(t, err)
	eventRepo.AssertExpectations(t)
}

func TestEventService_AddParticipant_UserNotFound(t *testing.T) {
	eventRepo := new(MockEventRepository)
	userRepo := new(MockUserRepository)
	svc := NewEventService(eventRepo, userRepo, new(MockParticipantRepository))
	ctx := context.Background()

	eventRepo.On("GetByID", ctx, "e1").Return(&models.Event{ID: "e1"}, nil)
	userRepo.On("GetByID", ctx, "ghost").Return(nil, errors.New("not found"))

	err := svc.AddParticipant(ctx, "e1", "ghost")

	assert.Error(t, err)
	eventRepo.AssertExpectations(t)
	userRepo.AssertExpectations(t)
}

func TestEventService_RemoveParticipant_Success(t *testing.T) {
	partRepo := new(MockParticipantRepository)
	svc := NewEventService(new(MockEventRepository), new(MockUserRepository), partRepo)
	ctx := context.Background()

	partRepo.On("RemoveParticipant", ctx, "e1", "u1").Return(nil)

	err := svc.RemoveParticipant(ctx, "e1", "u1")

	assert.NoError(t, err)
	partRepo.AssertExpectations(t)
}

func TestEventService_GetEventParticipants_Success(t *testing.T) {
	partRepo := new(MockParticipantRepository)
	svc := NewEventService(new(MockEventRepository), new(MockUserRepository), partRepo)
	ctx := context.Background()

	expected := []models.EventParticipant{{UserID: "u1"}, {UserID: "u2"}}
	partRepo.On("GetEventParticipants", ctx, "e1").Return(expected, nil)

	result, err := svc.GetEventParticipants(ctx, "e1")

	assert.NoError(t, err)
	assert.Equal(t, expected, result)
	partRepo.AssertExpectations(t)
}

func validSlots() []models.ProposedSlot {
	return []models.ProposedSlot{
		{
			StartTime: time.Date(2025, 1, 12, 9, 0, 0, 0, time.UTC),
			EndTime:   time.Date(2025, 1, 12, 10, 0, 0, 0, time.UTC),
			Timezone:  "UTC",
		},
	}
}

func baseEvent() *models.Event {
	return &models.Event{
		Title:           "Planning",
		OrganizerID:     "u1",
		DurationMinutes: 60,
		ProposedSlots:   validSlots(),
	}
}
