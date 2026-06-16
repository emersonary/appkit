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

	if err := migrateLegacyPublicAuth(ctx, db, cfg); err != nil {
		return wrapErr(ErrApplySchema, "migrate_legacy", err)
	}

	if err := migrateAccountNameColumns(ctx, db, cfg); err != nil {
		return wrapErr(ErrApplySchema, "migrate_names", err)
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
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	CONSTRAINT accounts_email_unique UNIQUE (email)
);

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

func migrateLegacyPublicAuth(ctx context.Context, db *sql.DB, cfg Config) error {
	var publicUsers bool
	err := db.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM information_schema.tables
			WHERE table_schema = 'public' AND table_name = 'users'
		)
	`).Scan(&publicUsers)
	if err != nil {
		return err
	}
	if !publicUsers {
		return nil
	}

	schema := quoteIdent(cfg.Schema)
	stmts := []string{
		`ALTER TABLE public.users SET SCHEMA ` + schema,
		`ALTER TABLE ` + schema + `.users RENAME TO accounts`,
	}
	if tableExists(ctx, db, "public", "oauth_accounts") {
		stmts = append(stmts,
			`ALTER TABLE public.oauth_accounts SET SCHEMA `+schema,
			`ALTER TABLE `+schema+`.oauth_accounts RENAME TO oauth_identities`,
			`ALTER TABLE `+schema+`.oauth_identities RENAME COLUMN user_id TO account_id`,
		)
	}
	if tableExists(ctx, db, "public", "email_verification_tokens") {
		stmts = append(stmts,
			`ALTER TABLE public.email_verification_tokens SET SCHEMA `+schema,
			`ALTER TABLE `+schema+`.email_verification_tokens RENAME COLUMN user_id TO account_id`,
		)
	}
	if tableExists(ctx, db, "public", "password_reset_tokens") {
		stmts = append(stmts,
			`ALTER TABLE public.password_reset_tokens SET SCHEMA `+schema,
			`ALTER TABLE `+schema+`.password_reset_tokens RENAME COLUMN user_id TO account_id`,
		)
	}

	for _, stmt := range stmts {
		if _, err := db.ExecContext(ctx, stmt); err != nil {
			// Idempotent: ignore if already migrated.
			continue
		}
	}
	return rebindMembershipFK(ctx, db, cfg.Schema)
}

func tableExists(ctx context.Context, db *sql.DB, schema, table string) bool {
	var ok bool
	_ = db.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM information_schema.tables
			WHERE table_schema = $1 AND table_name = $2
		)
	`, schema, table).Scan(&ok)
	return ok
}

func rebindMembershipFK(ctx context.Context, db *sql.DB, schema string) error {
	if !tableExists(ctx, db, "public", "memberships") {
		return nil
	}
	accounts := qualifiedName(schema, "accounts")
	_, err := db.ExecContext(ctx, `
		ALTER TABLE public.memberships DROP CONSTRAINT IF EXISTS memberships_user_id_fkey;
		ALTER TABLE public.memberships
			ADD CONSTRAINT memberships_user_id_fkey
			FOREIGN KEY (user_id) REFERENCES `+accounts+` (id) ON DELETE CASCADE;
	`)
	return err
}

func migrateAccountNameColumns(ctx context.Context, db *sql.DB, cfg Config) error {
	if !tableExists(ctx, db, cfg.Schema, "accounts") {
		return nil
	}
	accounts := qualifiedName(cfg.Schema, "accounts")

	if !columnExists(ctx, db, cfg.Schema, "accounts", "first_name") {
		if _, err := db.ExecContext(ctx, `ALTER TABLE `+accounts+` ADD COLUMN first_name TEXT`); err != nil {
			return err
		}
	}
	if !columnExists(ctx, db, cfg.Schema, "accounts", "last_name") {
		if _, err := db.ExecContext(ctx, `ALTER TABLE `+accounts+` ADD COLUMN last_name TEXT`); err != nil {
			return err
		}
	}

	if columnExists(ctx, db, cfg.Schema, "accounts", "display_name") {
		_, err := db.ExecContext(ctx, `
			UPDATE `+accounts+`
			SET
				first_name = CASE
					WHEN NULLIF(TRIM(first_name), '') IS NOT NULL THEN first_name
					WHEN NULLIF(TRIM(display_name), '') IS NULL THEN first_name
					WHEN POSITION(' ' IN TRIM(display_name)) = 0 THEN TRIM(display_name)
					ELSE SPLIT_PART(TRIM(display_name), ' ', 1)
				END,
				last_name = CASE
					WHEN NULLIF(TRIM(last_name), '') IS NOT NULL THEN last_name
					WHEN NULLIF(TRIM(display_name), '') IS NULL THEN last_name
					WHEN POSITION(' ' IN TRIM(display_name)) = 0 THEN NULL
					ELSE TRIM(SUBSTRING(TRIM(display_name) FROM POSITION(' ' IN TRIM(display_name)) + 1))
				END
			WHERE NULLIF(TRIM(display_name), '') IS NOT NULL
		`)
		if err != nil {
			return err
		}
		_, err = db.ExecContext(ctx, `ALTER TABLE `+accounts+` DROP COLUMN display_name`)
		if err != nil {
			return err
		}
	}

	return nil
}

func columnExists(ctx context.Context, db *sql.DB, schema, table, column string) bool {
	var ok bool
	_ = db.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM information_schema.columns
			WHERE table_schema = $1 AND table_name = $2 AND column_name = $3
		)
	`, schema, table, column).Scan(&ok)
	return ok
}
