package service

import (
	"context"
	"encoding/json"
	"log"

	"topspots-backend/internal/client"
	"topspots-backend/internal/model"
	"topspots-backend/internal/repository"
)

type PlacesService struct {
	client *client.MapsClient
	repo   *repository.PlacesRepo
}

func NewPlacesService(client *client.MapsClient, repo *repository.PlacesRepo) *PlacesService {
	return &PlacesService{client: client, repo: repo}
}

// Count - simple passthrough to Maps API
func (s *PlacesService) Count(coords []model.Coordinate, types []string) (int, error) {
	return s.client.GetCount(coords, types)
}

// CollectAndHydrate - orchestrates the full flow
func (s *PlacesService) CollectAndHydrate(ctx context.Context, coords []model.Coordinate, types []string) ([]*model.Place, error) {
	// Step 1: Get place IDs from Aggregate API
	placeIDs, err := s.client.GetInsightPlaces(coords, types)
	if err != nil {
		return nil, err
	}

	// Step 2: Async insert as pending (fire and forget)
	go func() {
		if err := s.repo.InsertPendingPlaces(context.Background(), placeIDs); err != nil {
			log.Printf("Warning: failed to insert pending places: %v", err)
		}
	}()

	// Step 3: Hydrate each place
	places := make([]*model.Place, 0, len(placeIDs))
	for _, id := range placeIDs {
		rawData, err := s.client.GetPlaceDetails(id)
		if err != nil {
			log.Printf("Warning: failed to get details for place %s: %v", id, err)
			continue
		}

		var apiResp model.PlaceAPIResponse
		if err := json.Unmarshal(rawData, &apiResp); err != nil {
			log.Printf("Warning: failed to unmarshal place %s: %v", id, err)
			continue
		}

		place := apiResp.ToPlace(rawData)

		if err := s.repo.UpsertPlace(ctx, place); err != nil {
			log.Printf("Warning: failed to upsert place %s: %v", id, err)
			// Still return the place even if DB write fails
		}

		places = append(places, place)
	}

	return places, nil
}
