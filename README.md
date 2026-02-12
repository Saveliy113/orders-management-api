# Orders Management API

A REST API for managing orders, built with Go (Fiber framework) and PostgreSQL.

## Project Overview

This API provides an endpoint to list orders. It follows a layered architecture: routes → controllers → services → dbservice.

## Project Structure

```
orders-management-api/
├── main.go              # Entry point, bootstraps app and server
├── loaders/
│   └── db.go            # PostgreSQL connection pool (pgx)
├── routes/
│   └── routes.go        # Route definitions
├── controllers/         # HTTP request handlers
├── services/            # Business logic
├── dbservice/           # Database queries (SQL)
├── models/              # Data structures
├── migrations/          # SQL migrations (run on first DB init)
├── docker-compose.yml   # PostgreSQL container
├── .env                 # Environment variables (DATABASE_URL)
└── .air.toml            # Live reload config
```

## Logic Flow

1. **Routes** — Define HTTP endpoints and map them to controllers
2. **Controllers** — Parse request, call services
3. **Services** — Business logic, call dbservice
4. **Dbservice** — Execute SQL queries via pgx pool

## Routes

| Method | Path | Description |
|--------|------|-------------|
| GET | `/alive` | Health check. Returns `{"status":"ok"}` |
| GET | `/orders` | List orders with pagination and filtering |

Full OpenAPI specification is available in [`api-docs.yaml`](api-docs.yaml).

### GET /orders

Returns a paginated, optionally filtered list of orders sorted by `created_at` descending.

#### Query Parameters

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `page` | int | `1` | Page number (min 1) |
| `limit` | int | `20` | Items per page (min 1, max 100) |
| `status` | string | — | Filter by status: `pending`, `completed`, `cancelled` |
| `date_from` | string | — | Inclusive lower bound on `created_at` (`YYYY-MM-DD` or RFC 3339) |
| `date_to` | string | — | Inclusive upper bound on `created_at` (`YYYY-MM-DD` or RFC 3339) |
| `amount_min` | float | — | Inclusive lower bound on order total |
| `amount_max` | float | — | Inclusive upper bound on order total |

All filter parameters are optional. When omitted, no filtering is applied for that field.

#### Response (200 OK)

```json
{
  "orders": [
    {
      "id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
      "customer_name": "Alice Johnson",
      "status": "pending",
      "total": 49.99,
      "created_at": "2025-02-06T12:00:00Z",
      "updated_at": "2025-02-07T12:00:00Z"
    }
  ],
  "total": 50,
  "total_pages": 3
}
```

- `total` — total number of matching records (computed via `SELECT COUNT(*)`)
- `total_pages` — `ceil(total / limit)`

#### Error Response (400 Bad Request)

Returned when a query parameter is invalid:

```json
{
  "error": "invalid status parameter, must be one of: pending, completed, cancelled"
}
```

#### Examples

```bash
# Default: first 20 orders
curl http://localhost:3000/orders

# Page 2 with 10 items per page
curl "http://localhost:3000/orders?page=2&limit=10"

# Only pending orders
curl "http://localhost:3000/orders?status=pending"

# Orders created in January 2025
curl "http://localhost:3000/orders?date_from=2025-01-01&date_to=2025-01-31"

# Orders between $50 and $200
curl "http://localhost:3000/orders?amount_min=50&amount_max=200"

# Combined: pending orders over $100, page 1, 5 per page
curl "http://localhost:3000/orders?status=pending&amount_min=100&page=1&limit=5"
```

## Database

PostgreSQL 16 with the following credentials (configurable via docker-compose):

- **User:** postgres
- **Password:** postgres
- **Database:** orders_management
- **Port:** 5432

### Schema

The `orders` table is created automatically by the migration in `migrations/001_create_orders.sql` on first database init. It also inserts 50 sample orders.

## How to Start

### Prerequisites

- Go 1.24+
- Docker & Docker Compose (for PostgreSQL)
- [Air](https://github.com/air-verse/air) (optional, for live reload)

### 1. Start the database

```bash
docker compose up -d
```

Migrations run automatically when the container is first created. To reset the database:

```bash
docker compose down -v
docker compose up -d
```

### 2. Configure environment

Create `.env` in the project root (or use the default):

```
DATABASE_URL=postgres://postgres:postgres@localhost:5432/orders_management?sslmode=disable
```

### 3. Run the API

```bash
# Standard run
go run .

# With live reload (Air)
air
```

The server listens on **http://localhost:3000**

## How to Run Tests

### API integration tests (API must be running)

Start the API first (e.g. `air` in one terminal), then:

```bash
go test . -v -run TestGetOrders
```

Tests hit the running API at `http://localhost:3000`. If the API is not available, tests are skipped.

**Override API URL:**
```bash
API_URL=http://localhost:8080 go test . -v -run TestGetOrders
```
