package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"meeting-slot-service/internal/config"
	"meeting-slot-service/internal/database"
	"meeting-slot-service/internal/handler"
	"meeting-slot-service/internal/middleware"
	"meeting-slot-service/internal/repository"
	"meeting-slot-service/internal/service"

	"github.com/gorilla/mux"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	log.Printf("Starting Meeting Slot Service on %s", cfg.Server.Address())

	// Create database (connection is lazy - will connect on first use)
	db := database.New(cfg)
	defer func(db *database.Database) {
		if err := db.Close(); err != nil {
			log.Printf("Error closing database connection: %v", err)
		}
	}(db)

	// Run migrations (this will trigger first DB connection)
	if err = db.RunMigrations(); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Initialize repositories (no DB connection yet if migrations were skipped)
	userRepo := repository.NewUserRepository(db)
	eventRepo := repository.NewEventRepository(db)
	availabilityRepo := repository.NewAvailabilityRepository(db)
	participantRepo := repository.NewParticipantRepository(db)

	// Initialize services
	userService := service.NewUserService(userRepo)
	eventService := service.NewEventService(eventRepo, userRepo, participantRepo)
	availabilityService := service.NewAvailabilityService(availabilityRepo, eventRepo, participantRepo, userRepo)
	recommendationService := service.NewRecommendationService(eventRepo, availabilityRepo, participantRepo)

	// Initialize handlers
	userHandler := handler.NewUserHandler(userService)
	eventHandler := handler.NewEventHandler(eventService)
	availabilityHandler := handler.NewAvailabilityHandler(availabilityService, recommendationService)

	// Setup router
	router := setupRouter(userHandler, eventHandler, availabilityHandler)

	// Create HTTP server
	srv := &http.Server{
		Addr:         cfg.Server.Address(),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Printf("Server listening on %s", cfg.Server.Address())
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shut down the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Graceful shutdown with 30 second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}

func setupRouter(
	userHandler *handler.UserHandler,
	eventHandler *handler.EventHandler,
	availabilityHandler *handler.AvailabilityHandler,
) *mux.Router {
	router := mux.NewRouter()

	// Apply global middleware
	router.Use(middleware.Recovery)
	router.Use(middleware.Logger)
	router.Use(middleware.CORS)

	// API v1 routes
	api := router.PathPrefix("/api/v1").Subrouter()

	// User routes
	api.HandleFunc("/users", userHandler.CreateUser).Methods(http.MethodPost)
	api.HandleFunc("/users", userHandler.ListUsers).Methods(http.MethodGet)
	api.HandleFunc("/users/{id}", userHandler.GetUser).Methods(http.MethodGet)
	api.HandleFunc("/users/{id}", userHandler.UpdateUser).Methods(http.MethodPut)
	api.HandleFunc("/users/{id}", userHandler.DeleteUser).Methods(http.MethodDelete)

	// Event routes
	api.HandleFunc("/events", eventHandler.CreateEvent).Methods(http.MethodPost)
	api.HandleFunc("/events", eventHandler.ListEvents).Methods(http.MethodGet)
	api.HandleFunc("/events/{id}", eventHandler.GetEvent).Methods(http.MethodGet)
	api.HandleFunc("/events/{id}", eventHandler.UpdateEvent).Methods(http.MethodPut)
	api.HandleFunc("/events/{id}", eventHandler.DeleteEvent).Methods(http.MethodDelete)

	// Participant routes
	api.HandleFunc("/events/{id}/participants", eventHandler.AddParticipant).Methods(http.MethodPost)
	api.HandleFunc("/events/{id}/participants", eventHandler.GetParticipants).Methods(http.MethodGet)
	api.HandleFunc("/events/{id}/participants/{user_id}", eventHandler.RemoveParticipant).Methods(http.MethodDelete)

	// Availability routes
	api.HandleFunc("/events/{id}/participants/{user_id}/availability", availabilityHandler.SubmitAvailability).Methods(http.MethodPost)
	api.HandleFunc("/events/{id}/participants/{user_id}/availability", availabilityHandler.UpdateAvailability).Methods(http.MethodPut)
	api.HandleFunc("/events/{id}/participants/{user_id}/availability", availabilityHandler.GetAvailability).Methods(http.MethodGet)

	// Recommendation routes
	api.HandleFunc("/events/{id}/recommendations", availabilityHandler.GetRecommendations).Methods(http.MethodGet)

	// Health check
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}).Methods(http.MethodGet)

	return router
}
