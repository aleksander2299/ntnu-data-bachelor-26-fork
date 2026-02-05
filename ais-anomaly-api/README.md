# AIS Anomaly API

A read-only REST API for accessing AIS anomaly data, built with [Go Fiber](https://gofiber.io/).

## Features

- 🚀 Fast and lightweight API built with Fiber
- 📊 Query anomaly groups by date range
- 🔍 Get individual anomaly groups and their anomalies
- � **Interactive Swagger/OpenAPI documentation**
- �🐳 Docker support for easy deployment
- 🏥 Health check endpoint

## 📚 API Documentation

**Interactive Swagger UI is available at:**

### `http://localhost:3000/swagger/index.html`

The Swagger documentation provides:
- Complete endpoint documentation
- Request/response schemas with examples
- Try-it-out functionality to test endpoints
- Model definitions

### Regenerating Swagger Docs

After modifying API endpoints or models:

```bash
cd ais-anomaly-api
swag init
# or
/usr/local/go/bin/swag init
```

## API Endpoints

### Health Check
```
GET /api/v1/health
```

### Get Anomaly Groups
```
GET /api/v1/anomaly-groups?start_date=YYYY-MM-DD&end_date=YYYY-MM-DD
```

Query parameters:
- `start_date` (optional): Start date filter (format: `YYYY-MM-DD` or RFC3339)
- `end_date` (optional): End date filter (format: `YYYY-MM-DD` or RFC3339)

If no dates are provided, defaults to the last 30 days.

**Response:**
```json
{
  "data": [
    {
      "id": 1,
      "type": "speed_anomaly",
      "mmsi": 123456789,
      "started_at": "2026-02-01T10:00:00Z",
      "last_activity_at": "2026-02-01T12:00:00Z",
      "latitude": 59.9139,
      "longitude": 10.7522
    }
  ],
  "total_count": 1,
  "start_date": "2026-01-04T00:00:00Z",
  "end_date": "2026-02-04T00:00:00Z"
}
```

### Get Anomaly Group by ID
```
GET /api/v1/anomaly-groups/:id
```

**Response:**
```json
{
  "id": 1,
  "type": "speed_anomaly",
  "mmsi": 123456789,
  "started_at": "2026-02-01T10:00:00Z",
  "last_activity_at": "2026-02-01T12:00:00Z",
  "latitude": 59.9139,
  "longitude": 10.7522
}
```

### Get Anomalies by Group ID
```
GET /api/v1/anomaly-groups/:id/anomalies
```

**Response:**
```json
{
  "anomaly_group_id": 1,
  "anomalies": [
    {
      "id": 1,
      "type": "speed_anomaly",
      "metadata": {...},
      "created_at": "2026-02-01T10:30:00Z",
      "mmsi": 123456789,
      "anomaly_group_id": 1,
      "data_source": "SYNTHETIC"
    }
  ],
  "total_count": 1
}
```

## Running Locally

### Prerequisites
- Go 1.25.1 or later
- PostgreSQL with PostGIS extension
- Database set up with the required schema

### Environment Variables

| Variable    | Default          | Description            |
|-------------|------------------|------------------------|
| DB_HOST     | localhost        | PostgreSQL host        |
| DB_PORT     | 5439             | PostgreSQL port        |
| DB_USER     | postgres         | Database user          |
| DB_PASSWORD | birdsarentreal   | Database password      |
| DB_NAME     | ais              | Database name          |
| PORT        | 3000             | API server port        |

### Run with Go

```bash
cd ais-anomaly-api

# Download dependencies
go mod tidy

# Run the application
go run .
```

### Run with Docker

```bash
cd ais-anomaly-api

# Build the image
docker build -t ais-anomaly-api .

# Run the container
docker run -p 3000:3000 \
  -e DB_HOST=host.docker.internal \
  -e DB_PORT=5439 \
  ais-anomaly-api
```

### Run with Docker Compose

From the project root:

```bash
docker compose up ais-anomaly-api
```

## Example Usage

```bash
# Get all anomaly groups from the last 30 days
curl http://localhost:3000/api/v1/anomaly-groups

# Get anomaly groups for a specific date range
curl "http://localhost:3000/api/v1/anomaly-groups?start_date=2026-01-01&end_date=2026-02-01"

# Get a specific anomaly group
curl http://localhost:3000/api/v1/anomaly-groups/1

# Get anomalies for a specific group
curl http://localhost:3000/api/v1/anomaly-groups/1/anomalies

# Health check
curl http://localhost:3000/api/v1/health
```

## Project Structure

```
ais-anomaly-api/
├── main.go              # Application entry point
├── Dockerfile           # Docker configuration
├── go.mod               # Go module definition
├── go.sum               # Go dependencies checksum
├── README.md            # This file
├── db/
│   └── connection.go    # Database connection logic
├── handlers/
│   └── anomaly_handler.go  # HTTP request handlers
└── models/
    └── anomaly.go       # Data models
```

## Vi er i hamn, Sam! 🚢
