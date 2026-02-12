package controllers

import (
	"log"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"

	"orders-management-api/models"
	"orders-management-api/services"
)

func GetOrders(c *fiber.Ctx) error {
	// Parse pagination params.
	page, err := strconv.Atoi(c.Query("page", "1"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid page parameter",
		})
	}
	if page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(c.Query("limit", "20"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid limit parameter",
		})
	}
	if limit < 1 {
		limit = 1
	}
	if limit > 100 {
		limit = 100
	}

	// Parse status filter.
	status := c.Query("status")
	if status != "" {
		validStatuses := map[string]bool{
			"pending":   true,
			"completed": true,
			"cancelled": true,
		}
		if !validStatuses[status] {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "invalid status parameter, must be one of: pending, completed, cancelled",
			})
		}
	}

	// Parse date range filters.
	var dateFrom, dateTo time.Time

	if df := c.Query("date_from"); df != "" {
		dateFrom, err = parseDate(df)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "invalid date_from parameter, expected format: YYYY-MM-DD or RFC 3339",
			})
		}
	}

	if dt := c.Query("date_to"); dt != "" {
		dateTo, err = parseDate(dt)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "invalid date_to parameter, expected format: YYYY-MM-DD or RFC 3339",
			})
		}
	}

	// Parse amount range filters.
	var amountMin, amountMax *float64

	if am := c.Query("amount_min"); am != "" {
		v, err := strconv.ParseFloat(am, 64)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "invalid amount_min parameter",
			})
		}
		amountMin = &v
	}

	if am := c.Query("amount_max"); am != "" {
		v, err := strconv.ParseFloat(am, 64)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "invalid amount_max parameter",
			})
		}
		amountMax = &v
	}

	filter := models.OrderFilter{
		Page:      page,
		Limit:     limit,
		Status:    status,
		DateFrom:  dateFrom,
		DateTo:    dateTo,
		AmountMin: amountMin,
		AmountMax: amountMax,
	}

	result, err := services.ListOrders(c.Context(), filter)
	if err != nil {
		log.Printf("ERROR GetOrders: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "internal server error",
		})
	}

	return c.JSON(result)
}

// parseDate tries RFC 3339 first, then falls back to YYYY-MM-DD.
func parseDate(s string) (time.Time, error) {
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t, nil
	}
	return time.Parse("2006-01-02", s)
}
