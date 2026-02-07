package dbservice

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"orders-management-api/loaders"
	"orders-management-api/models"
)

func ListOrders(ctx context.Context, page, limit int, filters models.OrderFilters) ([]models.Order, int64, error) {
	db := loaders.DB
	offset := (page - 1) * limit
	if offset < 0 {
		offset = 0
	}
	if limit < 1 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	where, args := buildWhereClause(filters)
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM orders%s", where)
	total := int64(0)
	err := db.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	args = append(args, limit, offset)
	limitOffset := fmt.Sprintf(" LIMIT $%d OFFSET $%d", len(args)-1, len(args))
	selectQuery := fmt.Sprintf(
		`SELECT id, customer_name, status, total, created_at, updated_at
		 FROM orders%s
		 ORDER BY created_at DESC%s`,
		where, limitOffset,
	)
	rows, err := db.Query(ctx, selectQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var orders []models.Order
	for rows.Next() {
		var o models.Order
		err := rows.Scan(&o.ID, &o.CustomerName, &o.Status, &o.Total, &o.CreatedAt, &o.UpdatedAt)
		if err != nil {
			return nil, 0, err
		}
		orders = append(orders, o)
	}

	return orders, total, rows.Err()
}

func CreateOrder(ctx context.Context, input models.CreateOrderInput) (models.Order, error) {
	db := loaders.DB

	id := uuid.New()

	var order models.Order
	err := db.QueryRow(ctx,
		`INSERT INTO orders (id, customer_name, status, total, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, NOW(), NOW())
		 RETURNING id, customer_name, status, total, created_at, updated_at`,
		id, input.CustomerName, input.Status, input.Total,
	).Scan(&order.ID, &order.CustomerName, &order.Status, &order.Total, &order.CreatedAt, &order.UpdatedAt)
	if err != nil {
		return models.Order{}, err
	}

	return order, nil
}

func buildWhereClause(filters models.OrderFilters) (string, []interface{}) {
	var conditions []string
	var args []interface{}
	paramNum := 1

	if filters.Status != nil && *filters.Status != "" {
		conditions = append(conditions, fmt.Sprintf("status = $%d", paramNum))
		args = append(args, *filters.Status)
		paramNum++
	}
	if filters.MinAmount != nil {
		conditions = append(conditions, fmt.Sprintf("total >= $%d", paramNum))
		args = append(args, *filters.MinAmount)
		paramNum++
	}
	if filters.MaxAmount != nil {
		conditions = append(conditions, fmt.Sprintf("total <= $%d", paramNum))
		args = append(args, *filters.MaxAmount)
		paramNum++
	}
	if filters.FromDate != nil {
		conditions = append(conditions, fmt.Sprintf("created_at >= $%d", paramNum))
		args = append(args, *filters.FromDate)
		paramNum++
	}
	if filters.ToDate != nil {
		conditions = append(conditions, fmt.Sprintf("created_at <= $%d", paramNum))
		args = append(args, *filters.ToDate)
	}

	if len(conditions) == 0 {
		return "", args
	}
	return " WHERE " + strings.Join(conditions, " AND "), args
}
