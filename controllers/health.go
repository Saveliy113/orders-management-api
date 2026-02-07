package controllers

import "github.com/gofiber/fiber/v2"

func Alive(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"status": "ok",
	})
}
