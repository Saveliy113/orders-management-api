package services

import (
	"context"

	"orders-management-api/dbservice"
	"orders-management-api/models"
)

func ListOrders(ctx context.Context, filter models.OrderFilter) (models.OrderListResponse, error) {
	return dbservice.ListOrders(ctx, filter)
}
