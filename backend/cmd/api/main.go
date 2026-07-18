package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/eci4ever/dc-go/configs"
	"github.com/eci4ever/dc-go/internal/academic"
	"github.com/eci4ever/dc-go/internal/auth"
	"github.com/eci4ever/dc-go/internal/organization"
	"github.com/eci4ever/dc-go/internal/storage"
	"github.com/eci4ever/dc-go/internal/team"
	"github.com/eci4ever/dc-go/internal/user"
	"github.com/eci4ever/dc-go/pkg/database"
	"github.com/eci4ever/dc-go/pkg/logger"
	"github.com/eci4ever/dc-go/pkg/ratelimit"
	"github.com/eci4ever/dc-go/pkg/redisclient"
	"github.com/eci4ever/dc-go/pkg/response"

	"github.com/gofiber/fiber/v2"
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
	redisClient, err := redisclient.Connect(ctx, cfg.RedisURL)
	if err != nil {
		slog.Error("failed to connect to Redis", "error", err)
		os.Exit(1)
	}
	defer redisClient.Close()
	slog.Info("connected to Redis")

	database.RunMigrations(ctx, pool)
	objectStore, err := storage.NewS3Store(ctx, storage.Config{
		Endpoint:       cfg.S3Endpoint,
		AccessKey:      cfg.S3AccessKey,
		SecretKey:      cfg.S3SecretKey,
		Bucket:         cfg.S3Bucket,
		Region:         cfg.S3Region,
		UseSSL:         cfg.S3UseSSL,
		ForcePathStyle: cfg.S3PathStyle,
	})
	if err != nil {
		slog.Error("failed to configure avatar storage", "error", err)
		os.Exit(1)
	}

	// Repositories
	userRepo := user.NewRepository(pool)
	authRepo := auth.NewRepository(pool)
	orgRepo := organization.NewRepository(pool)
	teamRepo := team.NewRepository(pool)
	// Academic records are scoped to the active institute organization.
	academicRepo := academic.NewRepository(pool)

	// JWT
	jwtSvc := auth.NewJWTService(cfg.JWTSecret, cfg.JWTIssuer, cfg.JWTAudience)

	// Services
	userSvc := user.NewService(userRepo, objectStore)
	authSvc := auth.NewService(authRepo, jwtSvc, userRepo)
	orgSvc := organization.NewService(orgRepo, objectStore)
	teamSvc := team.NewService(teamRepo)
	academicSvc := academic.NewService(academicRepo)

	// Handlers
	userHdl := user.NewHandler(userSvc)
	authHdl := auth.NewHandler(authSvc, cfg.CookieSecure)
	orgHdl := organization.NewHandler(orgSvc)
	teamHdl := team.NewHandler(teamSvc)
	academicHdl := academic.NewHandler(academicSvc)

	app := fiber.New(fiber.Config{
		DisableStartupMessage: false,
	})

	app.Use(recover.New())
	app.Use(requestid.New())

	v1 := app.Group("/api/v1")

	v1.Get("/health", func(c *fiber.Ctx) error {
		dbStatus, latency, err := database.Ping(pool)
		if err != nil {
			dbStatus = "disconnected"
		}
		redisStatus, redisLatency, redisErr := redisclient.Ping(c.UserContext(), redisClient)
		status := "running"
		if err != nil || redisErr != nil {
			status = "degraded"
		}
		return c.JSON(response.OK(fiber.Map{
			"status":  status,
			"db":      dbStatus,
			"latency": latency,
			"postgres": fiber.Map{
				"status":     dbStatus,
				"latency_ms": latency,
			},
			"redis": fiber.Map{
				"status":     redisStatus,
				"latency_ms": redisLatency,
			},
		}))
	})

	// Auth middleware
	authMw := auth.AuthMiddleware(jwtSvc)
	csrfMw := auth.CSRFMiddleware()
	adminMw := auth.RequireUserRole(userRepo, user.RoleAdmin)
	rateStore := ratelimit.NewRedisStore(redisClient)
	authLimits := auth.RateLimiters{
		Register:       ratelimit.Middleware(rateStore, ratelimit.Policy{Name: "auth-register", Limit: 3, Window: time.Hour, Key: ratelimit.ByIP}),
		Login:          ratelimit.Middleware(rateStore, ratelimit.Policy{Name: "auth-login", Limit: 5, Window: 15 * time.Minute, Key: ratelimit.ByIP}),
		Refresh:        ratelimit.Middleware(rateStore, ratelimit.Policy{Name: "auth-refresh", Limit: 30, Window: time.Minute, Key: ratelimit.ByIP}),
		ChangePassword: ratelimit.Middleware(rateStore, ratelimit.Policy{Name: "auth-password", Limit: 5, Window: time.Hour, Key: ratelimit.ByUser}),
	}
	invitationLimit := ratelimit.Middleware(rateStore, ratelimit.Policy{Name: "invitation-create", Limit: 20, Window: time.Hour, Key: ratelimit.ByParam("id")})

	// Register routes
	user.RegisterRoutes(v1, userHdl, authMw, csrfMw, adminMw)
	auth.RegisterRoutes(v1, authHdl, authMw, csrfMw, authLimits)
	organization.RegisterRoutes(v1, orgHdl, invitationLimit, authMw, csrfMw)
	organization.RegisterAdminRoutes(v1, orgHdl, authMw, csrfMw, adminMw)
	team.RegisterRoutes(v1, teamHdl, authMw, csrfMw)
	academic.RegisterRoutes(v1, academicHdl, authMw, csrfMw)

	staticDir := os.Getenv("STATIC_DIR")
	if staticDir == "" {
		staticDir = "./public"
	}
	app.Static("/", staticDir)
	app.Get("/*", func(c *fiber.Ctx) error {
		return c.SendFile(filepath.Join(staticDir, "index.html"))
	})

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
