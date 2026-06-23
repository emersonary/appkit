package permissions

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

func ApplySchema(ctx context.Context, db *sql.DB, setup Setup) error {
	setup.normalize()
	if err := setup.Validate(); err != nil {
		return err
	}

	if err := validateIdent(setup.Schema); err != nil {
		return ErrInvalidSchema.With("schema", err.Error())
	}

	if _, err := db.ExecContext(ctx, fmt.Sprintf(`CREATE SCHEMA IF NOT EXISTS %s`, quoteIdent(setup.Schema))); err != nil {
		return wrapErr(ErrApplySchema, "create_schema", err)
	}

	if _, err := db.ExecContext(ctx, buildSchemaSQL(setup)); err != nil {
		return wrapErr(ErrApplySchema, "exec", err)
	}

	return nil
}

func buildSchemaSQL(setup Setup) string {
	groups := qualifiedName(setup.Schema, "permission_groups")
	categories := qualifiedName(setup.Schema, "permission_categories")
	permissions := qualifiedName(setup.Schema, "permissions")
	profiles := qualifiedName(setup.Schema, "profiles")
	profilePerms := qualifiedName(setup.Schema, "profile_permissions")
	setUpdatedAt := qualifiedName(setup.Schema, "set_updated_at")

	var b strings.Builder
	b.WriteString(`
CREATE OR REPLACE FUNCTION `)
	b.WriteString(setUpdatedAt)
	b.WriteString(`()
RETURNS TRIGGER AS $$
BEGIN
	NEW.updated_at = NOW();
	RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TABLE IF NOT EXISTS `)
	b.WriteString(groups)
	b.WriteString(` (
	id_permission_group TEXT PRIMARY KEY,
	name TEXT NOT NULL,
	sort_order INTEGER NOT NULL DEFAULT 0,
	route_prefix TEXT NOT NULL DEFAULT '',
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS `)
	b.WriteString(categories)
	b.WriteString(` (
	id_permission_category TEXT PRIMARY KEY,
	name TEXT NOT NULL,
	id_permission_group TEXT NOT NULL REFERENCES `)
	b.WriteString(groups)
	b.WriteString(` (id_permission_group) ON DELETE CASCADE,
	sort_order INTEGER NOT NULL DEFAULT 0,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS permission_categories_group_idx ON `)
	b.WriteString(categories)
	b.WriteString(` (id_permission_group, sort_order);

CREATE TABLE IF NOT EXISTS `)
	b.WriteString(permissions)
	b.WriteString(` (
	id_permission TEXT PRIMARY KEY,
	name TEXT NOT NULL UNIQUE,
	id_permission_category TEXT NOT NULL REFERENCES `)
	b.WriteString(categories)
	b.WriteString(` (id_permission_category) ON DELETE CASCADE,
	id_parent TEXT REFERENCES `)
	b.WriteString(permissions)
	b.WriteString(` (id_permission) ON DELETE CASCADE,
	be_action INTEGER NOT NULL DEFAULT 0 CHECK (be_action >= 0),
	route_name TEXT NOT NULL DEFAULT '',
	icon TEXT NOT NULL DEFAULT '',
	enabled BOOLEAN NOT NULL DEFAULT true,
	sort_order INTEGER NOT NULL DEFAULT 0,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS permissions_category_idx ON `)
	b.WriteString(permissions)
	b.WriteString(` (id_permission_category, sort_order);

CREATE INDEX IF NOT EXISTS permissions_parent_idx ON `)
	b.WriteString(permissions)
	b.WriteString(` (id_parent);

`)
	b.WriteString(buildMigratePermissionIconColumnSQL(setup.Schema))

	b.WriteString(`CREATE TABLE IF NOT EXISTS `)
	b.WriteString(profiles)
	b.WriteString(` (
	id_profile TEXT PRIMARY KEY,
	name TEXT NOT NULL,
	is_superuser BOOLEAN NOT NULL DEFAULT false,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS `)
	b.WriteString(profilePerms)
	b.WriteString(` (
	id_profile TEXT NOT NULL REFERENCES `)
	b.WriteString(profiles)
	b.WriteString(` (id_profile) ON DELETE CASCADE,
	id_permission TEXT NOT NULL REFERENCES `)
	b.WriteString(permissions)
	b.WriteString(` (id_permission) ON DELETE CASCADE,
	granted_actions INTEGER CHECK (granted_actions IS NULL OR granted_actions >= 0),
	PRIMARY KEY (id_profile, id_permission)
);

DROP TRIGGER IF EXISTS permission_groups_set_updated_at ON `)
	b.WriteString(groups)
	b.WriteString(`;
CREATE TRIGGER permission_groups_set_updated_at
	BEFORE UPDATE ON `)
	b.WriteString(groups)
	b.WriteString(`
	FOR EACH ROW EXECUTE FUNCTION `)
	b.WriteString(setUpdatedAt)
	b.WriteString(`();

DROP TRIGGER IF EXISTS permission_categories_set_updated_at ON `)
	b.WriteString(categories)
	b.WriteString(`;
CREATE TRIGGER permission_categories_set_updated_at
	BEFORE UPDATE ON `)
	b.WriteString(categories)
	b.WriteString(`
	FOR EACH ROW EXECUTE FUNCTION `)
	b.WriteString(setUpdatedAt)
	b.WriteString(`();

DROP TRIGGER IF EXISTS permissions_set_updated_at ON `)
	b.WriteString(permissions)
	b.WriteString(`;
CREATE TRIGGER permissions_set_updated_at
	BEFORE UPDATE ON `)
	b.WriteString(permissions)
	b.WriteString(`
	FOR EACH ROW EXECUTE FUNCTION `)
	b.WriteString(setUpdatedAt)
	b.WriteString(`();

DROP TRIGGER IF EXISTS profiles_set_updated_at ON `)
	b.WriteString(profiles)
	b.WriteString(`;
CREATE TRIGGER profiles_set_updated_at
	BEFORE UPDATE ON `)
	b.WriteString(profiles)
	b.WriteString(`
	FOR EACH ROW EXECUTE FUNCTION `)
	b.WriteString(setUpdatedAt)
	b.WriteString(`();
`)

	b.WriteString(buildAccountsProfileColumnSQL(setup, profiles))

	if !setup.SkipSeed {
		b.WriteString(buildSeedSQL(setup, groups, categories, permissions, profiles, profilePerms))
	}

	return b.String()
}

func buildAccountsProfileColumnSQL(setup Setup, profiles string) string {
	accounts := qualifiedName(setup.AccountsSchema, "accounts")
	return fmt.Sprintf(`
DO $perm_accounts$
BEGIN
	IF EXISTS (
		SELECT 1 FROM information_schema.tables
		WHERE table_schema = %s AND table_name = 'accounts'
	) AND NOT EXISTS (
		SELECT 1 FROM information_schema.columns
		WHERE table_schema = %s AND table_name = 'accounts' AND column_name = 'id_profile'
	) THEN
		ALTER TABLE %s ADD COLUMN id_profile TEXT;
	END IF;

	IF EXISTS (
		SELECT 1 FROM information_schema.columns
		WHERE table_schema = %s AND table_name = 'accounts' AND column_name = 'id_profile'
	) AND NOT EXISTS (
		SELECT 1 FROM information_schema.table_constraints
		WHERE constraint_schema = %s
		  AND table_name = 'accounts'
		  AND constraint_name = 'accounts_id_profile_fkey'
	) THEN
		ALTER TABLE %s
			ADD CONSTRAINT accounts_id_profile_fkey
			FOREIGN KEY (id_profile) REFERENCES %s (id_profile) ON DELETE SET NULL;
	END IF;
END $perm_accounts$;
`, quoteLiteral(setup.AccountsSchema), quoteLiteral(setup.AccountsSchema), accounts,
		quoteLiteral(setup.AccountsSchema), quoteLiteral(setup.AccountsSchema), accounts, profiles)
}

func buildMigratePermissionIconColumnSQL(schema string) string {
	table := qualifiedName(schema, "permissions")
	return fmt.Sprintf(`
DO $migrate$
BEGIN
	IF EXISTS (
		SELECT 1 FROM information_schema.tables
		WHERE table_schema = %s AND table_name = 'permissions'
	) AND NOT EXISTS (
		SELECT 1 FROM information_schema.columns
		WHERE table_schema = %s AND table_name = 'permissions' AND column_name = 'icon'
	) THEN
		ALTER TABLE %s ADD COLUMN icon TEXT NOT NULL DEFAULT '';
	END IF;
END $migrate$;
`, quoteLiteral(schema), quoteLiteral(schema), table)
}
