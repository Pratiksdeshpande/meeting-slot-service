package repository

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"meeting-slot-service/internal/database"
	"meeting-slot-service/internal/models"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func setupEventRepoTest(t *testing.T) (*eventRepository, sqlmock.Sqlmock, func()) {
	mockDB, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	assert.NoError(t, err)

	db := &database.Database{}
	db.SetDB(mockDB)

	repo := &eventRepository{db: db}

	cleanup := func() {
		mockDB.Close()
	}

	return repo, mock, cleanup
}

func TestEventRepository_Create(t *testing.T) {
	t.Run("Success without proposed slots", func(t *testing.T) {
		repo, mock, cleanup := setupEventRepoTest(t)
		defer cleanup()

		event := &models.Event{
			ID:              "event-1",
			Title:           "Team Meeting",
			Description:     "Weekly sync",
			OrganizerID:     "user-1",
			DurationMinutes: 60,
			Status:          "draft",
		}

		mock.ExpectExec("INSERT INTO events").
			WithArgs(event.ID, event.Title, event.Description, event.OrganizerID,
				event.DurationMinutes, event.Status, sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := repo.Create(context.Background(), event)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Success with proposed slots", func(t *testing.T) {
		repo, mock, cleanup := setupEventRepoTest(t)
		defer cleanup()

		now := time.Now().UTC()
		event := &models.Event{
			ID:              "event-1",
			Title:           "Team Meeting",
			Description:     "Weekly sync",
			OrganizerID:     "user-1",
			DurationMinutes: 60,
			Status:          "draft",
			ProposedSlots: []models.ProposedSlot{
				{
					EventID:   "event-1",
					StartTime: now,
					EndTime:   now.Add(1 * time.Hour),
					Timezone:  "UTC",
				},
				{
					EventID:   "event-1",
					StartTime: now.Add(2 * time.Hour),
					EndTime:   now.Add(3 * time.Hour),
					Timezone:  "UTC",
				},
			},
		}

		mock.ExpectExec("INSERT INTO events").
			WithArgs(event.ID, event.Title, event.Description, event.OrganizerID,
				event.DurationMinutes, event.Status, sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(1, 1))

		for _, slot := range event.ProposedSlots {
			mock.ExpectExec("INSERT INTO proposed_slots").
				WithArgs(event.ID, slot.StartTime, slot.EndTime, slot.Timezone, sqlmock.AnyArg()).
				WillReturnResult(sqlmock.NewResult(1, 1))
		}

		err := repo.Create(context.Background(), event)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Database error on event creation", func(t *testing.T) {
		repo, mock, cleanup := setupEventRepoTest(t)
		defer cleanup()

		event := &models.Event{
			ID:              "event-1",
			Title:           "Team Meeting",
			Description:     "Weekly sync",
			OrganizerID:     "user-1",
			DurationMinutes: 60,
			Status:          "draft",
		}

		mock.ExpectExec("INSERT INTO events").
			WithArgs(event.ID, event.Title, event.Description, event.OrganizerID,
				event.DurationMinutes, event.Status, sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnError(errors.New("database error"))

		err := repo.Create(context.Background(), event)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create event")
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Database error on slot creation", func(t *testing.T) {
		repo, mock, cleanup := setupEventRepoTest(t)
		defer cleanup()

		now := time.Now().UTC()
		event := &models.Event{
			ID:              "event-1",
			Title:           "Team Meeting",
			Description:     "Weekly sync",
			OrganizerID:     "user-1",
			DurationMinutes: 60,
			Status:          "draft",
			ProposedSlots: []models.ProposedSlot{
				{
					EventID:   "event-1",
					StartTime: now,
					EndTime:   now.Add(1 * time.Hour),
					Timezone:  "UTC",
				},
			},
		}

		mock.ExpectExec("INSERT INTO events").
			WithArgs(event.ID, event.Title, event.Description, event.OrganizerID,
				event.DurationMinutes, event.Status, sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectExec("INSERT INTO proposed_slots").
			WithArgs(event.ID, event.ProposedSlots[0].StartTime, event.ProposedSlots[0].EndTime, event.ProposedSlots[0].Timezone, sqlmock.AnyArg()).
			WillReturnError(errors.New("slot creation error"))

		err := repo.Create(context.Background(), event)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create proposed slot")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestEventRepository_GetByID(t *testing.T) {
	t.Run("Success with slots and participants", func(t *testing.T) {
		repo, mock, cleanup := setupEventRepoTest(t)
		defer cleanup()

		eventID := "event-1"
		now := time.Now().UTC()

		// Event rows
		eventRows := sqlmock.NewRows([]string{"id", "title", "description", "organizer_id", "duration_minutes", "status", "created_at", "updated_at"}).
			AddRow(eventID, "Team Meeting", "Weekly sync", "user-1", 60, "draft", now, now)

		// Proposed slots rows
		slotRows := sqlmock.NewRows([]string{"id", "event_id", "start_time", "end_time", "timezone", "created_at"}).
			AddRow(1, eventID, now, now.Add(1*time.Hour), "UTC", now).
			AddRow(2, eventID, now.Add(2*time.Hour), now.Add(3*time.Hour), "UTC", now)

		// Participants rows
		participantRows := sqlmock.NewRows([]string{"id", "event_id", "user_id", "status", "created_at", "updated_at", "id", "name", "email", "created_at", "updated_at"}).
			AddRow(1, eventID, "user-2", "pending", now, now, "user-2", "John Doe", "john@example.com", now, now).
			AddRow(2, eventID, "user-3", "accepted", now, now, "user-3", "Jane Smith", "jane@example.com", now, now)

		mock.ExpectQuery("SELECT .+ FROM events WHERE id = (.+) AND deleted_at IS NULL").
			WithArgs(eventID).
			WillReturnRows(eventRows)

		mock.ExpectQuery("SELECT .+ FROM proposed_slots WHERE event_id = \\?").
			WithArgs(eventID).
			WillReturnRows(slotRows)

		mock.ExpectQuery("SELECT .+ FROM event_participants (.+) WHERE ep.event_id = \\?").
			WithArgs(eventID).
			WillReturnRows(participantRows)

		event, err := repo.GetByID(context.Background(), eventID)
		assert.NoError(t, err)
		assert.NotNil(t, event)
		assert.Equal(t, eventID, event.ID)
		assert.Equal(t, "Team Meeting", event.Title)
		assert.Len(t, event.ProposedSlots, 2)
		assert.Len(t, event.Participants, 2)
		assert.Equal(t, "John Doe", event.Participants[0].User.Name)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Event not found", func(t *testing.T) {
		repo, mock, cleanup := setupEventRepoTest(t)
		defer cleanup()

		eventID := "event-1"

		mock.ExpectQuery("SELECT .+ FROM events WHERE id = (.+) AND deleted_at IS NULL").
			WithArgs(eventID).
			WillReturnError(sql.ErrNoRows)

		event, err := repo.GetByID(context.Background(), eventID)
		assert.Error(t, err)
		assert.Nil(t, event)
		assert.Contains(t, err.Error(), "event not found")
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Database error on event fetch", func(t *testing.T) {
		repo, mock, cleanup := setupEventRepoTest(t)
		defer cleanup()

		eventID := "event-1"

		mock.ExpectQuery("SELECT .+ FROM events WHERE id = (.+) AND deleted_at IS NULL").
			WithArgs(eventID).
			WillReturnError(errors.New("database error"))

		event, err := repo.GetByID(context.Background(), eventID)
		assert.Error(t, err)
		assert.Nil(t, event)
		assert.Contains(t, err.Error(), "failed to get event")
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Error fetching proposed slots", func(t *testing.T) {
		repo, mock, cleanup := setupEventRepoTest(t)
		defer cleanup()

		eventID := "event-1"
		now := time.Now().UTC()

		eventRows := sqlmock.NewRows([]string{"id", "title", "description", "organizer_id", "duration_minutes", "status", "created_at", "updated_at"}).
			AddRow(eventID, "Team Meeting", "Weekly sync", "user-1", 60, "draft", now, now)

		mock.ExpectQuery("SELECT .+ FROM events WHERE id = (.+) AND deleted_at IS NULL").
			WithArgs(eventID).
			WillReturnRows(eventRows)

		mock.ExpectQuery("SELECT .+ FROM proposed_slots WHERE event_id = \\?").
			WithArgs(eventID).
			WillReturnError(errors.New("slot fetch error"))

		event, err := repo.GetByID(context.Background(), eventID)
		assert.Error(t, err)
		assert.Nil(t, event)
		assert.Contains(t, err.Error(), "failed to get proposed slots")
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Error scanning proposed slot", func(t *testing.T) {
		repo, mock, cleanup := setupEventRepoTest(t)
		defer cleanup()

		eventID := "event-1"
		now := time.Now().UTC()

		eventRows := sqlmock.NewRows([]string{"id", "title", "description", "organizer_id", "duration_minutes", "status", "created_at", "updated_at"}).
			AddRow(eventID, "Team Meeting", "Weekly sync", "user-1", 60, "draft", now, now)

		slotRows := sqlmock.NewRows([]string{"id", "event_id", "start_time", "end_time", "timezone", "created_at"}).
			AddRow(1, eventID, "invalid-time", now.Add(1*time.Hour), "UTC", now)

		mock.ExpectQuery("SELECT .+ FROM events WHERE id = (.+) AND deleted_at IS NULL").
			WithArgs(eventID).
			WillReturnRows(eventRows)

		mock.ExpectQuery("SELECT .+ FROM proposed_slots WHERE event_id = \\?").
			WithArgs(eventID).
			WillReturnRows(slotRows)

		event, err := repo.GetByID(context.Background(), eventID)
		assert.Error(t, err)
		assert.Nil(t, event)
		assert.Contains(t, err.Error(), "failed to scan proposed slot")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestEventRepository_Update(t *testing.T) {
	t.Run("Success without proposed slots", func(t *testing.T) {
		repo, mock, cleanup := setupEventRepoTest(t)
		defer cleanup()

		event := &models.Event{
			ID:              "event-1",
			Title:           "Updated Meeting",
			Description:     "Updated description",
			DurationMinutes: 90,
			Status:          "active",
		}

		mock.ExpectExec("UPDATE events SET").
			WithArgs(event.Title, event.Description, event.DurationMinutes, event.Status, event.ID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.Update(context.Background(), event)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Success with proposed slots", func(t *testing.T) {
		repo, mock, cleanup := setupEventRepoTest(t)
		defer cleanup()

		now := time.Now().UTC()
		event := &models.Event{
			ID:              "event-1",
			Title:           "Updated Meeting",
			Description:     "Updated description",
			DurationMinutes: 90,
			Status:          "active",
			ProposedSlots: []models.ProposedSlot{
				{
					EventID:   "event-1",
					StartTime: now,
					EndTime:   now.Add(1 * time.Hour),
					Timezone:  "UTC",
				},
			},
		}

		mock.ExpectExec("UPDATE events SET").
			WithArgs(event.Title, event.Description, event.DurationMinutes, event.Status, event.ID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		mock.ExpectExec("DELETE FROM proposed_slots WHERE event_id = \\?").
			WithArgs(event.ID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		mock.ExpectExec("INSERT INTO proposed_slots").
			WithArgs(event.ID, event.ProposedSlots[0].StartTime, event.ProposedSlots[0].EndTime, event.ProposedSlots[0].Timezone).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := repo.Update(context.Background(), event)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Event not found", func(t *testing.T) {
		repo, mock, cleanup := setupEventRepoTest(t)
		defer cleanup()

		event := &models.Event{
			ID:              "event-1",
			Title:           "Updated Meeting",
			Description:     "Updated description",
			DurationMinutes: 90,
			Status:          "active",
		}

		mock.ExpectExec("UPDATE events SET").
			WithArgs(event.Title, event.Description, event.DurationMinutes, event.Status, event.ID).
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := repo.Update(context.Background(), event)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "event not found")
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Database error on update", func(t *testing.T) {
		repo, mock, cleanup := setupEventRepoTest(t)
		defer cleanup()

		event := &models.Event{
			ID:              "event-1",
			Title:           "Updated Meeting",
			Description:     "Updated description",
			DurationMinutes: 90,
			Status:          "active",
		}

		mock.ExpectExec("UPDATE events SET").
			WithArgs(event.Title, event.Description, event.DurationMinutes, event.Status, event.ID).
			WillReturnError(errors.New("database error"))

		err := repo.Update(context.Background(), event)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to update event")
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Error deleting old slots", func(t *testing.T) {
		repo, mock, cleanup := setupEventRepoTest(t)
		defer cleanup()

		now := time.Now().UTC()
		event := &models.Event{
			ID:              "event-1",
			Title:           "Updated Meeting",
			Description:     "Updated description",
			DurationMinutes: 90,
			Status:          "active",
			ProposedSlots: []models.ProposedSlot{
				{
					EventID:   "event-1",
					StartTime: now,
					EndTime:   now.Add(1 * time.Hour),
					Timezone:  "UTC",
				},
			},
		}

		mock.ExpectExec("UPDATE events SET").
			WithArgs(event.Title, event.Description, event.DurationMinutes, event.Status, event.ID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		mock.ExpectExec("DELETE FROM proposed_slots WHERE event_id = \\?").
			WithArgs(event.ID).
			WillReturnError(errors.New("delete error"))

		err := repo.Update(context.Background(), event)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to delete existing proposed slots")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestEventRepository_Delete(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		repo, mock, cleanup := setupEventRepoTest(t)
		defer cleanup()

		eventID := "event-1"

		mock.ExpectExec("UPDATE events SET deleted_at = NOW\\(\\) WHERE id = \\? AND deleted_at IS NULL").
			WithArgs(eventID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.Delete(context.Background(), eventID)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Event not found", func(t *testing.T) {
		repo, mock, cleanup := setupEventRepoTest(t)
		defer cleanup()

		eventID := "event-1"

		mock.ExpectExec("UPDATE events SET deleted_at = NOW\\(\\) WHERE id = \\? AND deleted_at IS NULL").
			WithArgs(eventID).
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := repo.Delete(context.Background(), eventID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "event not found")
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Database error", func(t *testing.T) {
		repo, mock, cleanup := setupEventRepoTest(t)
		defer cleanup()

		eventID := "event-1"

		mock.ExpectExec("UPDATE events SET deleted_at = NOW\\(\\) WHERE id = \\? AND deleted_at IS NULL").
			WithArgs(eventID).
			WillReturnError(errors.New("database error"))

		err := repo.Delete(context.Background(), eventID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to delete event")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestEventRepository_List(t *testing.T) {
	t.Run("Success with filters", func(t *testing.T) {
		repo, mock, cleanup := setupEventRepoTest(t)
		defer cleanup()

		filter := models.EventFilter{
			OrganizerID: "user-1",
			Status:      "active",
			Page:        1,
			Limit:       10,
		}

		now := time.Now().UTC()

		// Count query
		countRows := sqlmock.NewRows([]string{"count"}).AddRow(2)
		mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM events WHERE deleted_at IS NULL AND organizer_id = \\? AND status = \\?").
			WithArgs(filter.OrganizerID, filter.Status).
			WillReturnRows(countRows)

		// List query
		eventRows := sqlmock.NewRows([]string{"id", "title", "description", "organizer_id", "duration_minutes", "status", "created_at", "updated_at"}).
			AddRow("event-1", "Meeting 1", "Description 1", "user-1", 60, "active", now, now).
			AddRow("event-2", "Meeting 2", "Description 2", "user-1", 90, "active", now, now)

		mock.ExpectQuery("SELECT .+ FROM events WHERE deleted_at IS NULL AND organizer_id = \\? AND status = \\? ORDER BY created_at DESC LIMIT \\? OFFSET \\?").
			WithArgs(filter.OrganizerID, filter.Status, filter.Limit, 0).
			WillReturnRows(eventRows)

		events, total, err := repo.List(context.Background(), filter)
		assert.NoError(t, err)
		assert.Len(t, events, 2)
		assert.Equal(t, 2, total)
		assert.Equal(t, "event-1", events[0].ID)
		assert.Equal(t, "event-2", events[1].ID)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Success without filters", func(t *testing.T) {
		repo, mock, cleanup := setupEventRepoTest(t)
		defer cleanup()

		filter := models.EventFilter{
			Page:  1,
			Limit: 0, // Will default to 20
		}

		now := time.Now().UTC()

		// Count query
		countRows := sqlmock.NewRows([]string{"count"}).AddRow(1)
		mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM events WHERE deleted_at IS NULL").
			WillReturnRows(countRows)

		// List query
		eventRows := sqlmock.NewRows([]string{"id", "title", "description", "organizer_id", "duration_minutes", "status", "created_at", "updated_at"}).
			AddRow("event-1", "Meeting 1", "Description 1", "user-1", 60, "active", now, now)

		mock.ExpectQuery("SELECT .+ FROM events WHERE deleted_at IS NULL ORDER BY created_at DESC LIMIT \\? OFFSET \\?").
			WithArgs(20, 0).
			WillReturnRows(eventRows)

		events, total, err := repo.List(context.Background(), filter)
		assert.NoError(t, err)
		assert.Len(t, events, 1)
		assert.Equal(t, 1, total)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Empty result", func(t *testing.T) {
		repo, mock, cleanup := setupEventRepoTest(t)
		defer cleanup()

		filter := models.EventFilter{
			Page:  1,
			Limit: 10,
		}

		// Count query
		countRows := sqlmock.NewRows([]string{"count"}).AddRow(0)
		mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM events WHERE deleted_at IS NULL").
			WillReturnRows(countRows)

		// List query
		eventRows := sqlmock.NewRows([]string{"id", "title", "description", "organizer_id", "duration_minutes", "status", "created_at", "updated_at"})

		mock.ExpectQuery("SELECT .+ FROM events WHERE deleted_at IS NULL ORDER BY created_at DESC LIMIT \\? OFFSET \\?").
			WithArgs(filter.Limit, 0).
			WillReturnRows(eventRows)

		events, total, err := repo.List(context.Background(), filter)
		assert.NoError(t, err)
		assert.Empty(t, events)
		assert.Equal(t, 0, total)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Error on count query", func(t *testing.T) {
		repo, mock, cleanup := setupEventRepoTest(t)
		defer cleanup()

		filter := models.EventFilter{
			Page:  1,
			Limit: 10,
		}

		mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM events WHERE deleted_at IS NULL").
			WillReturnError(errors.New("count error"))

		events, total, err := repo.List(context.Background(), filter)
		assert.Error(t, err)
		assert.Nil(t, events)
		assert.Equal(t, 0, total)
		assert.Contains(t, err.Error(), "failed to count events")
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Error on list query", func(t *testing.T) {
		repo, mock, cleanup := setupEventRepoTest(t)
		defer cleanup()

		filter := models.EventFilter{
			Page:  1,
			Limit: 10,
		}

		// Count query
		countRows := sqlmock.NewRows([]string{"count"}).AddRow(1)
		mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM events WHERE deleted_at IS NULL").
			WillReturnRows(countRows)

		mock.ExpectQuery("SELECT .+ FROM events WHERE deleted_at IS NULL ORDER BY created_at DESC LIMIT \\? OFFSET \\?").
			WithArgs(filter.Limit, 0).
			WillReturnError(errors.New("list error"))

		events, total, err := repo.List(context.Background(), filter)
		assert.Error(t, err)
		assert.Nil(t, events)
		assert.Equal(t, 0, total)
		assert.Contains(t, err.Error(), "failed to list events")
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Error scanning event", func(t *testing.T) {
		repo, mock, cleanup := setupEventRepoTest(t)
		defer cleanup()

		filter := models.EventFilter{
			Page:  1,
			Limit: 10,
		}

		now := time.Now().UTC()

		// Count query
		countRows := sqlmock.NewRows([]string{"count"}).AddRow(1)
		mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM events WHERE deleted_at IS NULL").
			WillReturnRows(countRows)

		// List query with invalid data
		eventRows := sqlmock.NewRows([]string{"id", "title", "description", "organizer_id", "duration_minutes", "status", "created_at", "updated_at"}).
			AddRow("event-1", "Meeting 1", "Description 1", "user-1", "invalid-number", "active", now, now)

		mock.ExpectQuery("SELECT .+ FROM events WHERE deleted_at IS NULL ORDER BY created_at DESC LIMIT \\? OFFSET \\?").
			WithArgs(filter.Limit, 0).
			WillReturnRows(eventRows)

		events, total, err := repo.List(context.Background(), filter)
		assert.Error(t, err)
		assert.Nil(t, events)
		assert.Equal(t, 0, total)
		assert.Contains(t, err.Error(), "failed to scan event")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
