package routes

import (
	"orders-management-api/controllers"

	"github.com/gofiber/fiber/v2"
)

func Register(app *fiber.App) {
	app.Get("/alive", controllers.Alive)
	app.Get("/orders", controllers.GetOrders)
}
