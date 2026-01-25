package repository

import (
	"context"

	"topspots-backend/internal/model"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PlacesRepo struct {
	pool *pgxpool.Pool
}

func NewPlacesRepo(pool *pgxpool.Pool) *PlacesRepo {
	return &PlacesRepo{pool: pool}
}

func (r *PlacesRepo) InsertPendingPlaces(ctx context.Context, placeIDs []string) error {
	for _, id := range placeIDs {
		_, err := r.pool.Exec(ctx, `
            INSERT INTO places (place_id, status)
            VALUES ($1, 'pending')
            ON CONFLICT (place_id) DO NOTHING
        `, id)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *PlacesRepo) UpsertPlace(ctx context.Context, p *model.Place) error {
	_, err := r.pool.Exec(ctx, `
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
    `, p.PlaceID, p.Status, p.Lat, p.Lng, p.Rating, p.UserRatingCount,
		p.GoogleMapsUri, p.PriceLevel, p.DisplayName, p.PrimaryTypeDisplayName,
		p.PriceMin, p.PriceMax, p.PriceCurrency, p.RawData)
	return err
}

// GetAllPlaces returns all places with status 'complete', filtered by min rating and review count
func (r *PlacesRepo) GetAllPlaces(ctx context.Context, minRating float64, minReviews int) ([]*model.Place, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT place_id, status, lat, lng, rating, user_rating_count,
		       google_maps_uri, price_level, display_name, primary_type_display_name,
		       price_min, price_max, price_currency
		FROM places
		WHERE status = 'complete'
		  AND rating >= $1
		  AND user_rating_count >= $2
		ORDER BY user_rating_count DESC
	`, minRating, minReviews)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var places []*model.Place
	for rows.Next() {
		p := &model.Place{}
		err := rows.Scan(
			&p.PlaceID, &p.Status, &p.Lat, &p.Lng, &p.Rating, &p.UserRatingCount,
			&p.GoogleMapsUri, &p.PriceLevel, &p.DisplayName, &p.PrimaryTypeDisplayName,
			&p.PriceMin, &p.PriceMax, &p.PriceCurrency,
		)
		if err != nil {
			return nil, err
		}
		places = append(places, p)
	}

	return places, rows.Err()
}
