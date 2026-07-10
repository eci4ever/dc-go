package database

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func Connect(ctx context.Context, dsn string) *pgxpool.Pool {

	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		slog.Error("failed to parse database config", "error", err)
		os.Exit(1)
	}

	cfg.MaxConns = 25
	cfg.MinConns = 5

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		slog.Error("failed to create connection pool", "error", err)
		os.Exit(1)
	}

	if err := pool.Ping(ctx); err != nil {
		slog.Error("failed to ping database", "error", err)
		os.Exit(1)
	}

	slog.Info("connected to PostgreSQL")
	return pool
}

func Ping(pool *pgxpool.Pool) (string, int64, error) {
	start := time.Now()
	err := pool.Ping(context.Background())
	latency := time.Since(start).Milliseconds()
	if err != nil {
		return "disconnected", latency, err
	}
	return "connected", latency, nil
}
