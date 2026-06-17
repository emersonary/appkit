package accounts

import (
	"context"
	"database/sql"
	"fmt"
)

// ApplySchemaFromFile loads block YAML and applies the accounts schema.
func ApplySchemaFromFile(ctx context.Context, db *sql.DB, configPath string) error {
	path := configPath
	if path == "" {
		path = defaultAppConfigPath
	}

	cfg, err := LoadConfig(path)
	if err != nil {
		return fmt.Errorf("load accounts config: %w", err)
	}

	if err := ApplySchema(ctx, db, cfg); err != nil {
		return fmt.Errorf("apply accounts schema: %w", err)
	}

	return nil
}

// Wire applies the accounts schema and builds the service when app.Enabled is true.
// Returns nil, nil when disabled.
func Wire(ctx context.Context, db *sql.DB, app AppConfig, opts Options) (*Service, error) {
	if !app.Enabled {
		return nil, nil
	}

	app.ApplyDefaults()

	blockCfg, err := LoadConfig(app.ConfigPath)
	if err != nil {
		return nil, fmt.Errorf("load accounts config: %w", err)
	}
	mergeAppConfig(&blockCfg, app)

	if err := ApplySchema(ctx, db, blockCfg); err != nil {
		return nil, fmt.Errorf("apply accounts schema: %w", err)
	}

	svc, err := New(db, blockCfg, app.secrets(), opts)
	if err != nil {
		return nil, err
	}

	return svc, nil
}
