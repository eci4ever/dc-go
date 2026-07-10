package handlers

import (
	"fmt"

	"dc-express/database"

	"github.com/gofiber/fiber/v2"
)

func HealthCheck(c *fiber.Ctx) error {
	dbStatus, latency, err := database.Ping()
	if err != nil {
		dbStatus = "disconnected"
	}

	return c.JSON(fiber.Map{
		"status":  "running",
		"db":      dbStatus,
		"latency": fmt.Sprintf("%dms", latency),
	})
}
