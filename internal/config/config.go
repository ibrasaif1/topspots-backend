package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL      string
	GoogleMapsAPIKey string
	Port             string
}

func Load() (*Config, error) {
	// Load .env file if it exists (dev only, silently ignored in prod)
	_ = godotenv.Load()

	cfg := &Config{
		DatabaseURL:      os.Getenv("DATABASE_URL"),
		GoogleMapsAPIKey: os.Getenv("GOOGLE_MAPS_API_KEY"),
		Port:             os.Getenv("PORT"),
	}

	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL not set")
	}
	if cfg.GoogleMapsAPIKey == "" {
		return nil, fmt.Errorf("GOOGLE_MAPS_API_KEY not set")
	}
	if cfg.Port == "" {
		cfg.Port = "8080"
	}

	return cfg, nil
}
