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
	languages := qualifiedName(schema, "tenant_languages")
	resourceTypes := qualifiedName(schema, "tenant_resource_types")
	resourceTypeTranslations := qualifiedName(schema, "tenant_resource_type_translations")
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
	business_type_id UUID,
	default_language_id UUID,
	timezone TEXT NOT NULL DEFAULT 'UTC',
	status TEXT NOT NULL DEFAULT 'trial' CHECK (status IN ('trial', 'active', 'suspended')),
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS ` + accounts + ` (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	tenant_id UUID NOT NULL REFERENCES ` + tenants + ` (id) ON DELETE CASCADE,
	account_id UUID NOT NULL,
	role TEXT NOT NULL CHECK (role IN ('owner', 'admin', 'staff', 'viewer')),
	resource_id UUID,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	UNIQUE (tenant_id, account_id)
);

CREATE INDEX IF NOT EXISTS idx_tenant_accounts_account ON ` + accounts + ` (account_id);
CREATE INDEX IF NOT EXISTS idx_tenant_accounts_tenant ON ` + accounts + ` (tenant_id);
CREATE INDEX IF NOT EXISTS idx_tenant_accounts_resource ON ` + accounts + ` (resource_id);

CREATE TABLE IF NOT EXISTS ` + invites + ` (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	tenant_id UUID NOT NULL REFERENCES ` + tenants + ` (id) ON DELETE CASCADE,
	email TEXT NOT NULL,
	role TEXT NOT NULL CHECK (role IN ('owner', 'admin', 'staff', 'viewer')),
	token_hash TEXT NOT NULL UNIQUE,
	invited_by_account_id UUID,
	expires_at TIMESTAMPTZ NOT NULL,
	accepted_at TIMESTAMPTZ,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS tenant_invites_pending_email
	ON ` + invites + ` (tenant_id, lower(email))
	WHERE accepted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_tenant_invites_tenant ON ` + invites + ` (tenant_id);
CREATE INDEX IF NOT EXISTS idx_tenant_invites_email_pending
	ON ` + invites + ` (lower(email))
	WHERE accepted_at IS NULL;

CREATE TABLE IF NOT EXISTS ` + languages + ` (
	tenant_id UUID NOT NULL REFERENCES ` + tenants + ` (id) ON DELETE CASCADE,
	language_id UUID NOT NULL,
	is_default BOOLEAN NOT NULL DEFAULT FALSE,
	is_enabled BOOLEAN NOT NULL DEFAULT TRUE,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	PRIMARY KEY (tenant_id, language_id)
);

CREATE TABLE IF NOT EXISTS ` + resourceTypes + ` (
	tenant_id UUID NOT NULL REFERENCES ` + tenants + ` (id) ON DELETE CASCADE,
	resource_type_id UUID NOT NULL,
	is_enabled BOOLEAN NOT NULL DEFAULT TRUE,
	is_default BOOLEAN NOT NULL DEFAULT FALSE,
	sort_order INT NOT NULL DEFAULT 0,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	PRIMARY KEY (tenant_id, resource_type_id)
);

CREATE TABLE IF NOT EXISTS ` + resourceTypeTranslations + ` (
	tenant_id UUID NOT NULL,
	resource_type_id UUID NOT NULL,
	language_id UUID NOT NULL,
	display_name TEXT NOT NULL,
	description TEXT,
	PRIMARY KEY (tenant_id, resource_type_id, language_id),
	FOREIGN KEY (tenant_id, resource_type_id)
		REFERENCES ` + resourceTypes + ` (tenant_id, resource_type_id) ON DELETE CASCADE
);

ALTER TABLE ` + tenants + ` ADD COLUMN IF NOT EXISTS business_type_id UUID;
ALTER TABLE ` + tenants + ` ADD COLUMN IF NOT EXISTS default_language_id UUID;

ALTER TABLE ` + accounts + ` ADD COLUMN IF NOT EXISTS id UUID DEFAULT gen_random_uuid();
ALTER TABLE ` + accounts + ` ADD COLUMN IF NOT EXISTS resource_id UUID;

ALTER TABLE ` + invites + ` ADD COLUMN IF NOT EXISTS invited_by_account_id UUID;

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
