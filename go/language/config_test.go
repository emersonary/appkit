package language

import (
	"os"
	"strings"
	"testing"
)

func TestLanguageCatalogContainsPortalLocales(t *testing.T) {
	for _, code := range []string{"en", "de", "fr", "pt", "ja", "ar", "he"} {
		if _, ok := LanguageCatalog[code]; !ok {
			t.Fatalf("expected catalog to contain %s", code)
		}
	}
}

func TestLanguageConfigRequiresDefaultInList(t *testing.T) {
	cfg := LanguageConfig{
		Schema:          "language",
		DefaultLanguage: "en",
		Languages:       []string{"pt"},
	}
	cfg.normalize()
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected default language validation error")
	}
}

func TestLanguageConfig_Resolved(t *testing.T) {
	cfg, err := (LanguageConfig{
		Schema:          "language",
		DefaultLanguage: "en",
		Languages:       []string{"en", "pt", "fr"},
	}).Resolved()
	if err != nil {
		t.Fatalf("Resolved: %v", err)
	}
	if cfg.DefaultLanguage != "en" {
		t.Fatalf("default: got %q", cfg.DefaultLanguage)
	}
}

func TestBuildSchemaSQLIncludesSeed(t *testing.T) {
	cfg := LanguageConfig{
		Schema:          "language",
		DefaultLanguage: "en",
		Languages:       []string{"en", "pt"},
	}
	cfg.normalize()

	sqlText := buildSchemaSQL(cfg)
	for _, fragment := range []string{
		`CREATE TABLE IF NOT EXISTS "language"."languages"`,
		`'en'`,
		`'pt'`,
		`id_language`,
		`ON CONFLICT (id_language) DO UPDATE SET`,
	} {
		if !strings.Contains(sqlText, fragment) {
			t.Fatalf("expected schema SQL to contain %q", fragment)
		}
	}
}

func TestValidateCodeUsesEnabledConfig(t *testing.T) {
	svc, err := NewService(nil, LanguageConfig{
		Schema:          "language",
		DefaultLanguage: "en",
		Languages:       []string{"en", "pt"},
	}, Options{})
	if err != nil {
		t.Fatal(err)
	}

	if err := svc.ValidateCode("pt"); err != nil {
		t.Fatalf("expected pt enabled: %v", err)
	}
	if err := svc.ValidateCode("fr"); err == nil {
		t.Fatal("expected fr disabled")
	}
}

func TestNewServiceRejectsDefaultNotInLanguages(t *testing.T) {
	_, err := NewService(nil, LanguageConfig{
		Schema:          "language",
		DefaultLanguage: "en",
		Languages:       []string{"pt"},
	}, Options{})
	if err == nil {
		t.Fatal("expected error when default language is not in languages list")
	}
}

func TestResolveBlockConfigValidatesAfterMerge(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/language.yaml"
	if err := os.WriteFile(path, []byte(`schema: language
default_language: pt
languages:
  - pt
`), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := ResolveBlockConfig(LanguageConfig{
		Enabled:         true,
		ConfigPath:      path,
		DefaultLanguage: "en",
	})
	if err == nil {
		t.Fatal("expected validation error when merge overrides default outside languages list")
	}
}

func TestWire_Disabled(t *testing.T) {
	svc, err := Wire(t.Context(), nil, LanguageConfig{Enabled: false}, WireOptions{})
	if err != nil {
		t.Fatalf("Wire: %v", err)
	}
	if svc != nil {
		t.Fatal("expected nil service when disabled")
	}
}
