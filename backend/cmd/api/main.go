package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"dc-express/configs"
	"dc-express/internal/auth"
	"dc-express/internal/organization"
	"dc-express/internal/team"
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

	cfg, err := configs.Load()
	if err != nil {
		slog.Error("invalid configuration", "error", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pool := database.Connect(ctx, cfg.DatabaseURL)
	defer pool.Close()

	database.RunMigrations(ctx, pool)

	// Repositories
	userRepo := user.NewRepository(pool)
	authRepo := auth.NewRepository(pool)
	orgRepo := organization.NewRepository(pool)
	teamRepo := team.NewRepository(pool)

	// JWT
	jwtSvc := auth.NewJWTService(cfg.JWTSecret, cfg.JWTIssuer, cfg.JWTAudience)

	// Services
	userSvc := user.NewService(userRepo)
	authSvc := auth.NewService(authRepo, jwtSvc, userRepo)
	orgSvc := organization.NewService(orgRepo)
	teamSvc := team.NewService(teamRepo)

	// Handlers
	userHdl := user.NewHandler(userSvc)
	authHdl := auth.NewHandler(authSvc, cfg.CookieSecure)
	orgHdl := organization.NewHandler(orgSvc)
	teamHdl := team.NewHandler(teamSvc)

	app := fiber.New(fiber.Config{
		DisableStartupMessage: false,
	})

	app.Use(recover.New())
	app.Use(requestid.New())
	app.Use(cors.New(cors.Config{AllowOrigins: strings.Join(cfg.AllowedOrigins, ","), AllowCredentials: true, AllowHeaders: "Origin, Content-Type, Accept, X-CSRF-Token"}))

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

	// Auth middleware
	authMw := auth.AuthMiddleware(jwtSvc)
	csrfMw := auth.CSRFMiddleware()

	// Register routes
	user.RegisterRoutes(v1, userHdl, authMw, csrfMw)
	auth.RegisterRoutes(v1, authHdl, authMw, csrfMw, cfg.CookieSecure)
	organization.RegisterRoutes(v1, orgHdl, authMw, csrfMw)
	team.RegisterRoutes(v1, teamHdl, authMw, csrfMw)

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
