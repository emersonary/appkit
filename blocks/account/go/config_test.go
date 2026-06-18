package accounts

import "testing"

func TestConfig_RegistrationEnabled_DefaultTrue(t *testing.T) {
	if !(Config{}).RegistrationEnabled() {
		t.Fatal("expected registration enabled by default")
	}

	disabled := false
	cfg := Config{Features: FeaturesConfig{RegistrationEnabled: &disabled}}
	if cfg.RegistrationEnabled() {
		t.Fatal("expected registration disabled")
	}
}

func TestConfig_RegisterAsAdmin_DefaultFalse(t *testing.T) {
	if (Config{}).RegisterAsAdmin() {
		t.Fatal("expected register_as_admin default false")
	}

	cfg := Config{Features: FeaturesConfig{RegisterAsAdmin: true}}
	if !cfg.RegisterAsAdmin() {
		t.Fatal("expected register_as_admin true")
	}
}

func TestConfig_SkipEmailVerification(t *testing.T) {
	if (Config{}).SkipEmailVerification() {
		t.Fatal("expected default false")
	}

	cfg := Config{Features: FeaturesConfig{SkipEmailVerification: true}}
	if !cfg.SkipEmailVerification() {
		t.Fatal("expected true when feature enabled")
	}
}

func TestAppConfig_ApplyDefaults_RegistrationEnabled(t *testing.T) {
	app := AppConfig{}
	app.ApplyDefaults()
	if app.RegistrationEnabled == nil || !*app.RegistrationEnabled {
		t.Fatal("expected registration_enabled default true in app config")
	}
}

func TestConfig_OAuthEnabled_DefaultTrue(t *testing.T) {
	if !(Config{}).OAuthEnabled() {
		t.Fatal("expected oauth enabled by default")
	}
	disabled := false
	cfg := Config{OAuth: OAccountConfig{Enabled: &disabled}}
	if cfg.OAuthEnabled() {
		t.Fatal("expected oauth disabled")
	}
}

func TestConfig_OAuthProviderEnabled(t *testing.T) {
	cfg := Config{
		OAuth: OAccountConfig{
			Google:   GoogleConfig{Enabled: true},
			Facebook: ProviderToggle{Enabled: false},
		},
	}
	if !cfg.OAuthProviderEnabled("google") {
		t.Fatal("expected google enabled")
	}
	if cfg.OAuthProviderEnabled("facebook") {
		t.Fatal("expected facebook disabled")
	}

	cfg = Config{OAuth: OAccountConfig{Enabled: boolPtr(false), Google: GoogleConfig{Enabled: true}}}
	if cfg.OAuthProviderEnabled("google") {
		t.Fatal("expected providers disabled when oauth parent is disabled")
	}
}

func boolPtr(v bool) *bool {
	return &v
}

func TestAppConfig_BlockConfig(t *testing.T) {
	app := AppConfig{
		Enabled: true,
		Schema:  "account",
		Tenancy: TenancyConfig{Enabled: false, DefaultTenantID: "sahar"},
		FrontendURL:  "http://localhost:5174",
		APIPublicURL: "http://localhost:8080",
	}
	cfg, err := app.BlockConfig()
	if err != nil {
		t.Fatalf("BlockConfig: %v", err)
	}
	if cfg.Schema != "account" {
		t.Fatalf("schema: got %q", cfg.Schema)
	}
	if cfg.Tenancy.DefaultTenantID != "sahar" {
		t.Fatalf("default_tenant_id: got %q", cfg.Tenancy.DefaultTenantID)
	}
}

func TestMergeAppConfig_Features(t *testing.T) {
	disabled := false
	block := Config{Features: FeaturesConfig{RegistrationEnabled: &disabled}}
	enabled := true
	mergeAppConfig(&block, AppConfig{RegistrationEnabled: &enabled})
	if !block.RegistrationEnabled() {
		t.Fatal("expected app config to enable registration")
	}

	block = Config{Features: FeaturesConfig{RegisterAsAdmin: true}}
	mergeAppConfig(&block, AppConfig{RegisterAsAdmin: false})
	if !block.RegisterAsAdmin() {
		t.Fatal("expected block register_as_admin to remain enabled when app is false")
	}

	block = Config{Features: FeaturesConfig{SkipEmailVerification: false}}
	mergeAppConfig(&block, AppConfig{SkipEmailVerification: true})
	if !block.SkipEmailVerification() {
		t.Fatal("expected app config to enable skip email verification")
	}
}
