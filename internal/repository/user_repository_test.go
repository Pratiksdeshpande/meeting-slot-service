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

func setupUserRepoTest(t *testing.T) (*userRepository, sqlmock.Sqlmock, func()) {
	mockDB, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	assert.NoError(t, err)

	db := &database.Database{}
	db.SetDB(mockDB)

	repo := &userRepository{db: db}

	cleanup := func() {
		mockDB.Close()
	}

	return repo, mock, cleanup
}

func TestUserRepository_Create(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		repo, mock, cleanup := setupUserRepoTest(t)
		defer cleanup()

		user := &models.User{
			ID:    "user-1",
			Name:  "Test User",
			Email: "test@example.com",
		}

		mock.ExpectExec("INSERT INTO users").
			WithArgs(user.ID, user.Name, user.Email, sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := repo.Create(context.Background(), user)
		assert.NoError(t, err)
		assert.NotZero(t, user.CreatedAt)
		assert.NotZero(t, user.UpdatedAt)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Database Error", func(t *testing.T) {
		repo, mock, cleanup := setupUserRepoTest(t)
		defer cleanup()

		user := &models.User{
			ID:    "user-1",
			Name:  "Test User",
			Email: "test@example.com",
		}

		mock.ExpectExec("INSERT INTO users").
			WithArgs(user.ID, user.Name, user.Email, sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnError(errors.New("database error"))

		err := repo.Create(context.Background(), user)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create user")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestUserRepository_GetByID(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		repo, mock, cleanup := setupUserRepoTest(t)
		defer cleanup()

		userID := "user-1"
		now := time.Now().UTC()

		rows := sqlmock.NewRows([]string{"id", "name", "email", "created_at", "updated_at"}).
			AddRow(userID, "Test User", "test@example.com", now, now)

		mock.ExpectQuery("SELECT .+ FROM users WHERE id = \\?").
			WithArgs(userID).
			WillReturnRows(rows)

		user, err := repo.GetByID(context.Background(), userID)
		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, userID, user.ID)
		assert.Equal(t, "Test User", user.Name)
		assert.Equal(t, "test@example.com", user.Email)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("User Not Found", func(t *testing.T) {
		repo, mock, cleanup := setupUserRepoTest(t)
		defer cleanup()

		userID := "user-1"

		mock.ExpectQuery("SELECT .+ FROM users WHERE id = \\?").
			WithArgs(userID).
			WillReturnError(sql.ErrNoRows)

		user, err := repo.GetByID(context.Background(), userID)
		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Contains(t, err.Error(), "user not found")
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Database Error", func(t *testing.T) {
		repo, mock, cleanup := setupUserRepoTest(t)
		defer cleanup()

		userID := "user-1"

		mock.ExpectQuery("SELECT .+ FROM users WHERE id = \\?").
			WithArgs(userID).
			WillReturnError(errors.New("database error"))

		user, err := repo.GetByID(context.Background(), userID)
		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Contains(t, err.Error(), "failed to get user")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestUserRepository_GetByEmail(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		repo, mock, cleanup := setupUserRepoTest(t)
		defer cleanup()

		email := "test@example.com"
		now := time.Now().UTC()

		rows := sqlmock.NewRows([]string{"id", "name", "email", "created_at", "updated_at"}).
			AddRow("user-1", "Test User", email, now, now)

		mock.ExpectQuery("SELECT .+ FROM users WHERE email = \\?").
			WithArgs(email).
			WillReturnRows(rows)

		user, err := repo.GetByEmail(context.Background(), email)
		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, email, user.Email)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("User Not Found", func(t *testing.T) {
		repo, mock, cleanup := setupUserRepoTest(t)
		defer cleanup()

		email := "test@example.com"

		mock.ExpectQuery("SELECT .+ FROM users WHERE email = \\?").
			WithArgs(email).
			WillReturnError(sql.ErrNoRows)

		user, err := repo.GetByEmail(context.Background(), email)
		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Contains(t, err.Error(), "user not found")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestUserRepository_Update(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		repo, mock, cleanup := setupUserRepoTest(t)
		defer cleanup()

		user := &models.User{
			ID:    "user-1",
			Name:  "Updated User",
			Email: "updated@example.com",
		}

		mock.ExpectExec("UPDATE users SET").
			WithArgs(user.Name, user.Email, user.ID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.Update(context.Background(), user)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("User Not Found", func(t *testing.T) {
		repo, mock, cleanup := setupUserRepoTest(t)
		defer cleanup()

		user := &models.User{
			ID:    "user-1",
			Name:  "Updated User",
			Email: "updated@example.com",
		}

		mock.ExpectExec("UPDATE users SET").
			WithArgs(user.Name, user.Email, user.ID).
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := repo.Update(context.Background(), user)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "user not found")
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Database Error", func(t *testing.T) {
		repo, mock, cleanup := setupUserRepoTest(t)
		defer cleanup()

		user := &models.User{
			ID:    "user-1",
			Name:  "Updated User",
			Email: "updated@example.com",
		}

		mock.ExpectExec("UPDATE users SET").
			WithArgs(user.Name, user.Email, user.ID).
			WillReturnError(errors.New("database error"))

		err := repo.Update(context.Background(), user)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to update user")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestUserRepository_Delete(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		repo, mock, cleanup := setupUserRepoTest(t)
		defer cleanup()

		userID := "user-1"

		mock.ExpectExec("DELETE FROM users WHERE id = \\?").
			WithArgs(userID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.Delete(context.Background(), userID)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("User Not Found", func(t *testing.T) {
		repo, mock, cleanup := setupUserRepoTest(t)
		defer cleanup()

		userID := "user-1"

		mock.ExpectExec("DELETE FROM users WHERE id = \\?").
			WithArgs(userID).
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := repo.Delete(context.Background(), userID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "user not found")
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Database Error", func(t *testing.T) {
		repo, mock, cleanup := setupUserRepoTest(t)
		defer cleanup()

		userID := "user-1"

		mock.ExpectExec("DELETE FROM users WHERE id = \\?").
			WithArgs(userID).
			WillReturnError(errors.New("database error"))

		err := repo.Delete(context.Background(), userID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to delete user")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestUserRepository_List(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		repo, mock, cleanup := setupUserRepoTest(t)
		defer cleanup()

		now := time.Now().UTC()
		rows := sqlmock.NewRows([]string{"id", "name", "email", "created_at", "updated_at"}).
			AddRow("user-1", "User 1", "user1@example.com", now, now).
			AddRow("user-2", "User 2", "user2@example.com", now, now)

		mock.ExpectQuery("SELECT .+ FROM users ORDER BY created_at DESC LIMIT \\? OFFSET \\?").
			WithArgs(10, 0).
			WillReturnRows(rows)

		users, err := repo.List(context.Background(), 10, 0)
		assert.NoError(t, err)
		assert.Len(t, users, 2)
		assert.Equal(t, "user-1", users[0].ID)
		assert.Equal(t, "user-2", users[1].ID)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Empty Result", func(t *testing.T) {
		repo, mock, cleanup := setupUserRepoTest(t)
		defer cleanup()

		rows := sqlmock.NewRows([]string{"id", "name", "email", "created_at", "updated_at"})

		mock.ExpectQuery("SELECT .+ FROM users ORDER BY created_at DESC LIMIT \\? OFFSET \\?").
			WithArgs(10, 0).
			WillReturnRows(rows)

		users, err := repo.List(context.Background(), 10, 0)
		assert.NoError(t, err)
		assert.Empty(t, users)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Database Error", func(t *testing.T) {
		repo, mock, cleanup := setupUserRepoTest(t)
		defer cleanup()

		mock.ExpectQuery("SELECT .+ FROM users ORDER BY created_at DESC LIMIT \\? OFFSET \\?").
			WithArgs(10, 0).
			WillReturnError(errors.New("database error"))

		users, err := repo.List(context.Background(), 10, 0)
		assert.Error(t, err)
		assert.Nil(t, users)
		assert.Contains(t, err.Error(), "failed to list users")
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Scan Error", func(t *testing.T) {
		repo, mock, cleanup := setupUserRepoTest(t)
		defer cleanup()

		rows := sqlmock.NewRows([]string{"id", "name", "email", "created_at", "updated_at"}).
			AddRow("user-1", "User 1", "user1@example.com", "invalid-date", time.Now())

		mock.ExpectQuery("SELECT .+ FROM users ORDER BY created_at DESC LIMIT \\? OFFSET \\?").
			WithArgs(10, 0).
			WillReturnRows(rows)

		users, err := repo.List(context.Background(), 10, 0)
		assert.Error(t, err)
		assert.Nil(t, users)
		assert.Contains(t, err.Error(), "failed to scan user")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
