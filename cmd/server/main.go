package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"topspots-backend/internal/infra/postgres"
	"topspots-backend/internal/seed"
)

func main() {
	// Flags
	austinFile := flag.String("austin", "", "Path to austin_restaurants.json")
	sanDiegoFile := flag.String("sandiego", "", "Path to san_diego_restaurants.json")
	dbURL := flag.String("db", "", "Database URL (overrides DATABASE_URL env var)")
	flag.Parse()

	// Load .env if present
	_ = godotenv.Load()

	// Determine database URL
	databaseURL := *dbURL
	if databaseURL == "" {
		databaseURL = os.Getenv("DATABASE_URL")
	}
	if databaseURL == "" {
		log.Fatal("DATABASE_URL not set. Use -db flag or set DATABASE_URL env var")
	}

	ctx := context.Background()

	// Connect to database
	pool, err := postgres.NewPool(ctx, databaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	// Create schema if it doesn't exist
	if err := seed.CreateSchema(ctx, pool); err != nil {
		log.Fatalf("Failed to create schema: %v", err)
	}
	fmt.Println("Schema created/verified")

	totalCount := 0

	// Load Austin data
	if *austinFile != "" {
		count, err := seed.LoadAustinData(ctx, pool, *austinFile)
		if err != nil {
			log.Fatalf("Failed to load Austin data: %v", err)
		}
		fmt.Printf("Loaded %d places from Austin\n", count)
		totalCount += count
	}

	// Load San Diego data
	if *sanDiegoFile != "" {
		count, err := seed.LoadSanDiegoData(ctx, pool, *sanDiegoFile)
		if err != nil {
			log.Fatalf("Failed to load San Diego data: %v", err)
		}
		fmt.Printf("Loaded %d places from San Diego\n", count)
		totalCount += count
	}

	if totalCount == 0 {
		fmt.Println("No data files specified. Use -austin and/or -sandiego flags.")
		fmt.Println("Example:")
		fmt.Println("  go run ./cmd/seed -austin=/path/to/austin_restaurants.json -sandiego=/path/to/san_diego_restaurants.json")
	} else {
		fmt.Printf("\nTotal: %d places loaded into database\n", totalCount)
	}
}
