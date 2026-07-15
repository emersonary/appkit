package language

import (
	"context"
	"database/sql"
	"fmt"

	"go.uber.org/zap"
)

// WireOptions configures optional language wiring behavior.
type WireOptions struct {
	Logger *zap.Logger
}

// ApplySchemaFromLanguageConfig resolves language config and applies the schema when enabled.
func ApplySchemaFromLanguageConfig(ctx context.Context, db *sql.DB, app LanguageConfig) error {
	if !app.Enabled {
		return nil
	}

	cfg, err := ResolveBlockConfig(app)
	if err != nil {
		return fmt.Errorf("resolve language config: %w", err)
	}

	if err := ApplySchema(ctx, db, cfg); err != nil {
		return fmt.Errorf("apply language schema: %w", err)
	}

	return nil
}

// ApplySchemaFromAppConfig is deprecated; use ApplySchemaFromLanguageConfig.
func ApplySchemaFromAppConfig(ctx context.Context, db *sql.DB, app LanguageConfig) error {
	return ApplySchemaFromLanguageConfig(ctx, db, app)
}

// ApplySchemaFromFile loads legacy block YAML and applies the language schema.
// Deprecated: prefer ApplySchemaFromLanguageConfig with the main config language node.
func ApplySchemaFromFile(ctx context.Context, db *sql.DB, configPath string) error {
	return ApplySchemaFromLanguageConfig(ctx, db, LanguageConfig{Enabled: true, ConfigPath: configPath})
}

// Wire applies the language schema and builds the service when app.Enabled is true.
// Returns nil, nil when disabled.
func Wire(ctx context.Context, db *sql.DB, app LanguageConfig, opts WireOptions) (*Service, error) {
	if !app.Enabled {
		return nil, nil
	}

	blockCfg, err := ResolveBlockConfig(app)
	if err != nil {
		return nil, fmt.Errorf("resolve language config: %w", err)
	}

	if err := ApplySchema(ctx, db, blockCfg); err != nil {
		return nil, fmt.Errorf("apply language schema: %w", err)
	}

	svc, err := NewService(db, blockCfg, Options{Logger: opts.Logger})
	if err != nil {
		return nil, err
	}

	svc.logger.Info("language block enabled",
		zap.String("default", blockCfg.DefaultLanguage),
	)
	return svc, nil
}
