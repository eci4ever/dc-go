package main

import (
	"log"
	"os"

	"dc-express/database"
	"dc-express/handlers"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func main() {
	database.Connect()
	defer database.Close()

	app := fiber.New()
	app.Use(cors.New())

	api := app.Group("/api")
	api.Get("/health", handlers.HealthCheck)

	staticDir := os.Getenv("STATIC_DIR")
	if staticDir == "" {
		staticDir = "./public"
	}
	app.Static("/", staticDir)

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	log.Fatal(app.Listen(":" + port))
}
