package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"topspots-backend/internal/repository"
)

type PlacesHandler struct {
	repo *repository.PlacesRepo
}

func NewPlacesHandler(repo *repository.PlacesRepo) *PlacesHandler {
	return &PlacesHandler{repo: repo}
}

type PlaceResponse struct {
	PlaceID                string   `json:"place_id"`
	Name                   string   `json:"name"`
	Rating                 float64  `json:"rating"`
	Reviews                int      `json:"reviews"`
	Lat                    float64  `json:"lat"`
	Lng                    float64  `json:"lng"`
	GoogleMapsUri          string   `json:"googleMapsUri,omitempty"`
	PriceLevel             string   `json:"priceLevel,omitempty"`
	PrimaryTypeDisplayName string   `json:"cuisine,omitempty"`
	PriceMin               *int     `json:"priceMin,omitempty"`
	PriceMax               *int     `json:"priceMax,omitempty"`
}

func (h *PlacesHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse query params with defaults
	minRating := 4.5
	minReviews := 1000

	if v := r.URL.Query().Get("minRating"); v != "" {
		if parsed, err := strconv.ParseFloat(v, 64); err == nil {
			minRating = parsed
		}
	}
	if v := r.URL.Query().Get("minReviews"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil {
			minReviews = parsed
		}
	}

	places, err := h.repo.GetAllPlaces(r.Context(), minRating, minReviews)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Convert to response format
	response := make([]PlaceResponse, 0, len(places))
	for _, p := range places {
		response = append(response, PlaceResponse{
			PlaceID:                p.PlaceID,
			Name:                   p.DisplayName,
			Rating:                 p.Rating,
			Reviews:                p.UserRatingCount,
			Lat:                    p.Lat,
			Lng:                    p.Lng,
			GoogleMapsUri:          p.GoogleMapsUri,
			PriceLevel:             p.PriceLevel,
			PrimaryTypeDisplayName: p.PrimaryTypeDisplayName,
			PriceMin:               p.PriceMin,
			PriceMax:               p.PriceMax,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
