package migrate

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

// RunMigrations applies pending goose migrations on startup.
func RunMigrations(ctx context.Context, pool *pgxpool.Pool, migrationsPath string) error {
	if err := pool.Ping(ctx); err != nil {
		return fmt.Errorf("ping postgres: %w", err)
	}

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("goose dialect: %w", err)
	}

	sqlDB := stdlib.OpenDBFromPool(pool)
	defer sqlDB.Close()

	if err := goose.UpContext(ctx, sqlDB, migrationsPath); err != nil {
		return fmt.Errorf("goose up: %w", err)
	}

	return nil
}

// RunGoose runs a goose CLI command (up, down, status, create, etc.).
func RunGoose(ctx context.Context, dsn, migrationsPath, command string, args []string) error {
	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("goose dialect: %w", err)
	}

	sqlDB, err := sql.Open("pgx", dsn)
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	defer sqlDB.Close()

	if err := sqlDB.PingContext(ctx); err != nil {
		return fmt.Errorf("ping database: %w", err)
	}

	return goose.RunContext(ctx, command, sqlDB, migrationsPath, args...)
}
