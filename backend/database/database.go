package database

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

var Pool *pgxpool.Pool

func Connect() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://postgres:postgres@localhost:5432/dc_express"
	}

	var err error
	Pool, err = pgxpool.New(context.Background(), dsn)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}

	if err = Pool.Ping(context.Background()); err != nil {
		log.Fatalf("Unable to ping database: %v", err)
	}

	fmt.Println("Connected to PostgreSQL")
}

func Ping() (string, int64, error) {
	start := time.Now()
	err := Pool.Ping(context.Background())
	latency := time.Since(start).Milliseconds()
	if err != nil {
		return "disconnected", latency, err
	}
	return "connected", latency, nil
}

func Close() {
	if Pool != nil {
		Pool.Close()
	}
}
