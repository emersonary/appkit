package tenants

import (
	"context"
	"database/sql"
	"fmt"
)

// ApplySchemaFromFile loads block YAML and applies the tenants schema.
func ApplySchemaFromFile(ctx context.Context, db *sql.DB, configPath string) error {
	path := configPath
	if path == "" {
		path = defaultAppConfigPath
	}

	cfg, err := LoadConfig(path)
	if err != nil {
		return fmt.Errorf("load tenants config: %w", err)
	}

	if err := ApplySchema(ctx, db, cfg); err != nil {
		return fmt.Errorf("apply tenants schema: %w", err)
	}

	return nil
}

// Wire applies the tenants schema and builds the service when app.Enabled is true.
func Wire(ctx context.Context, db *sql.DB, app AppConfig) (*Service, error) {
	if !app.Enabled {
		return nil, nil
	}

	app.ApplyDefaults()

	blockCfg, err := ResolveBlockConfig(app)
	if err != nil {
		return nil, fmt.Errorf("load tenants config: %w", err)
	}

	if err := ApplySchema(ctx, db, blockCfg); err != nil {
		return nil, fmt.Errorf("apply tenants schema: %w", err)
	}

	svc, err := New(db, blockCfg)
	if err != nil {
		return nil, err
	}

	if _, err := svc.SyncFixedCatalog(ctx); err != nil {
		return nil, fmt.Errorf("sync fixed tenant catalog: %w", err)
	}

	return svc, nil
}
