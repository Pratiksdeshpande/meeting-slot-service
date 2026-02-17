package handler

import (
	"encoding/json"
	"meeting-slot-service/internal/models"
	"meeting-slot-service/internal/service"
	"meeting-slot-service/internal/utils"
	"net/http"

	"github.com/gorilla/mux"
)

// UserHandler handles user-related HTTP requests
type UserHandler struct {
	userService *service.UserService
}

// NewUserHandler creates a new user handler
func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// CreateUser handles POST /api/v1/users
func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var user models.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		utils.WriteBadRequest(w, "Invalid request body")
		return
	}

	if err := h.userService.CreateUser(r.Context(), &user); err != nil {
		utils.WriteBadRequest(w, err.Error())
		return
	}

	utils.WriteSuccess(w, http.StatusCreated, user)
}

// GetUser handles GET /api/v1/users/{id}
func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["id"]

	user, err := h.userService.GetUser(r.Context(), userID)
	if err != nil {
		utils.WriteNotFound(w, "User not found")
		return
	}

	utils.WriteSuccess(w, http.StatusOK, user)
}

// UpdateUser handles PUT /api/v1/users/{id}
func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["id"]

	var user models.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		utils.WriteBadRequest(w, "Invalid request body")
		return
	}

	user.ID = userID
	if err := h.userService.UpdateUser(r.Context(), &user); err != nil {
		utils.WriteBadRequest(w, err.Error())
		return
	}

	utils.WriteSuccess(w, http.StatusOK, user)
}

// DeleteUser handles DELETE /api/v1/users/{id}
func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["id"]

	if err := h.userService.DeleteUser(r.Context(), userID); err != nil {
		utils.WriteNotFound(w, "User not found")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
