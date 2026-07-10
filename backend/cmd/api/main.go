package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"dc-express/configs"
	"dc-express/internal/user"
	"dc-express/pkg/database"
	"dc-express/pkg/logger"
	"dc-express/pkg/response"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
)

func main() {
	logger.Init()

	cfg := configs.Load()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pool := database.Connect(ctx)
	defer pool.Close()

	database.RunMigrations(ctx, pool)

	userRepo := user.NewRepository(pool)
	userSvc := user.NewService(userRepo)
	userHdl := user.NewHandler(userSvc)

	app := fiber.New(fiber.Config{
		DisableStartupMessage: false,
	})

	app.Use(recover.New())
	app.Use(requestid.New())
	app.Use(cors.New())

	v1 := app.Group("/api/v1")

	v1.Get("/health", func(c *fiber.Ctx) error {
		dbStatus, latency, err := database.Ping(pool)
		if err != nil {
			dbStatus = "disconnected"
		}
		return c.JSON(response.OK(fiber.Map{
			"status":  "running",
			"db":      dbStatus,
			"latency": latency,
		}))
	})

	user.RegisterRoutes(v1, userHdl)

	staticDir := os.Getenv("STATIC_DIR")
	if staticDir == "" {
		staticDir = "./public"
	}
	app.Static("/", staticDir)

	go func() {
		slog.Info("server starting", "port", cfg.Port)
		if err := app.Listen(":" + cfg.Port); err != nil {
			slog.Error("server error", "error", err)
			cancel()
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("shutting down...")
	app.Shutdown()
}
