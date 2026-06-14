package currency

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/emersonary/appkit/apperror"
)

func TestISO4217CatalogContainsMajorCurrencies(t *testing.T) {
	for _, code := range []string{"USD", "EUR", "BRL", "JPY", "GBP"} {
		if _, ok := ISO4217Catalog[code]; !ok {
			t.Fatalf("expected ISO4217 catalog to contain %s", code)
		}
	}

	if len(ISO4217Catalog) < 150 {
		t.Fatalf("expected a broad ISO4217 catalog, got %d entries", len(ISO4217Catalog))
	}
}

func TestLoadConfigValidatesAgainstISO4217(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "currency.yaml")
	content := `
schema: public
currencies:
  - USD
  - XXXINVALID
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := LoadConfig(path)
	if err == nil {
		t.Fatal("expected invalid currency code to fail")
	}
}

func TestLoadConfigRequiresBaseCurrencyInList(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "currency.yaml")
	content := `
schema: public
currencies:
  - EUR
  - BRL
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := LoadConfig(path)
	if err != ErrBaseCurrencyMissing {
		t.Fatalf("expected ErrBaseCurrencyMissing, got %v", err)
	}
}

func TestLoadConfigRequiresSchema(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "currency.yaml")
	content := `
currencies:
  - USD
  - EUR
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := LoadConfig(path)
	if err != ErrSchemaRequired {
		t.Fatalf("expected ErrSchemaRequired, got %v", err)
	}
}

func TestValidateCodeUsesEnabledConfig(t *testing.T) {
	svc, err := NewService(nil, Config{
		Schema:       "public",
		BaseCurrency: BaseCurrencyCode,
		Currencies:   []string{BaseCurrencyCode, "EUR"},
	}, Options{})
	if err != nil {
		t.Fatal(err)
	}

	if err := svc.ValidateCode("EUR"); err != nil {
		t.Fatalf("expected EUR to be enabled: %v", err)
	}

	if err := svc.ValidateCode("BRL"); err == nil {
		t.Fatal("expected BRL to be rejected when not in config")
	} else if appErr, ok := apperror.As(err); !ok || appErr.Code != ErrUnknownCurrency.Code {
		t.Fatalf("expected ErrUnknownCurrency, got %v", err)
	}
}
