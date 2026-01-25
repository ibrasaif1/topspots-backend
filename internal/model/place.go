package model

import (
	"encoding/json"
	"strconv"
	"time"
)

const PlaceFieldMask = "attributions,id,name,photos,addressComponents,adrFormatAddress,formattedAddress,location,plusCode,shortFormattedAddress,types,viewport,accessibilityOptions,businessStatus,containingPlaces,displayName,googleMapsLinks,googleMapsUri,iconBackgroundColor,iconMaskBaseUri,primaryType,primaryTypeDisplayName,pureServiceAreaBusiness,subDestinations,utcOffsetMinutes,currentOpeningHours,currentSecondaryOpeningHours,internationalPhoneNumber,nationalPhoneNumber,priceLevel,priceRange,rating,regularOpeningHours,regularSecondaryOpeningHours,userRatingCount,websiteUri"

type PlaceStatus string

const (
	StatusPending  PlaceStatus = "pending"
	StatusComplete PlaceStatus = "complete"
	StatusFailed   PlaceStatus = "failed"
)

// PlaceAPIResponse - for parsing Google API response
type PlaceAPIResponse struct {
	ID                     string      `json:"id"`
	Location               LatLng      `json:"location"`
	Rating                 float64     `json:"rating"`
	UserRatingCount        int         `json:"userRatingCount"`
	GoogleMapsUri          string      `json:"googleMapsUri"`
	PriceLevel             string      `json:"priceLevel"`
	DisplayName            DisplayName `json:"displayName"`
	PrimaryTypeDisplayName DisplayName `json:"primaryTypeDisplayName"`
	PriceRange             *PriceRange `json:"priceRange"`
}

type LatLng struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type DisplayName struct {
	Text         string `json:"text"`
	LanguageCode string `json:"languageCode"`
}

type PriceRange struct {
	StartPrice Price `json:"startPrice"`
	EndPrice   Price `json:"endPrice"`
}

type Price struct {
	CurrencyCode string `json:"currencyCode"`
	Units        string `json:"units"`
}

// Place - DB model with flat columns
type Place struct {
	PlaceID                string          `json:"placeId"`
	Status                 PlaceStatus     `json:"status"`
	Lat                    float64         `json:"lat"`
	Lng                    float64         `json:"lng"`
	Rating                 float64         `json:"rating"`
	UserRatingCount        int             `json:"userRatingCount"`
	GoogleMapsUri          string          `json:"googleMapsUri"`
	PriceLevel             string          `json:"priceLevel"`
	DisplayName            string          `json:"displayName"`
	PrimaryTypeDisplayName string          `json:"primaryTypeDisplayName"`
	PriceMin               *int            `json:"priceMin,omitempty"`
	PriceMax               *int            `json:"priceMax,omitempty"`
	PriceCurrency          *string         `json:"priceCurrency,omitempty"`
	RawData                json.RawMessage `json:"rawData"`
	CreatedAt              time.Time       `json:"createdAt"`
	UpdatedAt              time.Time       `json:"updatedAt"`
}

// Coordinate for API requests
type Coordinate struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

// ToPlace converts API response to DB model
func (r *PlaceAPIResponse) ToPlace(rawData json.RawMessage) *Place {
	p := &Place{
		PlaceID:                r.ID,
		Status:                 StatusComplete,
		Lat:                    r.Location.Latitude,
		Lng:                    r.Location.Longitude,
		Rating:                 r.Rating,
		UserRatingCount:        r.UserRatingCount,
		GoogleMapsUri:          r.GoogleMapsUri,
		PriceLevel:             r.PriceLevel,
		DisplayName:            r.DisplayName.Text,
		PrimaryTypeDisplayName: r.PrimaryTypeDisplayName.Text,
		RawData:                rawData,
	}
	if r.PriceRange != nil {
		min, _ := strconv.Atoi(r.PriceRange.StartPrice.Units)
		max, _ := strconv.Atoi(r.PriceRange.EndPrice.Units)
		p.PriceMin = &min
		p.PriceMax = &max
		p.PriceCurrency = &r.PriceRange.StartPrice.CurrencyCode
	}
	return p
}
