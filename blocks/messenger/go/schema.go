package messenger

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
	users := qualifiedName(schema, "tbl_users")
	devices := qualifiedName(schema, "tbl_devices")
	channels := qualifiedName(schema, "tbl_channels")
	direct := qualifiedName(schema, "tbl_direct_channels")
	members := qualifiedName(schema, "tbl_channel_members")
	messages := qualifiedName(schema, "tbl_messages")
	handoffs := qualifiedName(schema, "tbl_channel_handoffs")
	setUpdatedAt := qualifiedName(schema, "func_set_updated_at")
	accounts := cfg.accountsTable()
	tenants := cfg.tenantsTable()

	return `
CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE OR REPLACE FUNCTION ` + setUpdatedAt + `()
RETURNS TRIGGER AS $$
BEGIN
	NEW.dt_updated = NOW();
	RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TABLE IF NOT EXISTS ` + users + ` (
	id UUID PRIMARY KEY DEFAULT platform.func_gen_uuid_v7(),
	id_tenant UUID NOT NULL REFERENCES ` + tenants + ` (id) ON DELETE CASCADE,
	str_user_type CHAR(1) NOT NULL CHECK (str_user_type IN ('H', 'A', 'S')),
	id_account UUID REFERENCES ` + accounts + ` (id) ON DELETE SET NULL,
	str_display_name TEXT NOT NULL,
	str_avatar_url TEXT,
	str_external_ref TEXT,
	str_row_status CHAR(1) NOT NULL DEFAULT 'A' CHECK (str_row_status IN ('A', 'I', 'D', 'T')),
	dt_created TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	dt_updated TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	CONSTRAINT tbl_users_type_account_check CHECK (
		(str_user_type = 'H' AND id_account IS NOT NULL)
		OR (str_user_type IN ('A', 'S') AND id_account IS NULL)
	)
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_tbl_users_tenant_account
	ON ` + users + ` (id_tenant, id_account)
	WHERE id_account IS NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS idx_tbl_users_tenant_external_ref
	ON ` + users + ` (id_tenant, str_external_ref)
	WHERE str_external_ref IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_tbl_users_tenant_type
	ON ` + users + ` (id_tenant, str_user_type);

CREATE TABLE IF NOT EXISTS ` + devices + ` (
	id UUID PRIMARY KEY DEFAULT platform.func_gen_uuid_v7(),
	id_messenger_user UUID NOT NULL REFERENCES ` + users + ` (id) ON DELETE CASCADE,
	str_device_key TEXT NOT NULL,
	str_platform TEXT NOT NULL,
	str_push_token TEXT,
	dt_last_seen TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	str_row_status CHAR(1) NOT NULL DEFAULT 'A' CHECK (str_row_status IN ('A', 'I', 'D', 'T')),
	dt_created TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	dt_updated TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	UNIQUE (id_messenger_user, str_device_key)
);

CREATE TABLE IF NOT EXISTS ` + channels + ` (
	id UUID PRIMARY KEY DEFAULT platform.func_gen_uuid_v7(),
	id_tenant UUID NOT NULL REFERENCES ` + tenants + ` (id) ON DELETE CASCADE,
	str_kind CHAR(1) NOT NULL CHECK (str_kind IN ('D', 'G')),
	str_title TEXT,
	str_avatar_url TEXT,
	id_created_by UUID NOT NULL REFERENCES ` + users + ` (id) ON DELETE RESTRICT,
	str_row_status CHAR(1) NOT NULL DEFAULT 'A' CHECK (str_row_status IN ('A', 'I', 'D', 'T')),
	dt_created TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	dt_updated TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_tbl_channels_tenant
	ON ` + channels + ` (id_tenant, dt_updated DESC);

CREATE TABLE IF NOT EXISTS ` + direct + ` (
	id_tenant UUID NOT NULL REFERENCES ` + tenants + ` (id) ON DELETE CASCADE,
	id_messenger_user_a UUID NOT NULL REFERENCES ` + users + ` (id) ON DELETE CASCADE,
	id_messenger_user_b UUID NOT NULL REFERENCES ` + users + ` (id) ON DELETE CASCADE,
	id_channel UUID NOT NULL UNIQUE REFERENCES ` + channels + ` (id) ON DELETE CASCADE,
	PRIMARY KEY (id_tenant, id_messenger_user_a, id_messenger_user_b)
);

CREATE TABLE IF NOT EXISTS ` + members + ` (
	id UUID PRIMARY KEY DEFAULT platform.func_gen_uuid_v7(),
	id_channel UUID NOT NULL REFERENCES ` + channels + ` (id) ON DELETE CASCADE,
	id_messenger_user UUID NOT NULL REFERENCES ` + users + ` (id) ON DELETE CASCADE,
	str_role TEXT NOT NULL DEFAULT 'member' CHECK (str_role IN ('member', 'admin')),
	id_last_read_message UUID,
	dt_last_read TIMESTAMPTZ,
	i_unread_count INT NOT NULL DEFAULT 0,
	str_row_status CHAR(1) NOT NULL DEFAULT 'A' CHECK (str_row_status IN ('A', 'I', 'D', 'T')),
	dt_created TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	dt_updated TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	UNIQUE (id_channel, id_messenger_user)
);

CREATE INDEX IF NOT EXISTS idx_tbl_channel_members_user
	ON ` + members + ` (id_messenger_user);

CREATE TABLE IF NOT EXISTS ` + messages + ` (
	id UUID PRIMARY KEY DEFAULT platform.func_gen_uuid_v7(),
	id_channel UUID NOT NULL REFERENCES ` + channels + ` (id) ON DELETE CASCADE,
	id_sender_messenger_user UUID REFERENCES ` + users + ` (id) ON DELETE SET NULL,
	str_kind CHAR(1) NOT NULL CHECK (str_kind IN ('U', 'S')),
	str_event_type TEXT,
	jsn_event_payload JSONB,
	str_body TEXT,
	str_client_ref TEXT,
	id_reply_to UUID REFERENCES ` + messages + ` (id) ON DELETE SET NULL,
	str_row_status CHAR(1) NOT NULL DEFAULT 'A' CHECK (str_row_status IN ('A', 'I', 'D', 'T')),
	dt_created TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	dt_updated TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_tbl_messages_channel_created
	ON ` + messages + ` (id_channel, dt_created DESC);

CREATE UNIQUE INDEX IF NOT EXISTS idx_tbl_messages_client_ref
	ON ` + messages + ` (id_channel, str_client_ref)
	WHERE str_client_ref IS NOT NULL;

CREATE TABLE IF NOT EXISTS ` + qualifiedName(schema, "tbl_message_receipts") + ` (
	id UUID PRIMARY KEY DEFAULT platform.func_gen_uuid_v7(),
	id_message UUID NOT NULL REFERENCES ` + messages + ` (id) ON DELETE CASCADE,
	id_channel UUID NOT NULL REFERENCES ` + channels + ` (id) ON DELETE CASCADE,
	id_messenger_user UUID NOT NULL REFERENCES ` + users + ` (id) ON DELETE CASCADE,
	id_device UUID REFERENCES ` + devices + ` (id) ON DELETE SET NULL,
	dt_read TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	UNIQUE (id_message, id_messenger_user)
);

CREATE INDEX IF NOT EXISTS idx_tbl_message_receipts_channel
	ON ` + qualifiedName(schema, "tbl_message_receipts") + ` (id_channel, dt_read DESC);

CREATE TABLE IF NOT EXISTS ` + handoffs + ` (
	id UUID PRIMARY KEY DEFAULT platform.func_gen_uuid_v7(),
	id_channel UUID NOT NULL REFERENCES ` + channels + ` (id) ON DELETE CASCADE,
	id_from_messenger_user UUID NOT NULL REFERENCES ` + users + ` (id) ON DELETE RESTRICT,
	id_to_messenger_user UUID NOT NULL REFERENCES ` + users + ` (id) ON DELETE RESTRICT,
	str_reason TEXT,
	str_context_summary TEXT,
	dt_created TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

DROP TRIGGER IF EXISTS tbl_users_set_updated_at ON ` + users + `;
CREATE TRIGGER tbl_users_set_updated_at
	BEFORE UPDATE ON ` + users + `
	FOR EACH ROW EXECUTE PROCEDURE ` + setUpdatedAt + `();

DROP TRIGGER IF EXISTS tbl_devices_set_updated_at ON ` + devices + `;
CREATE TRIGGER tbl_devices_set_updated_at
	BEFORE UPDATE ON ` + devices + `
	FOR EACH ROW EXECUTE PROCEDURE ` + setUpdatedAt + `();

DROP TRIGGER IF EXISTS tbl_channels_set_updated_at ON ` + channels + `;
CREATE TRIGGER tbl_channels_set_updated_at
	BEFORE UPDATE ON ` + channels + `
	FOR EACH ROW EXECUTE PROCEDURE ` + setUpdatedAt + `();

DROP TRIGGER IF EXISTS tbl_channel_members_set_updated_at ON ` + members + `;
CREATE TRIGGER tbl_channel_members_set_updated_at
	BEFORE UPDATE ON ` + members + `
	FOR EACH ROW EXECUTE PROCEDURE ` + setUpdatedAt + `();

DROP TRIGGER IF EXISTS tbl_messages_set_updated_at ON ` + messages + `;
CREATE TRIGGER tbl_messages_set_updated_at
	BEFORE UPDATE ON ` + messages + `
	FOR EACH ROW EXECUTE PROCEDURE ` + setUpdatedAt + `();
`
}
