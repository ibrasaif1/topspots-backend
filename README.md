# TopSpots Backend

A Go backend service that integrates with Google Maps APIs to provide place insights and details.

## Endpoints

| Method | Path | Description |
|--------|------|-------------|
| POST | `/count` | Get place count in polygon via Aggregate API (INSIGHT_COUNT) |
| POST | `/collect` | Get place IDs in polygon via Aggregate API (INSIGHT_PLACES) |
| POST | `/hydrate` | Get place details (rating, userRatingCount, etc.) via Places API |
| GET | `/health` | Health check endpoint |

## Setup

### Prerequisites

- Go 1.21 or later
- Google Maps API key with access to:
  - Area Insights API
  - Places API (New)

### Environment Variables

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `GOOGLE_MAPS_API_KEY` | Yes | - | Your Google Maps API key |
| `PORT` | No | 8080 | Server port |

### Running Locally

```bash
export GOOGLE_MAPS_API_KEY="your-api-key"
go run main.go
```

### Building

```bash
go build -o topspots-backend
./topspots-backend
```

## API Usage

### Count Endpoint

Get the count of places within a polygon.

```bash
curl -X POST http://localhost:8080/count \
  -H "Content-Type: application/json" \
  -d '{
    "polygon": [
      {"lat": 37.7749, "lng": -122.4194},
      {"lat": 37.7849, "lng": -122.4194},
      {"lat": 37.7849, "lng": -122.4094},
      {"lat": 37.7749, "lng": -122.4094}
    ],
    "includedTypes": ["restaurant", "cafe"]
  }'
```

Response:
```json
{
  "count": 42
}
```

### Collect Endpoint

Get place IDs within a polygon.

```bash
curl -X POST http://localhost:8080/collect \
  -H "Content-Type: application/json" \
  -d '{
    "polygon": [
      {"lat": 37.7749, "lng": -122.4194},
      {"lat": 37.7849, "lng": -122.4194},
      {"lat": 37.7849, "lng": -122.4094},
      {"lat": 37.7749, "lng": -122.4094}
    ],
    "includedTypes": ["restaurant"]
  }'
```

Response:
```json
{
  "placeIds": ["ChIJN1t_tDeuEmsRUsoyG83frY4", "ChIJP3Sa8ziYEmsRUKgyFmh9AQM"]
}
```

### Hydrate Endpoint

Get details for specific place IDs.

```bash
curl -X POST http://localhost:8080/hydrate \
  -H "Content-Type: application/json" \
  -d '{
    "placeIds": ["ChIJN1t_tDeuEmsRUsoyG83frY4"]
  }'
```

Response:
```json
{
  "places": [
    {
      "placeId": "ChIJN1t_tDeuEmsRUsoyG83frY4",
      "name": "Example Restaurant",
      "rating": 4.5,
      "userRatingCount": 120
    }
  ]
}
```

## Deployment

The service is designed to be deployed as a container or on any platform that supports Go applications.

### Docker

```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o topspots-backend

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/topspots-backend .
EXPOSE 8080
CMD ["./topspots-backend"]
```

Build and run:
```bash
docker build -t topspots-backend .
docker run -p 8080:8080 -e GOOGLE_MAPS_API_KEY="your-key" topspots-backend
```

