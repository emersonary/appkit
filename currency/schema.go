package currency

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
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
	currencies := qualifiedName(cfg.Schema, "currencies")
	rates := qualifiedName(cfg.Schema, "currency_exchange_rates")
	history := qualifiedName(cfg.Schema, "currency_exchange_rate_history")
	setUpdatedAt := qualifiedName(cfg.Schema, "set_updated_at")

	var b strings.Builder
	b.WriteString(`
CREATE EXTENSION IF NOT EXISTS pgcrypto;

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
	b.WriteString(currencies)
	b.WriteString(` (
	code TEXT PRIMARY KEY,
	name TEXT NOT NULL,
	symbol TEXT NOT NULL,
	is_base BOOLEAN NOT NULL DEFAULT false,
	website_languages TEXT[] NOT NULL DEFAULT '{}',
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	CONSTRAINT currencies_code_upper CHECK (code ~ '^[A-Z]{3}$')
);

CREATE UNIQUE INDEX IF NOT EXISTS currencies_one_base_idx ON `)
	b.WriteString(currencies)
	b.WriteString(` (is_base) WHERE is_base = true;

CREATE TABLE IF NOT EXISTS `)
	b.WriteString(rates)
	b.WriteString(` (
	currency_code TEXT PRIMARY KEY REFERENCES `)
	b.WriteString(currencies)
	b.WriteString(` (code) ON DELETE CASCADE,
	rate NUMERIC(20, 8) NOT NULL CHECK (rate > 0),
	source TEXT NOT NULL DEFAULT '',
	fetched_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS `)
	b.WriteString(history)
	b.WriteString(` (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	currency_code TEXT NOT NULL REFERENCES `)
	b.WriteString(currencies)
	b.WriteString(` (code) ON DELETE CASCADE,
	rate NUMERIC(20, 8) NOT NULL CHECK (rate > 0),
	source TEXT NOT NULL DEFAULT '',
	recorded_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS currency_exchange_rate_history_code_recorded_idx
	ON `)
	b.WriteString(history)
	b.WriteString(` (currency_code, recorded_at DESC);

DROP TRIGGER IF EXISTS currencies_set_updated_at ON `)
	b.WriteString(currencies)
	b.WriteString(`;

CREATE TRIGGER currencies_set_updated_at
	BEFORE UPDATE ON `)
	b.WriteString(currencies)
	b.WriteString(`
	FOR EACH ROW EXECUTE FUNCTION `)
	b.WriteString(setUpdatedAt)
	b.WriteString(`();

DELETE FROM `)
	b.WriteString(history)
	b.WriteString(` WHERE currency_code = `)
	b.WriteString(quoteLiteral(cfg.BaseCurrency))
	b.WriteString(`;

DELETE FROM `)
	b.WriteString(rates)
	b.WriteString(` WHERE currency_code = `)
	b.WriteString(quoteLiteral(cfg.BaseCurrency))
	b.WriteString(`;
`)

	if !cfg.SkipSeed {
		b.WriteString(buildSeedSQL(currencies, cfg))
	}

	return b.String()
}

func buildSeedSQL(currencies string, cfg Config) string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("\nINSERT INTO %s (code, name, symbol, is_base, website_languages) VALUES\n", currencies))

	values := make([]string, 0, len(cfg.Currencies))
	for _, code := range cfg.Currencies {
		entry, ok := ISO4217Catalog[code]
		if !ok {
			continue
		}

		values = append(values, fmt.Sprintf(
			"\t(%s, %s, %s, %t, %s)",
			quoteLiteral(code),
			quoteLiteral(entry.Name),
			quoteLiteral(entry.Symbol),
			code == cfg.BaseCurrency,
			pgTextArray(WebsiteLanguagesForCurrency(code)),
		))
	}

	b.WriteString(strings.Join(values, ",\n"))
	b.WriteString("\nON CONFLICT (code) DO UPDATE SET\n")
	b.WriteString("\tname = EXCLUDED.name,\n")
	b.WriteString("\tsymbol = EXCLUDED.symbol,\n")
	b.WriteString("\tis_base = EXCLUDED.is_base,\n")
	b.WriteString("\twebsite_languages = EXCLUDED.website_languages;\n")

	return b.String()
}
