package dbservice

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"

	"orders-management-api/loaders"
	"orders-management-api/models"
)

func ListOrders(ctx context.Context, filter models.OrderFilter) (models.OrderListResponse, error) {
	db := loaders.DB
	if db == nil {
		return models.OrderListResponse{}, errors.New("database connection is not initialized")
	}

	var (
		conditions []string
		args       []any
		argIndex   int
	)

	if filter.Status != "" {
		argIndex++
		conditions = append(conditions, fmt.Sprintf("status = $%d", argIndex))
		args = append(args, filter.Status)
	}

	if !filter.DateFrom.IsZero() {
		argIndex++
		conditions = append(conditions, fmt.Sprintf("created_at >= $%d", argIndex))
		args = append(args, filter.DateFrom)
	}

	if !filter.DateTo.IsZero() {
		argIndex++
		conditions = append(conditions, fmt.Sprintf("created_at <= $%d", argIndex))
		args = append(args, filter.DateTo)
	}

	if filter.AmountMin != nil {
		argIndex++
		conditions = append(conditions, fmt.Sprintf("total >= $%d", argIndex))
		args = append(args, *filter.AmountMin)
	}

	if filter.AmountMax != nil {
		argIndex++
		conditions = append(conditions, fmt.Sprintf("total <= $%d", argIndex))
		args = append(args, *filter.AmountMax)
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = " WHERE " + strings.Join(conditions, " AND ")
	}

	// Count total matching records.
	countQuery := "SELECT COUNT(*) FROM orders" + whereClause

	var total int
	if err := db.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return models.OrderListResponse{}, err
	}

	totalPages := int(math.Ceil(float64(total) / float64(filter.Limit)))

	// Fetch the page of orders.
	offset := (filter.Page - 1) * filter.Limit

	argIndex++
	limitPlaceholder := fmt.Sprintf("$%d", argIndex)
	argIndex++
	offsetPlaceholder := fmt.Sprintf("$%d", argIndex)

	dataQuery := fmt.Sprintf(
		`SELECT id, customer_name, status, total, created_at, updated_at
		 FROM orders%s
		 ORDER BY created_at DESC
		 LIMIT %s OFFSET %s`,
		whereClause, limitPlaceholder, offsetPlaceholder,
	)

	dataArgs := make([]any, len(args), len(args)+2)
	copy(dataArgs, args)
	dataArgs = append(dataArgs, filter.Limit, offset)

	rows, err := db.Query(ctx, dataQuery, dataArgs...)
	if err != nil {
		return models.OrderListResponse{}, err
	}
	defer rows.Close()

	orders := make([]models.Order, 0, filter.Limit)
	for rows.Next() {
		var o models.Order
		if err := rows.Scan(&o.ID, &o.CustomerName, &o.Status, &o.Total, &o.CreatedAt, &o.UpdatedAt); err != nil {
			return models.OrderListResponse{}, err
		}
		orders = append(orders, o)
	}

	if err := rows.Err(); err != nil {
		return models.OrderListResponse{}, err
	}

	return models.OrderListResponse{
		Orders:     orders,
		Total:      total,
		TotalPages: totalPages,
	}, nil
}
