package tenants

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidateFeedID(t *testing.T) {
	valid := []string{"sahar", "viajeri", "emersonarydev", "hubexpress", "solidia"}
	for _, id := range valid {
		if err := validateFeedID(id); err != nil {
			t.Fatalf("expected valid feed id %q: %v", id, err)
		}
	}

	invalid := []string{"", "via-jeri", "Bad-ID", "1abc", "a_b"}
	for _, id := range invalid {
		if err := validateFeedID(id); err == nil {
			t.Fatalf("expected invalid feed id %q", id)
		}
	}
}

func TestConfig_FixedModeRequiresFeedCatalog(t *testing.T) {
	cfg := Config{Schema: "tenant", Mode: ModeFixed}
	cfg.normalize()
	if err := cfg.Validate(); err != ErrFeedCatalogRequired {
		t.Fatalf("got %v want ErrFeedCatalogRequired", err)
	}
}

func TestConfig_FixedModeValidatesFeedIDs(t *testing.T) {
	cfg := Config{
		Schema: "tenant",
		Mode:   ModeFixed,
		Feed: []FeedEntry{
			{ID: "via-jeri", Name: "Via Jeri"},
		},
	}
	cfg.normalize()
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected validation error for dashed feed id")
	}
}

func TestConfig_HasCatalogID(t *testing.T) {
	cfg := Config{
		Schema: "tenant",
		Mode:   ModeFixed,
		Feed: []FeedEntry{
			{ID: "sahar", Name: "Sahar"},
		},
	}
	if !cfg.HasCatalogID("sahar") {
		t.Fatal("expected sahar in catalog")
	}
	if cfg.HasCatalogID("missing") {
		t.Fatal("did not expect missing in catalog")
	}
}

func TestLoadConfig_FixedExample(t *testing.T) {
	path := filepath.Join("..", "tenants.example.yaml")
	if _, err := os.Stat(path); err != nil {
		t.Skip("example config missing")
	}

	cfg, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	if !cfg.IsFixedMode() || len(cfg.Feed) != 1 {
		t.Fatalf("unexpected config: mode=%s feeds=%d", cfg.Mode, len(cfg.Feed))
	}
}

func TestService_CreateTenantBlockedInFixedMode(t *testing.T) {
	svc := &Service{cfg: Config{Schema: "tenant", Mode: ModeFixed, Feed: []FeedEntry{{ID: "sahar", Name: "Sahar"}}}}
	_, err := svc.CreateTenant(t.Context(), "acct", "Acme", "acme", "UTC")
	if err != ErrFixedMode {
		t.Fatalf("got %v want ErrFixedMode", err)
	}
}
