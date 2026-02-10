# Add Pagination and Filtering to GET /orders

## Current State

The endpoint returns all orders with no filtering or pagination. The data flows through four layers:

```
controllers/order.go -> services/order.go -> dbservice/order.go
```

All functions currently accept only `context.Context` and return `([]models.Order, error)`.

## Query Parameters

- `page` — int, default `1`, min `1`
- `limit` — int, default `20`, min `1`, max `100`
- `status` — string, optional, one of: `pending`, `completed`, `cancelled`
- `date_from` — string (RFC 3339 / `2006-01-02`), optional, inclusive lower bound on `created_at`
- `date_to` — string (RFC 3339 / `2006-01-02`), optional, inclusive upper bound on `created_at`
- `amount_min` — float64, optional, inclusive lower bound on `total`
- `amount_max` — float64, optional, inclusive upper bound on `total`

## Changes by File

### 1. models/order.go — Add filter and response structs

Add two new structs:

```go
type OrderFilter struct {
    Page      int
    Limit     int
    Status    string
    DateFrom  time.Time
    DateTo    time.Time
    AmountMin *float64
    AmountMax *float64
}

type OrderListResponse struct {
    Orders     []Order `json:"orders"`
    Total      int     `json:"total"`
    TotalPages int     `json:"total_pages"`
}
```

- `AmountMin`/`AmountMax` are pointers so we can distinguish "not set" from `0`.
- `DateFrom`/`DateTo` use zero-value `time.Time` to indicate "not set".
- `Total` is **not** a database column — it is computed via `SELECT COUNT(*)` in the dbservice layer.
- Response only includes `total` (matching record count) and `total_pages` (ceiling of total/limit). `page` and `limit` are NOT returned in the response.

### 2. controllers/order.go — Parse query params and validate

- Parse `page`, `limit` from query string with defaults (`1` and `20`).
- Clamp `limit` to `[1, 100]`, clamp `page` to `>= 1`.
- Parse `status` and validate against allowed values.
- Parse `date_from` / `date_to` as `time.Time` (support both `2006-01-02` and RFC 3339).
- Parse `amount_min` / `amount_max` as `*float64`.
- Return `400 Bad Request` with a descriptive message for invalid inputs.
- Build an `OrderFilter` struct and pass it to `services.ListOrders`.
- Return the `OrderListResponse` JSON directly.

### 3. services/order.go — Pass filter through

Update signature to accept `models.OrderFilter` and return `models.OrderListResponse`:

```go
func ListOrders(ctx context.Context, filter models.OrderFilter) (models.OrderListResponse, error) {
    return dbservice.ListOrders(ctx, filter)
}
```

### 4. dbservice/order.go — Build dynamic SQL with filters

This is the core change. The function will:

- Build a `WHERE` clause dynamically using a slice of conditions and a `[]any` args slice, incrementing a placeholder counter (`$1`, `$2`, ...) for safe parameterized queries.
- Apply filters conditionally:
  - `status != ""` adds `status = $N`
  - `!DateFrom.IsZero()` adds `created_at >= $N`
  - `!DateTo.IsZero()` adds `created_at <= $N`
  - `AmountMin != nil` adds `total >= $N`
  - `AmountMax != nil` adds `total <= $N`
- Run a **separate** `SELECT COUNT(*) FROM orders` query with the same dynamically-built `WHERE` clause to compute the total number of matching records. This is the sole source of the `total` field in the response.
- Run the main `SELECT` query with `ORDER BY created_at DESC`, `LIMIT $N OFFSET $N` (where offset = `(page - 1) * limit`).
- Compute `TotalPages = ceil(total / limit)` and assemble `OrderListResponse{Orders, Total, TotalPages}`.

Existing indexes on `created_at DESC` and `status` already cover the most common filter paths.

## Data Flow After Changes

```
Client
  -> GET /orders?page=2&limit=10&status=pending
  -> controllers/order.go  (parse & validate query params)
  -> services/order.go     (pass through OrderFilter)
  -> dbservice/order.go    (build WHERE clause & args)
     -> PostgreSQL: SELECT COUNT(*) FROM orders WHERE ...  => total
     -> PostgreSQL: SELECT ... FROM orders WHERE ... LIMIT $N OFFSET $N  => rows
     -> Compute TotalPages = ceil(total / limit)
  <- OrderListResponse flows back through layers
  <- JSON response to client
```

## Example Response

```json
{
  "orders": [
    {
      "id": "...",
      "customer_name": "Alice Johnson",
      "status": "pending",
      "total": 49.99,
      "created_at": "...",
      "updated_at": "..."
    }
  ],
  "total": 23,
  "total_pages": 3
}
```

Note: `total` in the response is the count of all matching records (from `SELECT COUNT(*)`), not to be confused with the `total` column on individual orders (which is the order amount).
