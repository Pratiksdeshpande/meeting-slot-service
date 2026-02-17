package handler

import (
	"encoding/json"
	"meeting-slot-service/internal/models"
	"meeting-slot-service/internal/service"
	"meeting-slot-service/internal/utils"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

// EventHandler handles event-related HTTP requests
type EventHandler struct {
	eventService *service.EventService
}

// NewEventHandler creates a new event handler
func NewEventHandler(eventService *service.EventService) *EventHandler {
	return &EventHandler{
		eventService: eventService,
	}
}

// CreateEvent handles POST /api/v1/events
func (h *EventHandler) CreateEvent(w http.ResponseWriter, r *http.Request) {
	var event models.Event
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		utils.WriteBadRequest(w, "Invalid request body")
		return
	}

	if err := h.eventService.CreateEvent(r.Context(), &event); err != nil {
		utils.WriteBadRequest(w, err.Error())
		return
	}

	utils.WriteSuccess(w, http.StatusCreated, event)
}

// GetEvent handles GET /api/v1/events/{id}
func (h *EventHandler) GetEvent(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	eventID := vars["id"]

	event, err := h.eventService.GetEvent(r.Context(), eventID)
	if err != nil {
		utils.WriteNotFound(w, "Event not found")
		return
	}

	utils.WriteSuccess(w, http.StatusOK, event)
}

// UpdateEvent handles PUT /api/v1/events/{id}
func (h *EventHandler) UpdateEvent(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	eventID := vars["id"]

	var event models.Event
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		utils.WriteBadRequest(w, "Invalid request body")
		return
	}

	event.ID = eventID
	if err := h.eventService.UpdateEvent(r.Context(), &event); err != nil {
		utils.WriteBadRequest(w, err.Error())
		return
	}

	utils.WriteSuccess(w, http.StatusOK, event)
}

// DeleteEvent handles DELETE /api/v1/events/{id}
func (h *EventHandler) DeleteEvent(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	eventID := vars["id"]

	if err := h.eventService.DeleteEvent(r.Context(), eventID); err != nil {
		utils.WriteNotFound(w, "Event not found")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ListEvents handles GET /api/v1/events
func (h *EventHandler) ListEvents(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	// Parse query parameters
	page, _ := strconv.Atoi(query.Get("page"))
	limit, _ := strconv.Atoi(query.Get("limit"))

	filter := models.EventFilter{
		OrganizerID: query.Get("organizer_id"),
		Status:      query.Get("status"),
		Page:        page,
		Limit:       limit,
	}

	events, total, err := h.eventService.ListEvents(r.Context(), filter)
	if err != nil {
		utils.WriteInternalError(w, "Failed to list events")
		return
	}

	utils.WritePaginatedResponse(w, events, filter.Page, filter.Limit, total)
}

// AddParticipant handles POST /api/v1/events/{id}/participants
func (h *EventHandler) AddParticipant(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	eventID := vars["id"]

	var req struct {
		UserID string `json:"user_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteBadRequest(w, "Invalid request body")
		return
	}

	if err := h.eventService.AddParticipant(r.Context(), eventID, req.UserID); err != nil {
		utils.WriteBadRequest(w, err.Error())
		return
	}

	utils.WriteSuccess(w, http.StatusCreated, map[string]string{
		"message": "Participant added successfully",
	})
}

// RemoveParticipant handles DELETE /api/v1/events/{id}/participants/{user_id}
func (h *EventHandler) RemoveParticipant(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	eventID := vars["id"]
	userID := vars["user_id"]

	if err := h.eventService.RemoveParticipant(r.Context(), eventID, userID); err != nil {
		utils.WriteNotFound(w, "Participant not found")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetParticipants handles GET /api/v1/events/{id}/participants
func (h *EventHandler) GetParticipants(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	eventID := vars["id"]

	participants, err := h.eventService.GetEventParticipants(r.Context(), eventID)
	if err != nil {
		utils.WriteInternalError(w, "Failed to get participants")
		return
	}

	utils.WriteSuccess(w, http.StatusOK, participants)
}
