package accounts

import (
	"context"
	"database/sql"
	"fmt"
)

// ApplySchemaFromAppConfig resolves accounts config and applies the schema when enabled.
func ApplySchemaFromAppConfig(ctx context.Context, db *sql.DB, app AppConfig) error {
	if !app.Enabled {
		return nil
	}

	cfg, err := ResolveBlockConfig(app)
	if err != nil {
		return fmt.Errorf("resolve accounts config: %w", err)
	}
	if !cfg.IsEnabled() {
		return nil
	}

	if err := ApplySchema(ctx, db, cfg); err != nil {
		return fmt.Errorf("apply accounts schema: %w", err)
	}

	return nil
}

// ApplySchemaFromFile loads legacy block YAML and applies the accounts schema.
// Deprecated: prefer ApplySchemaFromAppConfig with the main config accounts node.
func ApplySchemaFromFile(ctx context.Context, db *sql.DB, configPath string) error {
	return ApplySchemaFromAppConfig(ctx, db, AppConfig{Enabled: true, ConfigPath: configPath})
}

// Wire applies the accounts schema and builds the service when app.Enabled is true.
// Returns nil, nil when disabled.
func Wire(ctx context.Context, db *sql.DB, app AppConfig, opts Options) (*Service, error) {
	if !app.Enabled {
		return nil, nil
	}

	blockCfg, err := ResolveBlockConfig(app)
	if err != nil {
		return nil, fmt.Errorf("resolve accounts config: %w", err)
	}

	if !blockCfg.IsEnabled() {
		return nil, nil
	}

	if err := ApplySchema(ctx, db, blockCfg); err != nil {
		return nil, fmt.Errorf("apply accounts schema: %w", err)
	}

	svc, err := New(db, blockCfg, app.secrets(), opts)
	if err != nil {
		return nil, err
	}

	return svc, nil
}
