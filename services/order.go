package services

import (
	"context"

	"orders-management-api/dbservice"
	"orders-management-api/models"
)

func ListOrders(ctx context.Context) ([]models.Order, error) {
	return dbservice.ListOrders(ctx)
}
