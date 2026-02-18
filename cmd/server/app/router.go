// Package app contains router configuration for the HTTP server.
package app

import (
	"net/http"

	"meeting-slot-service/internal/handler"
	"meeting-slot-service/internal/middleware"

	"github.com/gorilla/mux"
)

// NewRouter builds and returns the app router with all routes and
// global middleware registered.
func NewRouter(a *App) *mux.Router {
	router := mux.NewRouter()

	// Global middleware
	router.Use(middleware.Recovery)
	router.Use(middleware.Logger)
	router.Use(middleware.CORS)

	// Health check (outside the versioned API prefix)
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	}).Methods(http.MethodGet)

	// API v1 sub-router
	api := router.PathPrefix("/api/v1").Subrouter()

	registerUserRoutes(api, a.UserHandler)
	registerEventRoutes(api, a.EventHandler)
	registerAvailabilityRoutes(api, a.AvailabilityHandler)

	return router
}

func registerUserRoutes(api *mux.Router, h *handler.UserHandler) {
	api.HandleFunc("/users", h.CreateUser).Methods(http.MethodPost)
	api.HandleFunc("/users", h.ListUsers).Methods(http.MethodGet)
	api.HandleFunc("/users/{id}", h.GetUser).Methods(http.MethodGet)
	api.HandleFunc("/users/{id}", h.UpdateUser).Methods(http.MethodPut)
	api.HandleFunc("/users/{id}", h.DeleteUser).Methods(http.MethodDelete)
}

func registerEventRoutes(api *mux.Router, h *handler.EventHandler) {
	api.HandleFunc("/events", h.CreateEvent).Methods(http.MethodPost)
	api.HandleFunc("/events", h.GetEventList).Methods(http.MethodGet)
	api.HandleFunc("/events/{id}", h.GetEvent).Methods(http.MethodGet)
	api.HandleFunc("/events/{id}", h.UpdateEvent).Methods(http.MethodPut)
	api.HandleFunc("/events/{id}", h.DeleteEvent).Methods(http.MethodDelete)

	// Participants nested under events
	api.HandleFunc("/events/{id}/participants", h.AddParticipant).Methods(http.MethodPost)
	api.HandleFunc("/events/{id}/participants", h.GetParticipants).Methods(http.MethodGet)
	api.HandleFunc("/events/{id}/participants/{user_id}", h.RemoveParticipant).Methods(http.MethodDelete)
}

func registerAvailabilityRoutes(api *mux.Router, h *handler.AvailabilityHandler) {
	api.HandleFunc("/events/{id}/participants/{user_id}/availability", h.SubmitAvailability).Methods(http.MethodPost)
	api.HandleFunc("/events/{id}/participants/{user_id}/availability", h.UpdateAvailability).Methods(http.MethodPut)
	api.HandleFunc("/events/{id}/participants/{user_id}/availability", h.GetAvailability).Methods(http.MethodGet)

	// Recommendations nested under events
	api.HandleFunc("/events/{id}/recommendations", h.GetRecommendations).Methods(http.MethodGet)
}
