package migrate

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
)

// ApplyConfig configures RunMigrations.
type ApplyConfig struct {
	Instructions []Instruction
}

// RunMigrations applies pending SQL instructions on startup.
func RunMigrations(ctx context.Context, pool *pgxpool.Pool, cfg ApplyConfig) error {
	if err := pool.Ping(ctx); err != nil {
		return fmt.Errorf("ping postgres: %w", err)
	}
	if len(cfg.Instructions) == 0 {
		return nil
	}

	sqlDB := stdlib.OpenDBFromPool(pool)
	defer sqlDB.Close()

	runner := NewRunner(sqlDB)
	if err := runner.Register(cfg.Instructions...); err != nil {
		return err
	}
	if err := runner.Apply(ctx); err != nil {
		return fmt.Errorf("apply migrations: %w", err)
	}
	return nil
}

// OpenRunner opens a sql.DB runner for CLI use.
func OpenRunner(ctx context.Context, dsn string, cfg ApplyConfig) (*Runner, *sql.DB, error) {
	sqlDB, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, nil, fmt.Errorf("open database: %w", err)
	}
	if err := sqlDB.PingContext(ctx); err != nil {
		sqlDB.Close()
		return nil, nil, fmt.Errorf("ping database: %w", err)
	}
	runner := NewRunner(sqlDB)
	if err := runner.Register(cfg.Instructions...); err != nil {
		sqlDB.Close()
		return nil, nil, err
	}
	return runner, sqlDB, nil
}
