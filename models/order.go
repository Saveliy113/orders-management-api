package models

import (
	"time"

	"github.com/google/uuid"
)

type Order struct {
	ID           uuid.UUID `json:"id"`
	CustomerName string    `json:"customer_name"`
	Status       string    `json:"status"`
	Total        float64   `json:"total"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

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
