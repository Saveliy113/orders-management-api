package models

import (
	"time"

	"github.com/google/uuid"
)

type OrderFilters struct {
	Status    *string    `json:"status"`
	MinAmount *float64   `json:"min_amount"`
	MaxAmount *float64   `json:"max_amount"`
	FromDate  *time.Time `json:"from_date"`
	ToDate    *time.Time `json:"to_date"`
}

type Order struct {
	ID           uuid.UUID `json:"id"`
	CustomerName string    `json:"customer_name"`
	Status       string    `json:"status"`
	Total        float64   `json:"total"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type CreateOrderInput struct {
	CustomerName string  `json:"customer_name"`
	Status       string  `json:"status"`
	Total        float64 `json:"total"`
}

type PaginatedOrders struct {
	Data  []Order `json:"data"`
	Total int64   `json:"total"`
}
