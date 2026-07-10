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

		if _, err := pool.Exec(ctx, string(sql)); err != nil {
			slog.Error("migration failed", "file", f, "error", err)
			os.Exit(1)
		}

		fmt.Printf("  applied: %s\n", f)
	}

	slog.Info("all migrations applied")
}
