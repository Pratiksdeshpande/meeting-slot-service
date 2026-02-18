// Package app wires together all app dependencies.
package app

import (
	"log"

	"meeting-slot-service/internal/config"
	"meeting-slot-service/internal/database"
	"meeting-slot-service/internal/handler"
	"meeting-slot-service/internal/repository"
	"meeting-slot-service/internal/service"
)

// App holds the fully-wired handler dependencies and the database connection
// so that callers can close resources on shutdown.
type App struct {
	DB                  *database.Database
	UserHandler         *handler.UserHandler
	EventHandler        *handler.EventHandler
	AvailabilityHandler *handler.AvailabilityHandler
}

// New initialises the database, runs migrations, and wires all layers
// (repository → service → handler).  The caller is responsible for closing
// App.DB when the process exits.
func New(cfg *config.Config) (*App, error) {
	db := database.New(cfg)

	if err := db.RunMigrations(); err != nil {
		_ = db.Close()
		return nil, err
	}

	log.Println("Database migrations completed successfully")

	// Repositories
	userRepo := repository.NewUserRepository(db)
	eventRepo := repository.NewEventRepository(db)
	availabilityRepo := repository.NewAvailabilityRepository(db)
	participantRepo := repository.NewParticipantRepository(db)

	// Services
	userService := service.NewUserService(userRepo)
	eventService := service.NewEventService(eventRepo, userRepo, participantRepo)
	availabilityService := service.NewAvailabilityService(availabilityRepo, eventRepo, participantRepo, userRepo)
	recommendationService := service.NewRecommendationService(eventRepo, availabilityRepo, participantRepo)

	// Handlers
	userHandler := handler.NewUserHandler(userService)
	eventHandler := handler.NewEventHandler(eventService)
	availabilityHandler := handler.NewAvailabilityHandler(availabilityService, recommendationService)

	return &App{
		DB:                  db,
		UserHandler:         userHandler,
		EventHandler:        eventHandler,
		AvailabilityHandler: availabilityHandler,
	}, nil
}

// Close releases all resources owned by the app.
func (a *App) Close() {
	if err := a.DB.Close(); err != nil {
		log.Printf("Error closing database connection: %v", err)
	}
}
