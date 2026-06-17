package tenants

import (
	"context"
	"database/sql"
	"fmt"
)

func ApplySchema(ctx context.Context, db *sql.DB, cfg Config) error {
	cfg.normalize()
	if err := cfg.Validate(); err != nil {
		return err
	}

	if _, err := db.ExecContext(ctx, fmt.Sprintf(`CREATE SCHEMA IF NOT EXISTS %s`, quoteIdent(cfg.Schema))); err != nil {
		return wrapErr(ErrApplySchema, "create_schema", err)
	}

	if _, err := db.ExecContext(ctx, buildSchemaSQL(cfg)); err != nil {
		return wrapErr(ErrApplySchema, "exec", err)
	}

	return nil
}

func buildSchemaSQL(cfg Config) string {
	schema := cfg.Schema
	tenants := qualifiedName(schema, "tenants")
	accounts := qualifiedName(schema, "tenant_accounts")
	invites := qualifiedName(schema, "tenant_invites")
	setUpdatedAt := qualifiedName(schema, "set_updated_at")

	return `
CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE OR REPLACE FUNCTION ` + setUpdatedAt + `()
RETURNS TRIGGER AS $$
BEGIN
	NEW.updated_at = NOW();
	RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TABLE IF NOT EXISTS ` + tenants + ` (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	slug TEXT NOT NULL UNIQUE,
	name TEXT NOT NULL,
	timezone TEXT NOT NULL DEFAULT 'UTC',
	status TEXT NOT NULL DEFAULT 'trial' CHECK (status IN ('trial', 'active', 'suspended')),
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS ` + accounts + ` (
	tenant_id UUID NOT NULL REFERENCES ` + tenants + ` (id) ON DELETE CASCADE,
	account_id UUID NOT NULL,
	role TEXT NOT NULL CHECK (role IN ('owner', 'admin', 'staff', 'viewer')),
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	PRIMARY KEY (tenant_id, account_id)
);

CREATE INDEX IF NOT EXISTS idx_tenant_accounts_account ON ` + accounts + ` (account_id);

CREATE TABLE IF NOT EXISTS ` + invites + ` (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	tenant_id UUID NOT NULL REFERENCES ` + tenants + ` (id) ON DELETE CASCADE,
	email TEXT NOT NULL,
	role TEXT NOT NULL CHECK (role IN ('owner', 'admin', 'staff', 'viewer')),
	token_hash TEXT NOT NULL UNIQUE,
	expires_at TIMESTAMPTZ NOT NULL,
	accepted_at TIMESTAMPTZ,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_tenant_invites_tenant ON ` + invites + ` (tenant_id);

DROP TRIGGER IF EXISTS tenants_set_updated_at ON ` + tenants + `;
CREATE TRIGGER tenants_set_updated_at
	BEFORE UPDATE ON ` + tenants + `
	FOR EACH ROW
	EXECUTE PROCEDURE ` + setUpdatedAt + `();

DROP TRIGGER IF EXISTS tenant_accounts_set_updated_at ON ` + accounts + `;
CREATE TRIGGER tenant_accounts_set_updated_at
	BEFORE UPDATE ON ` + accounts + `
	FOR EACH ROW
	EXECUTE PROCEDURE ` + setUpdatedAt + `();
`
}
