package social

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

// SchemaConfig controls block-owned OAuth connection storage.
type SchemaConfig struct {
	Schema string
}

func (c SchemaConfig) normalized() SchemaConfig {
	out := c
	if strings.TrimSpace(out.Schema) == "" {
		out.Schema = "social"
	}
	return out
}

// ApplyConnectionSchema creates the social block connection table when not using a host migration.
func ApplyConnectionSchema(ctx context.Context, db *sql.DB, cfg SchemaConfig) error {
	cfg = cfg.normalized()
	schema := quoteIdent(cfg.Schema)
	table := schema + ".platform_connections"

	if _, err := db.ExecContext(ctx, fmt.Sprintf(`CREATE SCHEMA IF NOT EXISTS %s`, schema)); err != nil {
		return fmt.Errorf("create social schema: %w", err)
	}

	ddl := fmt.Sprintf(`
CREATE TABLE IF NOT EXISTS %s (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	tenant_id TEXT NOT NULL,
	platform_id TEXT NOT NULL,
	id_language CHAR(2) NOT NULL,
	account_id TEXT NOT NULL DEFAULT '',
	access_token TEXT NOT NULL,
	expires_at TIMESTAMPTZ NOT NULL,
	scopes TEXT NOT NULL DEFAULT '',
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	CONSTRAINT social_platform_connections_tenant_platform_language_key UNIQUE (tenant_id, platform_id, id_language)
);

CREATE INDEX IF NOT EXISTS idx_social_platform_connections_expires ON %s(expires_at);
`, table, table)

	if _, err := db.ExecContext(ctx, ddl); err != nil {
		return fmt.Errorf("apply social connection schema: %w", err)
	}
	return nil
}

func quoteIdent(name string) string {
	return `"` + strings.ReplaceAll(strings.TrimSpace(name), `"`, `""`) + `"`
}
