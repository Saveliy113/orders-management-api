package dbservice

import (
	"context"

	"orders-management-api/loaders"
	"orders-management-api/models"
)

func ListOrders(ctx context.Context) ([]models.Order, error) {
	db := loaders.DB

	rows, err := db.Query(ctx,
		`SELECT id, customer_name, status, total, created_at, updated_at
		 FROM orders
		 ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []models.Order
	for rows.Next() {
		var o models.Order
		err := rows.Scan(&o.ID, &o.CustomerName, &o.Status, &o.Total, &o.CreatedAt, &o.UpdatedAt)
		if err != nil {
			return nil, err
		}
		orders = append(orders, o)
	}

	return orders, rows.Err()
}
