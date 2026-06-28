package accounts

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
	if err := validateIdent(cfg.Schema); err != nil {
		return ErrInvalidSchema.With("schema", err.Error())
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
	accounts := qualifiedName(schema, "accounts")
	oauth := qualifiedName(schema, "oauth_identities")
	verify := qualifiedName(schema, "email_verification_tokens")
	reset := qualifiedName(schema, "password_reset_tokens")
	setUpdatedAt := qualifiedName(schema, "set_updated_at")

	var b string
	b += `
CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE OR REPLACE FUNCTION ` + setUpdatedAt + `()
RETURNS TRIGGER AS $$
BEGIN
	NEW.updated_at = NOW();
	RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TABLE IF NOT EXISTS ` + accounts + ` (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	email TEXT NOT NULL,
	password_hash TEXT,
	first_name TEXT,
	last_name TEXT,
	avatar_url TEXT,
	email_verified_at TIMESTAMPTZ,
	is_admin BOOLEAN NOT NULL DEFAULT false,
	id_profile TEXT,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	CONSTRAINT accounts_email_unique UNIQUE (email)
);

ALTER TABLE ` + accounts + ` ADD COLUMN IF NOT EXISTS id_profile TEXT;

CREATE TABLE IF NOT EXISTS ` + oauth + ` (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	account_id UUID NOT NULL REFERENCES ` + accounts + ` (id) ON DELETE CASCADE,
	provider TEXT NOT NULL,
	provider_user_id TEXT NOT NULL,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	CONSTRAINT oauth_identities_provider_user_unique UNIQUE (provider, provider_user_id)
);

CREATE INDEX IF NOT EXISTS idx_oauth_identities_account_id ON ` + oauth + ` (account_id);

CREATE TABLE IF NOT EXISTS ` + verify + ` (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	account_id UUID NOT NULL REFERENCES ` + accounts + ` (id) ON DELETE CASCADE,
	token_hash TEXT NOT NULL,
	expires_at TIMESTAMPTZ NOT NULL,
	used_at TIMESTAMPTZ,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_email_verification_tokens_account_id ON ` + verify + ` (account_id);

CREATE TABLE IF NOT EXISTS ` + reset + ` (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	account_id UUID NOT NULL REFERENCES ` + accounts + ` (id) ON DELETE CASCADE,
	token_hash TEXT NOT NULL,
	expires_at TIMESTAMPTZ NOT NULL,
	used_at TIMESTAMPTZ,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_password_reset_tokens_account_id ON ` + reset + ` (account_id);

DROP TRIGGER IF EXISTS accounts_set_updated_at ON ` + accounts + `;
CREATE TRIGGER accounts_set_updated_at
	BEFORE UPDATE ON ` + accounts + `
	FOR EACH ROW
	EXECUTE PROCEDURE ` + setUpdatedAt + `();
`

	if cfg.Tenancy.Enabled {
		tenants := qualifiedName(schema, "tenants")
		accountTenants := qualifiedName(schema, "account_tenants")
		b += `
CREATE TABLE IF NOT EXISTS ` + tenants + ` (
	id TEXT PRIMARY KEY,
	name TEXT NOT NULL,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS ` + accountTenants + ` (
	account_id UUID NOT NULL REFERENCES ` + accounts + ` (id) ON DELETE CASCADE,
	tenant_id TEXT NOT NULL REFERENCES ` + tenants + ` (id) ON DELETE CASCADE,
	role TEXT NOT NULL DEFAULT 'member' CHECK (role IN ('owner', 'admin', 'member')),
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	PRIMARY KEY (account_id, tenant_id)
);

CREATE INDEX IF NOT EXISTS idx_account_tenants_tenant_id ON ` + accountTenants + ` (tenant_id);

DROP TRIGGER IF EXISTS tenants_set_updated_at ON ` + tenants + `;
CREATE TRIGGER tenants_set_updated_at
	BEFORE UPDATE ON ` + tenants + `
	FOR EACH ROW
	EXECUTE PROCEDURE ` + setUpdatedAt + `();
`
	}

	return b
}
