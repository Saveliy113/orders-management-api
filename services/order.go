package services

import (
	"context"

	"orders-management-api/dbservice"
	"orders-management-api/models"
)

func ListOrders(ctx context.Context, page, limit int, filters models.OrderFilters) (models.PaginatedOrders, error) {
	orders, total, err := dbservice.ListOrders(ctx, page, limit, filters)
	if err != nil {
		return models.PaginatedOrders{}, err
	}

	return models.PaginatedOrders{
		Data:  orders,
		Total: total,
	}, nil
}

func CreateOrder(ctx context.Context, input models.CreateOrderInput) (models.Order, error) {
	if input.Status == "" {
		input.Status = "pending"
	}
	return dbservice.CreateOrder(ctx, input)
}
