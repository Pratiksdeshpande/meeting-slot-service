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

func setupParticipantRepoTest(t *testing.T) (*participantRepository, sqlmock.Sqlmock, func()) {
	mockDB, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	assert.NoError(t, err)

	db := &database.Database{}
	db.SetDB(mockDB)

	repo := &participantRepository{db: db}

	cleanup := func() {
		mockDB.Close()
	}

	return repo, mock, cleanup
}

func TestParticipantRepository_AddParticipant(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		repo, mock, cleanup := setupParticipantRepoTest(t)
		defer cleanup()

		participant := &models.EventParticipant{
			EventID: "event-1",
			UserID:  "user-1",
			Status:  "pending",
		}

		mock.ExpectExec("INSERT INTO event_participants").
			WithArgs(participant.EventID, participant.UserID, participant.Status).
			WillReturnResult(sqlmock.NewResult(5, 1))

		err := repo.AddParticipant(context.Background(), participant)
		assert.NoError(t, err)
		assert.Equal(t, uint(5), participant.ID)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Database error", func(t *testing.T) {
		repo, mock, cleanup := setupParticipantRepoTest(t)
		defer cleanup()

		participant := &models.EventParticipant{
			EventID: "event-1",
			UserID:  "user-1",
			Status:  "pending",
		}

		mock.ExpectExec("INSERT INTO event_participants").
			WithArgs(participant.EventID, participant.UserID, participant.Status).
			WillReturnError(errors.New("database error"))

		err := repo.AddParticipant(context.Background(), participant)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to add participant")
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("LastInsertId error", func(t *testing.T) {
		repo, mock, cleanup := setupParticipantRepoTest(t)
		defer cleanup()

		participant := &models.EventParticipant{
			EventID: "event-1",
			UserID:  "user-1",
			Status:  "pending",
		}

		mock.ExpectExec("INSERT INTO event_participants").
			WithArgs(participant.EventID, participant.UserID, participant.Status).
			WillReturnResult(sqlmock.NewErrorResult(errors.New("last insert id error")))

		err := repo.AddParticipant(context.Background(), participant)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get last insert id")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestParticipantRepository_GetEventParticipants(t *testing.T) {
	t.Run("Success with multiple participants", func(t *testing.T) {
		repo, mock, cleanup := setupParticipantRepoTest(t)
		defer cleanup()

		eventID := "event-1"
		now := time.Now().UTC()

		rows := sqlmock.NewRows([]string{
			"id", "event_id", "user_id", "status", "created_at", "updated_at",
			"id", "name", "email", "created_at", "updated_at",
		}).
			AddRow(1, eventID, "user-1", "pending", now, now, "user-1", "John Doe", "john@example.com", now, now).
			AddRow(2, eventID, "user-2", "accepted", now, now, "user-2", "Jane Smith", "jane@example.com", now, now)

		mock.ExpectQuery("SELECT .+ FROM event_participants ep (.+) WHERE ep.event_id = \\?").
			WithArgs(eventID).
			WillReturnRows(rows)

		participants, err := repo.GetEventParticipants(context.Background(), eventID)
		assert.NoError(t, err)
		assert.Len(t, participants, 2)
		assert.Equal(t, uint(1), participants[0].ID)
		assert.Equal(t, "user-1", participants[0].UserID)
		assert.Equal(t, "John Doe", participants[0].User.Name)
		assert.Equal(t, "user-2", participants[1].UserID)
		assert.Equal(t, "Jane Smith", participants[1].User.Name)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Empty result", func(t *testing.T) {
		repo, mock, cleanup := setupParticipantRepoTest(t)
		defer cleanup()

		eventID := "event-1"

		rows := sqlmock.NewRows([]string{
			"id", "event_id", "user_id", "status", "created_at", "updated_at",
			"id", "name", "email", "created_at", "updated_at",
		})

		mock.ExpectQuery("SELECT .+ FROM event_participants ep (.+) WHERE ep.event_id = \\?").
			WithArgs(eventID).
			WillReturnRows(rows)

		participants, err := repo.GetEventParticipants(context.Background(), eventID)
		assert.NoError(t, err)
		assert.Empty(t, participants)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Database error", func(t *testing.T) {
		repo, mock, cleanup := setupParticipantRepoTest(t)
		defer cleanup()

		eventID := "event-1"

		mock.ExpectQuery("SELECT .+ FROM event_participants ep (.+) WHERE ep.event_id = \\?").
			WithArgs(eventID).
			WillReturnError(errors.New("database error"))

		participants, err := repo.GetEventParticipants(context.Background(), eventID)
		assert.Error(t, err)
		assert.Nil(t, participants)
		assert.Contains(t, err.Error(), "failed to get participants")
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Scan error", func(t *testing.T) {
		repo, mock, cleanup := setupParticipantRepoTest(t)
		defer cleanup()

		eventID := "event-1"
		now := time.Now().UTC()

		rows := sqlmock.NewRows([]string{
			"id", "event_id", "user_id", "status", "created_at", "updated_at",
			"id", "name", "email", "created_at", "updated_at",
		}).
			AddRow("invalid-id", eventID, "user-1", "pending", now, now, "user-1", "John Doe", "john@example.com", now, now)

		mock.ExpectQuery("SELECT .+ FROM event_participants ep (.+) WHERE ep.event_id = \\?").
			WithArgs(eventID).
			WillReturnRows(rows)

		participants, err := repo.GetEventParticipants(context.Background(), eventID)
		assert.Error(t, err)
		assert.Nil(t, participants)
		assert.Contains(t, err.Error(), "failed to scan participant")
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Row iteration error", func(t *testing.T) {
		repo, mock, cleanup := setupParticipantRepoTest(t)
		defer cleanup()

		eventID := "event-1"
		now := time.Now().UTC()

		rows := sqlmock.NewRows([]string{
			"id", "event_id", "user_id", "status", "created_at", "updated_at",
			"id", "name", "email", "created_at", "updated_at",
		}).
			AddRow(1, eventID, "user-1", "pending", now, now, "user-1", "John Doe", "john@example.com", now, now).
			RowError(0, errors.New("row iteration error"))

		mock.ExpectQuery("SELECT .+ FROM event_participants ep (.+) WHERE ep.event_id = \\?").
			WithArgs(eventID).
			WillReturnRows(rows)

		participants, err := repo.GetEventParticipants(context.Background(), eventID)
		assert.Error(t, err)
		assert.Nil(t, participants)
		assert.Contains(t, err.Error(), "error iterating rows")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestParticipantRepository_GetParticipant(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		repo, mock, cleanup := setupParticipantRepoTest(t)
		defer cleanup()

		eventID := "event-1"
		userID := "user-1"
		now := time.Now().UTC()

		rows := sqlmock.NewRows([]string{"id", "event_id", "user_id", "status", "created_at", "updated_at"}).
			AddRow(1, eventID, userID, "accepted", now, now)

		mock.ExpectQuery("SELECT .+ FROM event_participants WHERE event_id = \\? AND user_id = \\?").
			WithArgs(eventID, userID).
			WillReturnRows(rows)

		participant, err := repo.GetParticipant(context.Background(), eventID, userID)
		assert.NoError(t, err)
		assert.NotNil(t, participant)
		assert.Equal(t, uint(1), participant.ID)
		assert.Equal(t, eventID, participant.EventID)
		assert.Equal(t, userID, participant.UserID)
		assert.Equal(t, "accepted", participant.Status)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Participant not found", func(t *testing.T) {
		repo, mock, cleanup := setupParticipantRepoTest(t)
		defer cleanup()

		eventID := "event-1"
		userID := "user-1"

		mock.ExpectQuery("SELECT .+ FROM event_participants WHERE event_id = \\? AND user_id = \\?").
			WithArgs(eventID, userID).
			WillReturnError(sql.ErrNoRows)

		participant, err := repo.GetParticipant(context.Background(), eventID, userID)
		assert.Error(t, err)
		assert.Nil(t, participant)
		assert.Contains(t, err.Error(), "participant not found")
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Database error", func(t *testing.T) {
		repo, mock, cleanup := setupParticipantRepoTest(t)
		defer cleanup()

		eventID := "event-1"
		userID := "user-1"

		mock.ExpectQuery("SELECT .+ FROM event_participants WHERE event_id = \\? AND user_id = \\?").
			WithArgs(eventID, userID).
			WillReturnError(errors.New("database error"))

		participant, err := repo.GetParticipant(context.Background(), eventID, userID)
		assert.Error(t, err)
		assert.Nil(t, participant)
		assert.Contains(t, err.Error(), "participant not found")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestParticipantRepository_RemoveParticipant(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		repo, mock, cleanup := setupParticipantRepoTest(t)
		defer cleanup()

		eventID := "event-1"
		userID := "user-1"

		mock.ExpectExec("DELETE FROM event_participants WHERE event_id = \\? AND user_id = \\?").
			WithArgs(eventID, userID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.RemoveParticipant(context.Background(), eventID, userID)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Participant not found", func(t *testing.T) {
		repo, mock, cleanup := setupParticipantRepoTest(t)
		defer cleanup()

		eventID := "event-1"
		userID := "user-1"

		mock.ExpectExec("DELETE FROM event_participants WHERE event_id = \\? AND user_id = \\?").
			WithArgs(eventID, userID).
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := repo.RemoveParticipant(context.Background(), eventID, userID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "participant not found")
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Database error", func(t *testing.T) {
		repo, mock, cleanup := setupParticipantRepoTest(t)
		defer cleanup()

		eventID := "event-1"
		userID := "user-1"

		mock.ExpectExec("DELETE FROM event_participants WHERE event_id = \\? AND user_id = \\?").
			WithArgs(eventID, userID).
			WillReturnError(errors.New("database error"))

		err := repo.RemoveParticipant(context.Background(), eventID, userID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to remove participant")
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("RowsAffected error", func(t *testing.T) {
		repo, mock, cleanup := setupParticipantRepoTest(t)
		defer cleanup()

		eventID := "event-1"
		userID := "user-1"

		mock.ExpectExec("DELETE FROM event_participants WHERE event_id = \\? AND user_id = \\?").
			WithArgs(eventID, userID).
			WillReturnResult(sqlmock.NewErrorResult(errors.New("rows affected error")))

		err := repo.RemoveParticipant(context.Background(), eventID, userID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get rows affected")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestParticipantRepository_UpdateParticipantStatus(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		repo, mock, cleanup := setupParticipantRepoTest(t)
		defer cleanup()

		eventID := "event-1"
		userID := "user-1"
		status := "accepted"

		mock.ExpectExec("UPDATE event_participants SET status = \\?, updated_at = NOW\\(\\) WHERE event_id = \\? AND user_id = \\?").
			WithArgs(status, eventID, userID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.UpdateParticipantStatus(context.Background(), eventID, userID, status)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Participant not found", func(t *testing.T) {
		repo, mock, cleanup := setupParticipantRepoTest(t)
		defer cleanup()

		eventID := "event-1"
		userID := "user-1"
		status := "accepted"

		mock.ExpectExec("UPDATE event_participants SET status = \\?, updated_at = NOW\\(\\) WHERE event_id = \\? AND user_id = \\?").
			WithArgs(status, eventID, userID).
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := repo.UpdateParticipantStatus(context.Background(), eventID, userID, status)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "participant not found")
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Database error", func(t *testing.T) {
		repo, mock, cleanup := setupParticipantRepoTest(t)
		defer cleanup()

		eventID := "event-1"
		userID := "user-1"
		status := "accepted"

		mock.ExpectExec("UPDATE event_participants SET status = \\?, updated_at = NOW\\(\\) WHERE event_id = \\? AND user_id = \\?").
			WithArgs(status, eventID, userID).
			WillReturnError(errors.New("database error"))

		err := repo.UpdateParticipantStatus(context.Background(), eventID, userID, status)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to update participant status")
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("RowsAffected error", func(t *testing.T) {
		repo, mock, cleanup := setupParticipantRepoTest(t)
		defer cleanup()

		eventID := "event-1"
		userID := "user-1"
		status := "accepted"

		mock.ExpectExec("UPDATE event_participants SET status = \\?, updated_at = NOW\\(\\) WHERE event_id = \\? AND user_id = \\?").
			WithArgs(status, eventID, userID).
			WillReturnResult(sqlmock.NewErrorResult(errors.New("rows affected error")))

		err := repo.UpdateParticipantStatus(context.Background(), eventID, userID, status)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get rows affected")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
