package service

import (
	"context"
	"errors"
	"testing"

	"meeting-slot-service/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUserService_CreateUser_Success(t *testing.T) {
	repo := new(MockUserRepository)
	svc := NewUserService(repo)
	ctx := context.Background()

	user := &models.User{Name: "Alice", Email: "alice@example.com"}

	repo.On("GetByEmail", ctx, "alice@example.com").Return(nil, errors.New("not found"))
	repo.On("Create", ctx, mock.AnythingOfType("*models.User")).Return(nil)

	err := svc.CreateUser(ctx, user)

	assert.NoError(t, err)
	assert.NotEmpty(t, user.ID) // ID should be generated
	repo.AssertExpectations(t)
}

func TestUserService_CreateUser_MissingEmail(t *testing.T) {
	svc := NewUserService(new(MockUserRepository))
	ctx := context.Background()

	err := svc.CreateUser(ctx, &models.User{Name: "Bob"})

	assert.EqualError(t, err, "email is required")
}

func TestUserService_CreateUser_DuplicateEmail(t *testing.T) {
	repo := new(MockUserRepository)
	svc := NewUserService(repo)
	ctx := context.Background()

	existing := &models.User{ID: "u1", Email: "dup@example.com"}
	repo.On("GetByEmail", ctx, "dup@example.com").Return(existing, nil)

	err := svc.CreateUser(ctx, &models.User{Name: "Carol", Email: "dup@example.com"})

	assert.EqualError(t, err, "email already exists")
	repo.AssertExpectations(t)
}

func TestUserService_CreateUser_RepoError(t *testing.T) {
	repo := new(MockUserRepository)
	svc := NewUserService(repo)
	ctx := context.Background()

	repo.On("GetByEmail", ctx, "err@example.com").Return(nil, errors.New("not found"))
	repo.On("Create", ctx, mock.AnythingOfType("*models.User")).Return(errors.New("db error"))

	err := svc.CreateUser(ctx, &models.User{Name: "Dan", Email: "err@example.com"})

	assert.EqualError(t, err, "db error")
	repo.AssertExpectations(t)
}

func TestUserService_GetUser_Success(t *testing.T) {
	repo := new(MockUserRepository)
	svc := NewUserService(repo)
	ctx := context.Background()

	expected := &models.User{ID: "u1", Name: "Alice", Email: "alice@example.com"}
	repo.On("GetByID", ctx, "u1").Return(expected, nil)

	user, err := svc.GetUser(ctx, "u1")

	assert.NoError(t, err)
	assert.Equal(t, expected, user)
	repo.AssertExpectations(t)
}

func TestUserService_GetUser_NotFound(t *testing.T) {
	repo := new(MockUserRepository)
	svc := NewUserService(repo)
	ctx := context.Background()

	repo.On("GetByID", ctx, "missing").Return(nil, errors.New("not found"))

	user, err := svc.GetUser(ctx, "missing")

	assert.Error(t, err)
	assert.Nil(t, user)
	repo.AssertExpectations(t)
}

func TestUserService_UpdateUser_Success(t *testing.T) {
	repo := new(MockUserRepository)
	svc := NewUserService(repo)
	ctx := context.Background()

	user := &models.User{ID: "u1", Name: "Updated", Email: "alice@example.com"}
	repo.On("GetByID", ctx, "u1").Return(user, nil)
	repo.On("Update", ctx, user).Return(nil)

	err := svc.UpdateUser(ctx, user)

	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestUserService_UpdateUser_NotFound(t *testing.T) {
	repo := new(MockUserRepository)
	svc := NewUserService(repo)
	ctx := context.Background()

	repo.On("GetByID", ctx, "ghost").Return(nil, errors.New("not found"))

	err := svc.UpdateUser(ctx, &models.User{ID: "ghost"})

	assert.Error(t, err)
	repo.AssertExpectations(t)
}

func TestUserService_DeleteUser_Success(t *testing.T) {
	repo := new(MockUserRepository)
	svc := NewUserService(repo)
	ctx := context.Background()

	repo.On("Delete", ctx, "u1").Return(nil)

	err := svc.DeleteUser(ctx, "u1")

	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestUserService_DeleteUser_RepoError(t *testing.T) {
	repo := new(MockUserRepository)
	svc := NewUserService(repo)
	ctx := context.Background()

	repo.On("Delete", ctx, "u1").Return(errors.New("db error"))

	err := svc.DeleteUser(ctx, "u1")

	assert.EqualError(t, err, "db error")
	repo.AssertExpectations(t)
}

func TestUserService_ListUsers_DefaultPagination(t *testing.T) {
	repo := new(MockUserRepository)
	svc := NewUserService(repo)
	ctx := context.Background()

	users := []*models.User{{ID: "u1"}, {ID: "u2"}}
	// page=0 → normalised to 1, limit=0 → normalised to 20, offset=0
	repo.On("List", ctx, 20, 0).Return(users, nil)

	result, err := svc.ListUsers(ctx, 0, 0)

	assert.NoError(t, err)
	assert.Equal(t, users, result)
	repo.AssertExpectations(t)
}

func TestUserService_ListUsers_LimitCappedAt100(t *testing.T) {
	repo := new(MockUserRepository)
	svc := NewUserService(repo)
	ctx := context.Background()

	repo.On("List", ctx, 100, 0).Return([]*models.User{}, nil)

	_, err := svc.ListUsers(ctx, 1, 999)

	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestUserService_ListUsers_OffsetCalculated(t *testing.T) {
	repo := new(MockUserRepository)
	svc := NewUserService(repo)
	ctx := context.Background()

	// page=3, limit=10 → offset=20
	repo.On("List", ctx, 10, 20).Return([]*models.User{}, nil)

	_, err := svc.ListUsers(ctx, 3, 10)

	assert.NoError(t, err)
	repo.AssertExpectations(t)
}
