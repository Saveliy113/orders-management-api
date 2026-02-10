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
| GET | `/orders` | List all orders |

### GET /orders

**Response:**
```json
[
  {
    "id": "uuid",
    "customer_name": "string",
    "status": "string",
    "total": 99.99,
    "created_at": "2025-02-07T12:00:00Z",
    "updated_at": "2025-02-07T12:00:00Z"
  }
]
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
