package handler

import (
	"encoding/json"
	"meeting-slot-service/internal/models"
	"meeting-slot-service/internal/service"
	"meeting-slot-service/internal/utils"
	"net/http"

	"github.com/gorilla/mux"
)

// AvailabilityHandler handles availability-related HTTP requests
type AvailabilityHandler struct {
	availabilityService   *service.AvailabilityService
	recommendationService *service.RecommendationService
}

// NewAvailabilityHandler creates a new availability handler
func NewAvailabilityHandler(
	availabilityService *service.AvailabilityService,
	recommendationService *service.RecommendationService,
) *AvailabilityHandler {
	return &AvailabilityHandler{
		availabilityService:   availabilityService,
		recommendationService: recommendationService,
	}
}

// SubmitAvailability handles POST /api/v1/events/{id}/participants/{user_id}/availability
func (h *AvailabilityHandler) SubmitAvailability(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	eventID := vars["id"]
	userID := vars["user_id"]

	var req struct {
		AvailableSlots []models.AvailabilitySlot `json:"available_slots"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteBadRequest(w, "Invalid request body")
		return
	}

	if err := h.availabilityService.SubmitAvailability(r.Context(), eventID, userID, req.AvailableSlots); err != nil {
		utils.WriteBadRequest(w, err.Error())
		return
	}

	utils.WriteSuccess(w, http.StatusOK, map[string]string{
		"message": "Availability submitted successfully",
	})
}

// UpdateAvailability handles PUT /api/v1/events/{id}/participants/{user_id}/availability
func (h *AvailabilityHandler) UpdateAvailability(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	eventID := vars["id"]
	userID := vars["user_id"]

	var req struct {
		AvailableSlots []models.AvailabilitySlot `json:"available_slots"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteBadRequest(w, "Invalid request body")
		return
	}

	if err := h.availabilityService.UpdateAvailability(r.Context(), eventID, userID, req.AvailableSlots); err != nil {
		utils.WriteBadRequest(w, err.Error())
		return
	}

	utils.WriteSuccess(w, http.StatusOK, map[string]string{
		"message": "Availability updated successfully",
	})
}

// GetAvailability handles GET /api/v1/events/{id}/participants/{user_id}/availability
func (h *AvailabilityHandler) GetAvailability(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	eventID := vars["id"]
	userID := vars["user_id"]

	slots, err := h.availabilityService.GetAvailability(r.Context(), eventID, userID)
	if err != nil {
		utils.WriteInternalError(w, "Failed to get availability")
		return
	}

	// Return empty array instead of null when no availability
	if slots == nil {
		slots = []models.AvailabilitySlot{}
	}

	utils.WriteSuccess(w, http.StatusOK, slots)
}

// GetRecommendations handles GET /api/v1/events/{id}/recommendations
func (h *AvailabilityHandler) GetRecommendations(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	eventID := vars["id"]

	recommendations, err := h.recommendationService.GetRecommendations(r.Context(), eventID)
	if err != nil {
		utils.WriteInternalError(w, err.Error())
		return
	}

	utils.WriteSuccess(w, http.StatusOK, recommendations)
}
