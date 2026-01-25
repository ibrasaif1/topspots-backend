package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"topspots-backend/internal/model"
)

const (
	aggregateAPIURL    = "https://areainsights.googleapis.com/v1:computeInsights"
	placeDetailsAPIURL = "https://places.googleapis.com/v1/places"
)

type MapsClient struct {
	apiKey     string
	httpClient *http.Client
}

func NewMapsClient(apiKey string) *MapsClient {
	return &MapsClient{
		apiKey:     apiKey,
		httpClient: &http.Client{},
	}
}

// aggregateRequest represents the request body for the Aggregate API
type aggregateRequest struct {
	Insights []string        `json:"insights"`
	Filter   aggregateFilter `json:"filter"`
}

type aggregateFilter struct {
	LocationFilter locationFilter `json:"locationFilter"`
	TypeFilter     typeFilter     `json:"typeFilter"`
}

type locationFilter struct {
	CustomArea customArea `json:"customArea"`
}

type customArea struct {
	Polygon polygon `json:"polygon"`
}

type polygon struct {
	Coordinates []coordinate `json:"coordinates"`
}

type coordinate struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type typeFilter struct {
	IncludedTypes []string `json:"includedTypes"`
}

// aggregateResponse represents the response from the Aggregate API
type aggregateResponse struct {
	Count    int      `json:"count"`
	PlaceIDs []string `json:"placeIds"`
}

// GetCount calls the Aggregate API with INSIGHT_COUNT
func (c *MapsClient) GetCount(coords []model.Coordinate, includedTypes []string) (int, error) {
	resp, err := c.callAggregateAPI("INSIGHT_COUNT", coords, includedTypes)
	if err != nil {
		return 0, err
	}
	return resp.Count, nil
}

// GetInsightPlaces calls the Aggregate API with INSIGHT_PLACES
func (c *MapsClient) GetInsightPlaces(coords []model.Coordinate, includedTypes []string) ([]string, error) {
	resp, err := c.callAggregateAPI("INSIGHT_PLACES", coords, includedTypes)
	if err != nil {
		return nil, err
	}
	return resp.PlaceIDs, nil
}

func (c *MapsClient) callAggregateAPI(insightType string, coords []model.Coordinate, includedTypes []string) (*aggregateResponse, error) {
	// Convert coordinates to API format
	apiCoords := make([]coordinate, len(coords))
	for i, coord := range coords {
		apiCoords[i] = coordinate{
			Latitude:  coord.Lat,
			Longitude: coord.Lng,
		}
	}

	reqBody := aggregateRequest{
		Insights: []string{insightType},
		Filter: aggregateFilter{
			LocationFilter: locationFilter{
				CustomArea: customArea{
					Polygon: polygon{
						Coordinates: apiCoords,
					},
				},
			},
			TypeFilter: typeFilter{
				IncludedTypes: includedTypes,
			},
		},
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", aggregateAPIURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Goog-Api-Key", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	var result aggregateResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &result, nil
}

// GetPlaceDetails calls Place Details API and returns raw JSON bytes
func (c *MapsClient) GetPlaceDetails(placeID string) ([]byte, error) {
	url := fmt.Sprintf("%s/%s", placeDetailsAPIURL, placeID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Goog-Api-Key", c.apiKey)
	req.Header.Set("X-Goog-FieldMask", model.PlaceFieldMask)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	return body, nil
}
