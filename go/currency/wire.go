package currency

import (
	"context"
	"database/sql"
	"fmt"

	"go.uber.org/zap"
)

// WireOptions configures optional currency wiring behavior.
type WireOptions struct {
	Logger    *zap.Logger
	WorkerCtx context.Context
}

// ApplySchemaFromAppConfig resolves currency config and applies the schema when enabled.
func ApplySchemaFromAppConfig(ctx context.Context, db *sql.DB, app AppConfig) error {
	if !app.Enabled {
		return nil
	}

	cfg, err := ResolveBlockConfig(app)
	if err != nil {
		return fmt.Errorf("resolve currency config: %w", err)
	}

	if err := ApplySchema(ctx, db, cfg); err != nil {
		return fmt.Errorf("apply currency schema: %w", err)
	}

	return nil
}

// ApplySchemaFromFile loads legacy block YAML and applies the currency schema.
// Deprecated: prefer ApplySchemaFromAppConfig with the main config currency node.
func ApplySchemaFromFile(ctx context.Context, db *sql.DB, configPath string) error {
	return ApplySchemaFromAppConfig(ctx, db, AppConfig{Enabled: true, ConfigPath: configPath})
}

// Wire applies the currency schema and builds the service when app.Enabled is true.
// Returns nil, nil when disabled. Starts the exchange-rate updater when WorkerCtx is set.
func Wire(ctx context.Context, db *sql.DB, app AppConfig, opts WireOptions) (*Service, error) {
	if !app.Enabled {
		return nil, nil
	}

	blockCfg, err := ResolveBlockConfig(app)
	if err != nil {
		return nil, fmt.Errorf("resolve currency config: %w", err)
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

	svc.logger.Info("currency block enabled",
		zap.Duration("interval", app.UpdateInterval),
		zap.String("api", app.APIURL),
	)
	return svc, nil
}
