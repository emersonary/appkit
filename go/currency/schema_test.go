package currency

import (
	"strings"
	"testing"
)

func TestBuildSchemaSQLIncludesConfiguredSeed(t *testing.T) {
	cfg := Config{
		Schema:       "public",
		BaseCurrency: BaseCurrencyCode,
		Currencies:   []string{BaseCurrencyCode, "EUR", "BRL"},
	}
	cfg.normalize()

	sqlText := buildSchemaSQL(cfg)

	for _, fragment := range []string{
		`CREATE TABLE IF NOT EXISTS "public"."currencies"`,
		`CREATE TABLE IF NOT EXISTS "public"."currency_exchange_rates"`,
		`DELETE FROM "public"."currency_exchange_rates" WHERE id_currency = 'USD'`,
		`'EUR'`,
		`'BRL'`,
		`ARRAY['pt']`,
		`ON CONFLICT (id_currency) DO UPDATE SET`,
	} {
		if !strings.Contains(sqlText, fragment) {
			t.Fatalf("expected schema SQL to contain %q", fragment)
		}
	}
}

func TestBuildSchemaSQLCustomSchema(t *testing.T) {
	cfg := Config{
		Schema:       "billing",
		BaseCurrency: BaseCurrencyCode,
		Currencies:   []string{BaseCurrencyCode, "EUR"},
		SkipSeed:     true,
	}
	cfg.normalize()

	sqlText := buildSchemaSQL(cfg)

	if !strings.Contains(sqlText, `"billing"."currencies"`) {
		t.Fatalf("expected qualified billing schema, got: %s", sqlText)
	}
}

func TestApplySchemaRejectsInvalidSchema(t *testing.T) {
	cfg := Config{
		Schema:       "bad-name",
		BaseCurrency: BaseCurrencyCode,
		Currencies:   []string{BaseCurrencyCode, "EUR"},
	}

	err := ApplySchema(t.Context(), nil, cfg)
	if err == nil {
		t.Fatal("expected validation error")
	}
}
