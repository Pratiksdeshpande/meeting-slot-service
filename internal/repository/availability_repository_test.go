package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"meeting-slot-service/internal/database"
	"meeting-slot-service/internal/models"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func setupAvailabilityRepoTest(t *testing.T) (*availabilityRepository, sqlmock.Sqlmock, func()) {
	mockDB, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	assert.NoError(t, err)

	db := &database.Database{}
	db.SetDB(mockDB)

	repo := &availabilityRepository{db: db}

	cleanup := func() {
		mockDB.Close()
	}

	return repo, mock, cleanup
}

func TestAvailabilityRepository_CreateSlots(t *testing.T) {
	t.Run("Success with multiple slots", func(t *testing.T) {
		repo, mock, cleanup := setupAvailabilityRepoTest(t)
		defer cleanup()

		now := time.Now().UTC()
		slots := []models.AvailabilitySlot{
			{
				EventID:   "event-1",
				UserID:    "user-1",
				StartTime: now,
				EndTime:   now.Add(1 * time.Hour),
				Timezone:  "UTC",
			},
			{
				EventID:   "event-1",
				UserID:    "user-1",
				StartTime: now.Add(2 * time.Hour),
				EndTime:   now.Add(3 * time.Hour),
				Timezone:  "UTC",
			},
		}

		for _, slot := range slots {
			mock.ExpectExec("INSERT INTO availability_slots").
				WithArgs(slot.EventID, slot.UserID, slot.StartTime, slot.EndTime, slot.Timezone).
				WillReturnResult(sqlmock.NewResult(1, 1))
		}

		err := repo.CreateSlots(context.Background(), slots)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Empty slots array", func(t *testing.T) {
		repo, _, cleanup := setupAvailabilityRepoTest(t)
		defer cleanup()

		var slots []models.AvailabilitySlot

		err := repo.CreateSlots(context.Background(), slots)
		assert.NoError(t, err)
	})

	t.Run("Database error", func(t *testing.T) {
		repo, mock, cleanup := setupAvailabilityRepoTest(t)
		defer cleanup()

		now := time.Now().UTC()
		slots := []models.AvailabilitySlot{
			{
				EventID:   "event-1",
				UserID:    "user-1",
				StartTime: now,
				EndTime:   now.Add(1 * time.Hour),
				Timezone:  "UTC",
			},
		}

		mock.ExpectExec("INSERT INTO availability_slots").
			WithArgs(slots[0].EventID, slots[0].UserID, slots[0].StartTime, slots[0].EndTime, slots[0].Timezone).
			WillReturnError(errors.New("database error"))

		err := repo.CreateSlots(context.Background(), slots)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create availability slot")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestAvailabilityRepository_GetByEventAndUser(t *testing.T) {
	t.Run("Success with multiple slots", func(t *testing.T) {
		repo, mock, cleanup := setupAvailabilityRepoTest(t)
		defer cleanup()

		eventID := "event-1"
		userID := "user-1"
		now := time.Now().UTC()

		rows := sqlmock.NewRows([]string{"id", "event_id", "user_id", "start_time", "end_time", "timezone", "created_at", "updated_at"}).
			AddRow(1, eventID, userID, now, now.Add(1*time.Hour), "UTC", now, now).
			AddRow(2, eventID, userID, now.Add(2*time.Hour), now.Add(3*time.Hour), "UTC", now, now)

		mock.ExpectQuery("SELECT .+ FROM availability_slots WHERE event_id = \\? AND user_id = \\? ORDER BY start_time ASC").
			WithArgs(eventID, userID).
			WillReturnRows(rows)

		slots, err := repo.GetByEventAndUser(context.Background(), eventID, userID)
		assert.NoError(t, err)
		assert.Len(t, slots, 2)
		assert.Equal(t, uint(1), slots[0].ID)
		assert.Equal(t, eventID, slots[0].EventID)
		assert.Equal(t, userID, slots[0].UserID)
		assert.Equal(t, "UTC", slots[0].Timezone)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Empty result", func(t *testing.T) {
		repo, mock, cleanup := setupAvailabilityRepoTest(t)
		defer cleanup()

		eventID := "event-1"
		userID := "user-1"

		rows := sqlmock.NewRows([]string{"id", "event_id", "user_id", "start_time", "end_time", "timezone", "created_at", "updated_at"})

		mock.ExpectQuery("SELECT .+ FROM availability_slots WHERE event_id = \\? AND user_id = \\? ORDER BY start_time ASC").
			WithArgs(eventID, userID).
			WillReturnRows(rows)

		slots, err := repo.GetByEventAndUser(context.Background(), eventID, userID)
		assert.NoError(t, err)
		assert.Empty(t, slots)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Database error", func(t *testing.T) {
		repo, mock, cleanup := setupAvailabilityRepoTest(t)
		defer cleanup()

		eventID := "event-1"
		userID := "user-1"

		mock.ExpectQuery("SELECT .+ FROM availability_slots WHERE event_id = \\? AND user_id = \\? ORDER BY start_time ASC").
			WithArgs(eventID, userID).
			WillReturnError(errors.New("database error"))

		slots, err := repo.GetByEventAndUser(context.Background(), eventID, userID)
		assert.Error(t, err)
		assert.Nil(t, slots)
		assert.Contains(t, err.Error(), "failed to get availability slots")
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Scan error", func(t *testing.T) {
		repo, mock, cleanup := setupAvailabilityRepoTest(t)
		defer cleanup()

		eventID := "event-1"
		userID := "user-1"
		now := time.Now().UTC()

		rows := sqlmock.NewRows([]string{"id", "event_id", "user_id", "start_time", "end_time", "timezone", "created_at", "updated_at"}).
			AddRow("invalid-id", eventID, userID, now, now.Add(1*time.Hour), "UTC", now, now)

		mock.ExpectQuery("SELECT .+ FROM availability_slots WHERE event_id = \\? AND user_id = \\? ORDER BY start_time ASC").
			WithArgs(eventID, userID).
			WillReturnRows(rows)

		slots, err := repo.GetByEventAndUser(context.Background(), eventID, userID)
		assert.Error(t, err)
		assert.Nil(t, slots)
		assert.Contains(t, err.Error(), "failed to scan availability slot")
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Row iteration error", func(t *testing.T) {
		repo, mock, cleanup := setupAvailabilityRepoTest(t)
		defer cleanup()

		eventID := "event-1"
		userID := "user-1"
		now := time.Now().UTC()

		rows := sqlmock.NewRows([]string{"id", "event_id", "user_id", "start_time", "end_time", "timezone", "created_at", "updated_at"}).
			AddRow(1, eventID, userID, now, now.Add(1*time.Hour), "UTC", now, now).
			RowError(0, errors.New("row iteration error"))

		mock.ExpectQuery("SELECT .+ FROM availability_slots WHERE event_id = \\? AND user_id = \\? ORDER BY start_time ASC").
			WithArgs(eventID, userID).
			WillReturnRows(rows)

		slots, err := repo.GetByEventAndUser(context.Background(), eventID, userID)
		assert.Error(t, err)
		assert.Nil(t, slots)
		assert.Contains(t, err.Error(), "error iterating rows")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestAvailabilityRepository_GetByEvent(t *testing.T) {
	t.Run("Success with multiple users", func(t *testing.T) {
		repo, mock, cleanup := setupAvailabilityRepoTest(t)
		defer cleanup()

		eventID := "event-1"
		now := time.Now().UTC()

		rows := sqlmock.NewRows([]string{"id", "event_id", "user_id", "start_time", "end_time", "timezone", "created_at", "updated_at"}).
			AddRow(1, eventID, "user-1", now, now.Add(1*time.Hour), "UTC", now, now).
			AddRow(2, eventID, "user-1", now.Add(2*time.Hour), now.Add(3*time.Hour), "UTC", now, now).
			AddRow(3, eventID, "user-2", now, now.Add(1*time.Hour), "America/New_York", now, now)

		mock.ExpectQuery("SELECT .+ FROM availability_slots WHERE event_id = \\? ORDER BY user_id ASC, start_time ASC").
			WithArgs(eventID).
			WillReturnRows(rows)

		slots, err := repo.GetByEvent(context.Background(), eventID)
		assert.NoError(t, err)
		assert.Len(t, slots, 3)
		assert.Equal(t, "user-1", slots[0].UserID)
		assert.Equal(t, "user-2", slots[2].UserID)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Empty result", func(t *testing.T) {
		repo, mock, cleanup := setupAvailabilityRepoTest(t)
		defer cleanup()

		eventID := "event-1"

		rows := sqlmock.NewRows([]string{"id", "event_id", "user_id", "start_time", "end_time", "timezone", "created_at", "updated_at"})

		mock.ExpectQuery("SELECT .+ FROM availability_slots WHERE event_id = \\? ORDER BY user_id ASC, start_time ASC").
			WithArgs(eventID).
			WillReturnRows(rows)

		slots, err := repo.GetByEvent(context.Background(), eventID)
		assert.NoError(t, err)
		assert.Empty(t, slots)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Database error", func(t *testing.T) {
		repo, mock, cleanup := setupAvailabilityRepoTest(t)
		defer cleanup()

		eventID := "event-1"

		mock.ExpectQuery("SELECT .+ FROM availability_slots WHERE event_id = \\? ORDER BY user_id ASC, start_time ASC").
			WithArgs(eventID).
			WillReturnError(errors.New("database error"))

		slots, err := repo.GetByEvent(context.Background(), eventID)
		assert.Error(t, err)
		assert.Nil(t, slots)
		assert.Contains(t, err.Error(), "failed to get availability slots")
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Scan error", func(t *testing.T) {
		repo, mock, cleanup := setupAvailabilityRepoTest(t)
		defer cleanup()

		eventID := "event-1"
		now := time.Now().UTC()

		rows := sqlmock.NewRows([]string{"id", "event_id", "user_id", "start_time", "end_time", "timezone", "created_at", "updated_at"}).
			AddRow(1, eventID, "user-1", "invalid-time", now.Add(1*time.Hour), "UTC", now, now) // Invalid time type

		mock.ExpectQuery("SELECT .+ FROM availability_slots WHERE event_id = \\? ORDER BY user_id ASC, start_time ASC").
			WithArgs(eventID).
			WillReturnRows(rows)

		slots, err := repo.GetByEvent(context.Background(), eventID)
		assert.Error(t, err)
		assert.Nil(t, slots)
		assert.Contains(t, err.Error(), "failed to scan availability slot")
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Row iteration error", func(t *testing.T) {
		repo, mock, cleanup := setupAvailabilityRepoTest(t)
		defer cleanup()

		eventID := "event-1"
		now := time.Now().UTC()

		rows := sqlmock.NewRows([]string{"id", "event_id", "user_id", "start_time", "end_time", "timezone", "created_at", "updated_at"}).
			AddRow(1, eventID, "user-1", now, now.Add(1*time.Hour), "UTC", now, now).
			RowError(0, errors.New("iteration error"))

		mock.ExpectQuery("SELECT .+ FROM availability_slots WHERE event_id = \\? ORDER BY user_id ASC, start_time ASC").
			WithArgs(eventID).
			WillReturnRows(rows)

		slots, err := repo.GetByEvent(context.Background(), eventID)
		assert.Error(t, err)
		assert.Nil(t, slots)
		assert.Contains(t, err.Error(), "error iterating rows")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestAvailabilityRepository_UpdateUserSlots(t *testing.T) {
	t.Run("Success with new slots", func(t *testing.T) {
		repo, mock, cleanup := setupAvailabilityRepoTest(t)
		defer cleanup()

		eventID := "event-1"
		userID := "user-1"
		now := time.Now().UTC()
		slots := []models.AvailabilitySlot{
			{
				EventID:   eventID,
				UserID:    userID,
				StartTime: now,
				EndTime:   now.Add(1 * time.Hour),
				Timezone:  "UTC",
			},
			{
				EventID:   eventID,
				UserID:    userID,
				StartTime: now.Add(2 * time.Hour),
				EndTime:   now.Add(3 * time.Hour),
				Timezone:  "UTC",
			},
		}

		mock.ExpectBegin()
		mock.ExpectExec("DELETE FROM availability_slots WHERE event_id = \\? AND user_id = \\?").
			WithArgs(eventID, userID).
			WillReturnResult(sqlmock.NewResult(0, 2))

		for _, slot := range slots {
			mock.ExpectExec("INSERT INTO availability_slots").
				WithArgs(slot.EventID, slot.UserID, slot.StartTime, slot.EndTime, slot.Timezone).
				WillReturnResult(sqlmock.NewResult(1, 1))
		}

		mock.ExpectCommit()

		err := repo.UpdateUserSlots(context.Background(), eventID, userID, slots)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Success with empty slots (delete all)", func(t *testing.T) {
		repo, mock, cleanup := setupAvailabilityRepoTest(t)
		defer cleanup()

		eventID := "event-1"
		userID := "user-1"
		var slots []models.AvailabilitySlot

		mock.ExpectBegin()
		mock.ExpectExec("DELETE FROM availability_slots WHERE event_id = \\? AND user_id = \\?").
			WithArgs(eventID, userID).
			WillReturnResult(sqlmock.NewResult(0, 2))
		mock.ExpectCommit()

		err := repo.UpdateUserSlots(context.Background(), eventID, userID, slots)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Error beginning transaction", func(t *testing.T) {
		repo, mock, cleanup := setupAvailabilityRepoTest(t)
		defer cleanup()

		eventID := "event-1"
		userID := "user-1"
		var slots []models.AvailabilitySlot

		mock.ExpectBegin().WillReturnError(errors.New("begin tx error"))

		err := repo.UpdateUserSlots(context.Background(), eventID, userID, slots)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to begin transaction")
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Error deleting old slots", func(t *testing.T) {
		repo, mock, cleanup := setupAvailabilityRepoTest(t)
		defer cleanup()

		eventID := "event-1"
		userID := "user-1"
		var slots []models.AvailabilitySlot

		mock.ExpectBegin()
		mock.ExpectExec("DELETE FROM availability_slots WHERE event_id = \\? AND user_id = \\?").
			WithArgs(eventID, userID).
			WillReturnError(errors.New("delete error"))
		mock.ExpectRollback()

		err := repo.UpdateUserSlots(context.Background(), eventID, userID, slots)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to delete old slots")
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Error creating new slot", func(t *testing.T) {
		repo, mock, cleanup := setupAvailabilityRepoTest(t)
		defer cleanup()

		eventID := "event-1"
		userID := "user-1"
		now := time.Now().UTC()
		slots := []models.AvailabilitySlot{
			{
				EventID:   eventID,
				UserID:    userID,
				StartTime: now,
				EndTime:   now.Add(1 * time.Hour),
				Timezone:  "UTC",
			},
		}

		mock.ExpectBegin()
		mock.ExpectExec("DELETE FROM availability_slots WHERE event_id = \\? AND user_id = \\?").
			WithArgs(eventID, userID).
			WillReturnResult(sqlmock.NewResult(0, 0))

		mock.ExpectExec("INSERT INTO availability_slots").
			WithArgs(slots[0].EventID, slots[0].UserID, slots[0].StartTime, slots[0].EndTime, slots[0].Timezone).
			WillReturnError(errors.New("insert error"))

		mock.ExpectRollback()

		err := repo.UpdateUserSlots(context.Background(), eventID, userID, slots)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create new slot")
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Error committing transaction", func(t *testing.T) {
		repo, mock, cleanup := setupAvailabilityRepoTest(t)
		defer cleanup()

		eventID := "event-1"
		userID := "user-1"
		var slots []models.AvailabilitySlot

		mock.ExpectBegin()
		mock.ExpectExec("DELETE FROM availability_slots WHERE event_id = \\? AND user_id = \\?").
			WithArgs(eventID, userID).
			WillReturnResult(sqlmock.NewResult(0, 0))
		mock.ExpectCommit().WillReturnError(errors.New("commit error"))

		err := repo.UpdateUserSlots(context.Background(), eventID, userID, slots)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to commit transaction")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestAvailabilityRepository_DeleteUserSlots(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		repo, mock, cleanup := setupAvailabilityRepoTest(t)
		defer cleanup()

		eventID := "event-1"
		userID := "user-1"

		mock.ExpectExec("DELETE FROM availability_slots WHERE event_id = \\? AND user_id = \\?").
			WithArgs(eventID, userID).
			WillReturnResult(sqlmock.NewResult(0, 3))

		err := repo.DeleteUserSlots(context.Background(), eventID, userID)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Success with no rows deleted", func(t *testing.T) {
		repo, mock, cleanup := setupAvailabilityRepoTest(t)
		defer cleanup()

		eventID := "event-1"
		userID := "user-1"

		mock.ExpectExec("DELETE FROM availability_slots WHERE event_id = \\? AND user_id = \\?").
			WithArgs(eventID, userID).
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := repo.DeleteUserSlots(context.Background(), eventID, userID)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Database error", func(t *testing.T) {
		repo, mock, cleanup := setupAvailabilityRepoTest(t)
		defer cleanup()

		eventID := "event-1"
		userID := "user-1"

		mock.ExpectExec("DELETE FROM availability_slots WHERE event_id = \\? AND user_id = \\?").
			WithArgs(eventID, userID).
			WillReturnError(errors.New("database error"))

		err := repo.DeleteUserSlots(context.Background(), eventID, userID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to delete availability slots")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
