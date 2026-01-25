package main

import (
	"context"
	"log"
	"net/http"

	"topspots-backend/internal/client"
	"topspots-backend/internal/config"
	"topspots-backend/internal/handler/api"
	"topspots-backend/internal/infra/postgres"
	"topspots-backend/internal/repository"
	"topspots-backend/internal/seed"
	"topspots-backend/internal/service"
)

func main() {
	ctx := context.Background()

	// Load config (reads .env in dev, real env vars in prod)
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	// Database
	pool, err := postgres.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	if err := seed.CreateSchema(ctx, pool); err != nil {
		log.Fatal(err)
	}

	// Dependencies
	mapsClient := client.NewMapsClient(cfg.GoogleMapsAPIKey)
	placesRepo := repository.NewPlacesRepo(pool)
	placesService := service.NewPlacesService(mapsClient, placesRepo)

	// Handlers
	countHandler := api.NewCountHandler(placesService)
	collectHandler := api.NewCollectHandler(placesService)
	placesHandler := api.NewPlacesHandler(placesRepo)

	// Routes
	http.Handle("/count", countHandler)
	http.Handle("/collect", collectHandler)
	http.Handle("/places", placesHandler)
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Start server
	log.Printf("Server starting on :%s", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, nil))
}
