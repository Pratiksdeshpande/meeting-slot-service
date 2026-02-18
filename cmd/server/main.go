package main

import (
	"log"

	"meeting-slot-service/cmd/server/app"
	"meeting-slot-service/internal/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	log.Printf("Starting Meeting Slot Service on %s", cfg.Server.Address())

	application, err := app.New(cfg)
	if err != nil {
		log.Fatalf("Failed to initialise app: %v", err)
	}
	defer application.Close()

	router := app.NewRouter(application)
	server := app.NewServer(cfg.Server.Address(), router)
	server.Start()
}
