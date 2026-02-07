package controllers

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"

	"orders-management-api/models"
	"orders-management-api/services"
)

func GetOrders(c *fiber.Ctx) error {
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 10)

	filters := parseOrderFilters(c)

	result, err := services.ListOrders(c.Context(), page, limit, filters)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(result)
}

func PostOrder(c *fiber.Ctx) error {
	var input models.CreateOrderInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body: " + err.Error(),
		})
	}

	if input.CustomerName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "customer_name is required",
		})
	}

	order, err := services.CreateOrder(c.Context(), input)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(order)
}

func parseOrderFilters(c *fiber.Ctx) models.OrderFilters {
	var filters models.OrderFilters

	if s := c.Query("status"); s != "" {
		filters.Status = &s
	}
	if s := c.Query("min_amount"); s != "" {
		if v, err := strconv.ParseFloat(s, 64); err == nil {
			filters.MinAmount = &v
		}
	}
	if s := c.Query("max_amount"); s != "" {
		if v, err := strconv.ParseFloat(s, 64); err == nil {
			filters.MaxAmount = &v
		}
	}
	if s := c.Query("from_date"); s != "" {
		if t, err := time.Parse(time.RFC3339, s); err == nil {
			filters.FromDate = &t
		} else if t, err := time.Parse("2006-01-02", s); err == nil {
			filters.FromDate = &t
		}
	}
	if s := c.Query("to_date"); s != "" {
		if t, err := time.Parse(time.RFC3339, s); err == nil {
			filters.ToDate = &t
		} else if t, err := time.Parse("2006-01-02", s); err == nil {
			filters.ToDate = &t
		}
	}

	return filters
}
