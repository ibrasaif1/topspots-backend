package seed

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/jackc/pgx/v5/pgxpool"
)

const createPlacesTable = `
CREATE TABLE IF NOT EXISTS places (
    place_id TEXT PRIMARY KEY,
    status TEXT NOT NULL DEFAULT 'pending',
    lat DOUBLE PRECISION,
    lng DOUBLE PRECISION,
    rating NUMERIC(2,1),
    user_rating_count INTEGER,
    google_maps_uri TEXT,
    price_level TEXT,
    display_name TEXT,
    primary_type_display_name TEXT,
    price_min INTEGER,
    price_max INTEGER,
    price_currency TEXT,
    raw_data JSONB,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);
`

const createIndexes = `
CREATE INDEX IF NOT EXISTS idx_places_status ON places(status);
CREATE INDEX IF NOT EXISTS idx_places_rating_count ON places(rating, user_rating_count);
`

func CreateSchema(ctx context.Context, pool *pgxpool.Pool) error {
	if _, err := pool.Exec(ctx, createPlacesTable); err != nil {
		return fmt.Errorf("create places table: %w", err)
	}
	if _, err := pool.Exec(ctx, createIndexes); err != nil {
		return fmt.Errorf("create indexes: %w", err)
	}
	return nil
}

// AustinPlace represents the flat structure in austin_restaurants.json
type AustinPlace struct {
	PlaceID        string `json:"place_id"`
	Name           string `json:"name"`
	Rating         float64 `json:"rating"`
	Reviews        int     `json:"reviews"`
	Address        string  `json:"address"`
	GPSCoordinates struct {
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
	} `json:"gps_coordinates"`
}

// SanDiegoFile represents the wrapper structure in san_diego_restaurants.json
type SanDiegoFile struct {
	City        string          `json:"city"`
	TotalPlaces int             `json:"totalPlaces"`
	Places      []SanDiegoPlace `json:"places"`
}

// SanDiegoPlace represents a place in san_diego_restaurants.json
type SanDiegoPlace struct {
	ID                     string  `json:"id"`
	Name                   string  `json:"name"`
	GoogleMapsUri          string  `json:"googleMapsUri"`
	PrimaryType            string  `json:"primaryType"`
	PrimaryTypeDisplayName string  `json:"primaryTypeDisplayName"`
	Rating                 float64 `json:"rating"`
	UserRatingCount        int     `json:"userRatingCount"`
	PriceLevel             string  `json:"priceLevel"`
	PriceRange             *struct {
		StartPrice struct {
			CurrencyCode string `json:"currencyCode"`
			Units        string `json:"units"`
		} `json:"startPrice"`
		EndPrice struct {
			CurrencyCode string `json:"currencyCode"`
			Units        string `json:"units"`
		} `json:"endPrice"`
	} `json:"priceRange"`
	Location struct {
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
	} `json:"location"`
}

const upsertSQL = `
INSERT INTO places (
    place_id, status, lat, lng, rating, user_rating_count,
    google_maps_uri, price_level, display_name, primary_type_display_name,
    price_min, price_max, price_currency, raw_data, updated_at
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, NOW())
ON CONFLICT (place_id) DO UPDATE SET
    status = $2, lat = $3, lng = $4, rating = $5, user_rating_count = $6,
    google_maps_uri = $7, price_level = $8, display_name = $9,
    primary_type_display_name = $10, price_min = $11, price_max = $12,
    price_currency = $13, raw_data = $14, updated_at = NOW()
`

// LoadAustinData loads austin_restaurants.json into the database
func LoadAustinData(ctx context.Context, pool *pgxpool.Pool, filePath string) (int, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return 0, fmt.Errorf("read file: %w", err)
	}

	var places []AustinPlace
	if err := json.Unmarshal(data, &places); err != nil {
		return 0, fmt.Errorf("unmarshal json: %w", err)
	}

	count := 0
	for _, p := range places {
		rawData, _ := json.Marshal(p)
		_, err := pool.Exec(ctx, upsertSQL,
			p.PlaceID,
			"complete",
			p.GPSCoordinates.Latitude,
			p.GPSCoordinates.Longitude,
			p.Rating,
			p.Reviews,
			"", // google_maps_uri not in austin data
			"", // price_level not in austin data
			p.Name,
			"Restaurant", // default type
			nil, nil, nil, // price_min, price_max, price_currency
			rawData,
		)
		if err != nil {
			return count, fmt.Errorf("insert place %s: %w", p.PlaceID, err)
		}
		count++
	}

	return count, nil
}

// LoadSanDiegoData loads san_diego_restaurants.json into the database
func LoadSanDiegoData(ctx context.Context, pool *pgxpool.Pool, filePath string) (int, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return 0, fmt.Errorf("read file: %w", err)
	}

	var file SanDiegoFile
	if err := json.Unmarshal(data, &file); err != nil {
		return 0, fmt.Errorf("unmarshal json: %w", err)
	}

	count := 0
	for _, p := range file.Places {
		rawData, _ := json.Marshal(p)

		var priceMin, priceMax *int
		var priceCurrency *string

		if p.PriceRange != nil {
			if min, err := strconv.Atoi(p.PriceRange.StartPrice.Units); err == nil {
				priceMin = &min
			}
			if max, err := strconv.Atoi(p.PriceRange.EndPrice.Units); err == nil {
				priceMax = &max
			}
			if p.PriceRange.StartPrice.CurrencyCode != "" {
				priceCurrency = &p.PriceRange.StartPrice.CurrencyCode
			}
		}

		_, err := pool.Exec(ctx, upsertSQL,
			p.ID,
			"complete",
			p.Location.Latitude,
			p.Location.Longitude,
			p.Rating,
			p.UserRatingCount,
			p.GoogleMapsUri,
			p.PriceLevel,
			p.Name,
			p.PrimaryTypeDisplayName,
			priceMin,
			priceMax,
			priceCurrency,
			rawData,
		)
		if err != nil {
			return count, fmt.Errorf("insert place %s: %w", p.ID, err)
		}
		count++
	}

	return count, nil
}
