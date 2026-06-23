package language

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

func ApplySchema(ctx context.Context, db *sql.DB, cfg LanguageConfig) error {
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

func buildSchemaSQL(cfg LanguageConfig) string {
	languages := qualifiedName(cfg.Schema, "languages")
	setUpdatedAt := qualifiedName(cfg.Schema, "set_updated_at")

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
	b.WriteString(languages)
	b.WriteString(` (
	id_language CHAR(2) PRIMARY KEY,
	name TEXT NOT NULL,
	native_name TEXT NOT NULL DEFAULT '',
	direction TEXT NOT NULL DEFAULT 'ltr' CHECK (direction IN ('ltr', 'rtl')),
	is_default BOOLEAN NOT NULL DEFAULT false,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	CONSTRAINT languages_id_language_lower CHECK (id_language ~ '^[a-z]{2}$')
);

CREATE UNIQUE INDEX IF NOT EXISTS languages_one_default_idx ON `)
	b.WriteString(languages)
	b.WriteString(` (is_default) WHERE is_default = true;

DROP TRIGGER IF EXISTS languages_set_updated_at ON `)
	b.WriteString(languages)
	b.WriteString(`;

CREATE TRIGGER languages_set_updated_at
	BEFORE UPDATE ON `)
	b.WriteString(languages)
	b.WriteString(`
	FOR EACH ROW EXECUTE FUNCTION `)
	b.WriteString(setUpdatedAt)
	b.WriteString(`();
`)

	if !cfg.SkipSeed {
		b.WriteString(buildSeedSQL(languages, cfg))
	}

	return b.String()
}

func buildSeedSQL(languages string, cfg LanguageConfig) string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("\nINSERT INTO %s (id_language, name, native_name, direction, is_default) VALUES\n", languages))

	values := make([]string, 0, len(cfg.Languages))
	for _, code := range cfg.Languages {
		entry, ok := LanguageCatalog[code]
		if !ok {
			continue
		}

		direction := entry.Direction
		if direction == "" {
			direction = "ltr"
		}

		values = append(values, fmt.Sprintf(
			"\t(%s, %s, %s, %s, %t)",
			quoteLiteral(code),
			quoteLiteral(entry.Name),
			quoteLiteral(entry.NativeName),
			quoteLiteral(direction),
			code == cfg.DefaultLanguage,
		))
	}

	b.WriteString(strings.Join(values, ",\n"))
	b.WriteString("\nON CONFLICT (id_language) DO UPDATE SET\n")
	b.WriteString("\tname = EXCLUDED.name,\n")
	b.WriteString("\tnative_name = EXCLUDED.native_name,\n")
	b.WriteString("\tdirection = EXCLUDED.direction,\n")
	b.WriteString("\tis_default = EXCLUDED.is_default;\n")

	return b.String()
}
