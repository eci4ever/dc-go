package database

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

func RunMigrations(ctx context.Context, pool *pgxpool.Pool) {
	if _, err := pool.Exec(ctx, `CREATE TABLE IF NOT EXISTS schema_migrations (version TEXT PRIMARY KEY, applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW())`); err != nil {
		fatal("failed to create migration ledger", err)
	}
	entries, err := os.ReadDir("migrations")
	if err != nil {
		slog.Error("failed to read migrations directory", "error", err)
		os.Exit(1)
	}

	var files []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".sql") {
			files = append(files, e.Name())
		}
	}
	sort.Strings(files)

	for _, f := range files {
		path := filepath.Join("migrations", f)
		sql, err := os.ReadFile(path)
		if err != nil {
			slog.Error("failed to read migration", "file", f, "error", err)
			os.Exit(1)
		}

		var applied bool
		if err := pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version=$1)`, f).Scan(&applied); err != nil {
			fatal("failed to read migration ledger", err)
		}
		if applied {
			continue
		}
		tx, err := pool.Begin(ctx)
		if err != nil {
			fatal("failed to begin migration", err)
		}
		if _, err := tx.Exec(ctx, string(sql)); err != nil {
			tx.Rollback(ctx)
			fatal("migration failed: "+f, err)
		}
		if _, err := tx.Exec(ctx, `INSERT INTO schema_migrations (version) VALUES ($1)`, f); err != nil {
			tx.Rollback(ctx)
			fatal("failed to record migration", err)
		}
		if err := tx.Commit(ctx); err != nil {
			fatal("failed to commit migration", err)
		}

		fmt.Printf("  applied: %s\n", f)
	}

	slog.Info("all migrations applied")
}

func fatal(message string, err error) {
	slog.Error(message, "error", err)
	panic(fmt.Sprintf("%s: %v", message, err))
}
