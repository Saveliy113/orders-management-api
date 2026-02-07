package main

import (
	"log"

	"orders-management-api/loaders"
	"orders-management-api/routes"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	log.Println("Starting orders-management-api...")
	defer loaders.CloseDB()

	log.Println("Initializing loaders...")
	loaders.InitDB()
	log.Println("Loaders initialized successfully")

	app := fiber.New()
	routes.Register(app)

	log.Println("Server listening on :3000")
	if err := app.Listen(":3000"); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
