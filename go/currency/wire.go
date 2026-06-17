package currency

import (
	"context"
	"database/sql"
	"fmt"
	"os"

	"go.uber.org/zap"
)

// WireOptions configures optional currency wiring behavior.
type WireOptions struct {
	Logger    *zap.Logger
	WorkerCtx context.Context
}

// ApplySchemaFromFile loads block YAML and applies the currency schema.
func ApplySchemaFromFile(ctx context.Context, db *sql.DB, configPath string) error {
	path := configPath
	if path == "" {
		path = defaultAppConfigPath
	}
	if override := os.Getenv("CURRENCY_CONFIG"); override != "" {
		path = override
	}

	cfg, err := LoadConfig(path)
	if err != nil {
		return fmt.Errorf("load currency config: %w", err)
	}

	if err := ApplySchema(ctx, db, cfg); err != nil {
		return fmt.Errorf("apply currency schema: %w", err)
	}

	return nil
}

// Wire applies the currency schema and builds the service when app.Enabled is true.
// Returns nil, nil when disabled. Starts the exchange-rate updater when WorkerCtx is set.
func Wire(ctx context.Context, db *sql.DB, app AppConfig, opts WireOptions) (*Service, error) {
	if !app.Enabled {
		return nil, nil
	}

	app.ApplyDefaults()

	blockCfg, err := LoadConfig(app.blockConfigPath())
	if err != nil {
		return nil, fmt.Errorf("load currency config: %w", err)
	}

	if err := ApplySchema(ctx, db, blockCfg); err != nil {
		return nil, fmt.Errorf("apply currency schema: %w", err)
	}

	svc, err := NewService(db, blockCfg, Options{
		Logger: opts.Logger,
		APIURL: app.APIURL,
	})
	if err != nil {
		return nil, err
	}

	if opts.WorkerCtx != nil && app.UpdateInterval > 0 {
		go svc.RunExchangeRateUpdater(opts.WorkerCtx, app.UpdateInterval)
	}

	return svc, nil
}
