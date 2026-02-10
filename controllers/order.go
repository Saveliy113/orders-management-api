package controllers

import (
	"github.com/gofiber/fiber/v2"

	"orders-management-api/services"
)

func GetOrders(c *fiber.Ctx) error {
	orders, err := services.ListOrders(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(orders)
}
