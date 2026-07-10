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
	database.InitSchema()
	defer database.Close()

	app := fiber.New()
	app.Use(cors.New())

	api := app.Group("/api")
	api.Get("/users", handlers.GetUsers)
	api.Get("/users/:id", handlers.GetUser)
	api.Post("/users", handlers.CreateUser)
	api.Put("/users/:id", handlers.UpdateUser)
	api.Delete("/users/:id", handlers.DeleteUser)

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
