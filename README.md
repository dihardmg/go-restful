# Go REST API

A complete REST API built with Go, featuring Product CRUD operations, PostgreSQL database, Swagger UI documentation, and Docker Compose orchestration.

## Tech Stack

- **Go**: 1.25.6
- **Web Framework**: Gin
- **ORM**: GORM
- **Database**: PostgreSQL
- **API Documentation**: Swagger UI
- **Containerization**: Docker Compose

## Features

- Product CRUD operations (Create, Read, Update, Delete)
- RESTful API design
- PostgreSQL database with GORM ORM
- Swagger UI documentation
- Docker Compose for easy deployment
- Unit tests with mock repository
- Environment-based configuration
- Input validation
- Pagination support
- **Request tracing with unique trace ID for debugging and monitoring**


## Getting Started

### Prerequisites

- Go 1.25.6 or higher
- Docker and Docker Compose
- PostgreSQL (if running locally)

### Running with Docker Compose (Recommended)

1. Clone the repository
2. Copy environment template:
   ```bash
   cp .env.example .env
   ```
3. Start the services:
   ```bash
   docker-compose up --build
   ```
4. Access the API at `http://localhost:8080`
5. Access Swagger UI at `http://localhost:8080/swagger/index.html`

### Running Locally

1. Install dependencies:
   ```bash
   go mod download
   ```

2. Set up PostgreSQL database

3. Configure environment variables (copy `.env.example` to `.env` and update)

4. Run the application:
   ```bash
   go run cmd/server/main.go
   ```

## API Endpoints

### Products

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/products` | Create a new product |
| GET | `/api/v1/products` | Get all products (with pagination) |
| GET | `/api/v1/products/:id` | Get a product by ID |
| PUT | `/api/v1/products/:id` | Update a product |
| DELETE | `/api/v1/products/:id` | Delete a product |

### Health Check

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/ping` | Health check endpoint |

## Example Usage

### Create a Product

```bash
curl -X POST http://localhost:8080/api/v1/products \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Gaming Laptop",
    "description": "High-performance gaming laptop",
    "price": 1500.00,
    "stock": 10
  }'
```

### Get All Products

```bash
curl http://localhost:8080/api/v1/products?page=1&limit=10
```

### Get a Product by ID

```bash
curl http://localhost:8080/api/v1/products/1
```

### Update a Product

```bash
curl -X PUT http://localhost:8080/api/v1/products/1 \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Updated Gaming Laptop",
    "description": "High-end gaming laptop",
    "price": 1800.00,
    "stock": 5
  }'
```

### Delete a Product

```bash
curl -X DELETE http://localhost:8080/api/v1/products/1
```

### Example API Responses with Trace ID

#### Success Response (Single Product)
```json
{
  "data": {
    "id": 2,
    "name": "Gaming Laptop",
    "description": "High-performance gaming laptop",
    "price": 1500,
    "stock": 10,
    "created_at": "2026-01-28T13:35:27.813525Z",
    "updated_at": null
  },
  "meta": {
    "trace_id": "e6b70517-9537-4e07-8ae1-ff400c027553"
  }
}
```

**Note**: `updated_at` is `null` when product is first created. It will contain a timestamp only when the product is updated.

#### Success Response (Paginated)
```json
{
  "data": [
    {
      "id": 2,
      "name": "Gaming Laptop",
      "price": 1500,
      "stock": 10,
      "created_at": "2026-01-28T13:35:27.813525Z",
      "updated_at": "2026-01-28T14:30:15.123456Z"
    }
  ],
  "meta": {
    "trace_id": "e6b70517-9537-4e07-8ae1-ff400c027553",
    "total": 14,
    "page": 1,
    "limit": 10,
    "total_pages": 2
  }
}
```

**Note**: `updated_at` will be:
- `null` for newly created products (never updated)
- Timestamp for products that have been updated at least once

#### Error Response
```json
{
  "error": {
    "code": "NOT_FOUND",
    "message": "Product not found"
  },
  "meta": {
    "timestamp": "2026-02-02T04:54:12.123456789Z",
    "trace_id": "e6b70517-9537-4e07-8ae1-ff400c027553"
  }
}
```

#### Validation Error Response (422)
When request validation fails, the API returns field-level errors in the `details` field:

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid request data"
  },
  "details": {
    "name": ["min length is 3"],
    "price": ["must be greater than 0"],
    "stock": ["must not be negative"]
  },
  "meta": {
    "timestamp": "2026-02-02T04:54:12.123456789Z",
    "trace_id": "e6b70517-9537-4e07-8ae1-ff400c027553"
  }
}
```

**Validation Rules:**
- `name`: Required, minimum 3 characters
- `price`: Required, must be a number, must be greater than 0 (cannot be 0 or negative)
- `stock`: Required, must be a number, cannot be negative (can be 0)

## Running Tests

Run all tests:
```bash
go test ./... -v
```

Run tests with coverage:
```bash
go test ./... -cover
```

## Configuration

The application uses environment variables for configuration. See `.env.example` for available options:

- `SERVER_PORT`: Server port (default: 8080)
- `DB_HOST`: Database host (default: localhost)
- `DB_PORT`: Database port (default: 5432)
- `DB_USER`: Database user (default: postgres)
- `DB_PASSWORD`: Database password (default: postgres)
- `DB_NAME`: Database name (default: go_rest_db)
- `DB_SSLMODE`: SSL mode (default: disable)

## Swagger Documentation

Access the Swagger UI documentation at:
```
http://localhost:8080/swagger/index.html
```

![Swagger UI Screenshot](img/swagger.png)

This provides an interactive API documentation where you can test all endpoints with:
- Request duration display
- Syntax highlighting
- Try it out functionality
- Request/response examples

## Request Tracing & Debugging with Trace ID

Every API request includes a unique `trace_id` that helps you track and debug requests through the system. This is essential for monitoring, troubleshooting, and log analysis.

### What is Trace ID?

A `trace_id` is a UUID v4 (e.g., `e6b70517-9537-4e07-8ae1-ff400c027553`) that uniquely identifies each request throughout its lifecycle - from the moment it hits the API until the response is sent.

### Where to Find Trace ID?

#### 1. In Response Headers
Every HTTP response includes the trace ID in the headers:

```bash
curl -i http://localhost:8080/api/v1/products/999
```

Response headers:
```
HTTP/1.1 404 Not Found
Content-Type: application/json
X-Trace-ID: e6b70517-9537-4e07-8ae1-ff400c027553
```

#### 2. In Response Body (Success)
```json
{
  "data": {
    "id": 2,
    "name": "Gaming Laptop",
    "price": 1500,
    "stock": 10,
    "created_at": "2026-01-28T13:35:27.813525Z",
    "updated_at": "2026-01-28T13:35:27.813525Z"
  },
  "meta": {
    "trace_id": "e6b70517-9537-4e07-8ae1-ff400c027553"
  }
}
```

#### 3. In Response Body (Error)
```json
{
  "error": {
    "code": "NOT_FOUND",
    "message": "Product not found"
  },
  "meta": {
    "timestamp": "2026-02-02T04:54:12.123456789Z",
    "trace_id": "e6b70517-9537-4e07-8ae1-ff400c027553"
  }
}
```

#### 4. In Response Body (Paginated)
```json
{
  "data": [...],
  "meta": {
    "trace_id": "e6b70517-9537-4e07-8ae1-ff400c027553",
    "total": 14,
    "page": 1,
    "limit": 10,
    "total_pages": 2
  }
}
```

### Server Logs Format

All requests are automatically logged with the following format:

```
[trace_id] timestamp | status | duration | client_ip | method | path
```

Example:
```
[e6b70517-9537-4e07-8ae1-ff400c027553] 2026/02/02 - 04:54:12 | 404 |    1.2ms |      172.18.0.1 | GET      /api/v1/products/999
```

### CLI Commands for Tracing

#### Search for a Specific Trace ID

```bash
# Basic search
docker-compose logs api | grep "e6b70517-9537-4e07-8ae1-ff400c027553"

# Or using docker logs
docker logs go-rest-api | grep "e6b70517-9537-4e07-8ae1-ff400c027553"
```

#### Search with Context

```bash
# Show 5 lines before and after the match
docker-compose logs api | grep -B 5 -A 5 "e6b70517-9537-4e07-8ae1-ff400c027553"

# Show 10 lines before and after
docker-compose logs api | grep -B 10 -A 10 "e6b70517-9537-4e07-8ae1-ff400c027553"
```

#### Real-time Monitoring for Specific Trace ID

```bash
# Follow logs and filter by trace ID
docker-compose logs -f api | grep --line-buffered "e6b70517-9537-4e07-8ae1-ff400c027553"
```

#### Search with Color Highlight

```bash
# Highlight the trace ID in output
docker-compose logs api | grep --color=always "e6b70517-9537-4e07-8ae1-ff400c027553"
```

#### Search Recent Logs

```bash
# Search in last 100 lines
docker-compose logs --tail=100 api | grep "e6b70517-9537-4e07-8ae1-ff400c027553"

# Search in last 500 lines
docker-compose logs --tail=500 api | grep "e6b70517-9537-4e07-8ae1-ff400c027553"
```

#### Filter Only Errors

```bash
# Search trace ID and show only errors (4xx and 5xx)
docker-compose logs api | grep "e6b70517-9537-4e07-8ae1-ff400c027553" | grep -E " (4|5)\d{2}"
```

#### Save Search Results to File

```bash
# Save to file
docker-compose logs api | grep "e6b70517-9537-4e07-8ae1-ff400c027553" > trace_search.log

# View the file
cat trace_search.log
```

#### Search Multiple Trace IDs

```bash
# Search for multiple trace IDs at once
docker-compose logs api | grep -E "e6b70517-9537-4e07-8ae1-ff400c027553|another-trace-id-here|yet-another-id"
```

#### Real-time Logs with Timestamps

```bash
# Follow logs with timestamps and filter by trace ID
docker-compose logs -f --timestamps api | grep "e6b70517-9537-4e07-8ae1-ff400c027553"
```

### PowerShell Commands (Windows)

```powershell
# Basic search
docker-compose logs api | Select-String "e6b70517-9537-4e07-8ae1-ff400c027553"

# Search with context
docker-compose logs api | Select-String -Context 5,5 "e6b70517-9537-4e07-8ae1-ff400c027553"

# Real-time monitoring
docker-compose logs -f api | Select-String "e6b70517-9537-4e07-8ae1-ff400c027553"
```

### Troubleshooting Workflow

When you encounter an error, follow these steps:

#### Step 1: Get the Trace ID from Response

```bash
curl -X GET http://localhost:8080/api/v1/products/999
```

Extract the `trace_id` from the response:
```json
{
  "error": {
    "code": "NOT_FOUND",
    "message": "Product not found"
  },
  "meta": {
    "trace_id": "e6b70517-9537-4e07-8ae1-ff400c027553"
  }
}
```

#### Step 2: Search in Server Logs

```bash
docker-compose logs api | grep "e6b70517-9537-4e07-8ae1-ff400c027553"
```

#### Step 3: Analyze the Log Entry

```
[e6b70517-9537-4e07-8ae1-ff400c027553] 2026/02/02 - 04:54:12 | 404 |    1.2ms |      172.18.0.1 | GET      /api/v1/products/999
```

From this log, you can see:
- **Timestamp**: When the request occurred
- **Status**: 404 (Not Found)
- **Duration**: 1.2ms (Response time)
- **Client IP**: 172.18.0.1 (Who made the request)
- **Method**: GET
- **Path**: /api/v1/products/999 (What was requested)

### Advanced Log Analysis

#### Count Requests by Status Code

```bash
# Count 404 errors
docker-compose logs api | grep " 404 " | wc -l

# Count 500 errors
docker-compose logs api | grep " 500 " | wc -l

# Count all errors (4xx and 5xx)
docker-compose logs api | grep -E " (4|5)\d{2} " | wc -l
```

#### Find Slow Requests (>100ms)

```bash
docker-compose logs api | awk '{if ($5 > 100000) print}'
```

#### Show All Requests from Specific IP

```bash
docker-compose logs api | grep "172.18.0.1"
```

#### Export All Logs for Analysis

```bash
# Export all logs to file
docker-compose logs api > full_logs.txt

# Export with date
docker-compose logs api > "logs_$(date +%Y%m%d_%H%M%S).txt"
```

### Integration with Monitoring Tools

The trace ID can be integrated with various monitoring platforms:

- **Sentry**: For error tracking and performance monitoring
- **Datadog**: For APM and log aggregation
- **ELK Stack**: For Elasticsearch, Logstash, Kibana
- **Grafana Loki**: For log aggregation and visualization
- **CloudWatch**: For AWS log monitoring
- **New Relic**: For full-stack observability

### Best Practices

1. **Always Log Trace ID**: Include trace ID in all application logs for correlation
2. **Client-Side Logging**: Store trace IDs from API responses for debugging
3. **Error Reports**: When reporting bugs, always include the trace ID
4. **Performance Analysis**: Use trace ID to correlate slow requests across services
5. **Security Auditing**: Trace IDs help audit who accessed what and when

### Example Debugging Session

```bash
# 1. Make a request that fails
curl -X GET http://localhost:8080/api/v1/products/999

# 2. Get the trace_id from response (e.g., e6b70517-9537-4e07-8ae1-ff400c027553)

# 3. Search in logs with context
docker-compose logs api | grep -B 5 -A 5 "e6b70517-9537-4e07-8ae1-ff400c027553"

# 4. Analyze the error
# The log shows: 404 | GET /api/v1/products/999

# 5. Check if there are any database errors for this trace
docker-compose logs api | grep "e6b70517-9537-4e07-8ae1-ff400c027553" | grep -i "error"
```

## Docker Commands

### Build and start services
```bash
docker-compose up -d
```

### Stop services
```bash
docker-compose down
```

### View logs
```bash
docker-compose logs -f
```

### Execute commands in container
```bash
docker-compose exec api /bin/sh
```

## Development

### Adding New Features

1. Add your model in `internal/models/`
2. Create repository in `internal/repository/`
3. Add handlers in `internal/handlers/`
4. Register routes in `cmd/server/main.go`
5. Add Swagger annotations to handlers
6. Regenerate Swagger docs:
   ```bash
   swag init -g cmd/server/main.go -o docs/swagger
   ```
