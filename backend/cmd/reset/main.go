// Command reset is intentionally development-only. It drops the public schema.
package main

import (
	"context"
	"github.com/eci4ever/dc-go/pkg/database"
	"log"
	"os"
)

func main() {
	if os.Getenv("ENVIRONMENT") != "development" {
		log.Fatal("reset is only available with ENVIRONMENT=development")
	}
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL is required")
	}
	pool := database.Connect(context.Background(), dsn)
	defer pool.Close()
	if _, err := pool.Exec(context.Background(), "DROP SCHEMA public CASCADE; CREATE SCHEMA public"); err != nil {
		log.Fatal(err)
	}
	database.RunMigrations(context.Background(), pool)
}
