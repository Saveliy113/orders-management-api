# Orders Management API

A REST API for managing orders, built with Go (Fiber framework) and PostgreSQL.

## Project Overview

This API provides endpoints to create and list orders with support for pagination and filtering. It follows a layered architecture similar to Node.js projects: routes → controllers → services → dbservice.

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
2. **Controllers** — Parse request (query params, body), validate input, call services
3. **Services** — Business logic (e.g. default status), call dbservice
4. **Dbservice** — Execute SQL queries via pgx pool

## Routes

| Method | Path | Description |
|--------|------|-------------|
| GET | `/alive` | Health check. Returns `{"status":"ok"}` |
| GET | `/orders` | List orders with pagination and filters |
| POST | `/orders` | Create a new order |

### GET /orders

**Query parameters:**

| Param | Type | Default | Description |
|-------|------|---------|-------------|
| `page` | int | 1 | Page number |
| `limit` | int | 10 | Items per page (max 100) |
| `status` | string | — | Filter by status (e.g. `pending`, `completed`, `cancelled`) |
| `min_amount` | float | — | Minimum order total |
| `max_amount` | float | — | Maximum order total |
| `from_date` | date | — | Orders created on or after (format: `2006-01-02` or RFC3339) |
| `to_date` | date | — | Orders created on or before |

**Response:**
```json
{
  "data": [
    {
      "id": "uuid",
      "customer_name": "string",
      "status": "string",
      "total": 99.99,
      "created_at": "2025-02-07T12:00:00Z",
      "updated_at": "2025-02-07T12:00:00Z"
    }
  ],
  "total": 50
}
```

### POST /orders

**Request body:**
```json
{
  "customer_name": "John Doe",
  "status": "pending",
  "total": 99.99
}
```

- `customer_name` — Required
- `status` — Optional, defaults to `"pending"`
- `total` — Optional, defaults to `0`

**Response (201):** Created order object

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

### Unit tests (no API or database required)

```bash
go test ./dbservice/... -v
```

Tests `buildWhereClause` for filters (status, amount, date range).

### API integration tests (API must be running)

Start the API first (e.g. `air` in one terminal), then:

```bash
go test . -v -run "TestGetOrders|TestPostOrder"
```

Tests hit the running API at `http://localhost:3000`. If the API is not available, tests are skipped.

**Override API URL:**
```bash
API_URL=http://localhost:8080 go test . -v -run "TestGetOrders|TestPostOrder"
```

### Run all tests

```bash
go test ./...
```

Unit tests always run; API tests are skipped when the API is not reachable.
