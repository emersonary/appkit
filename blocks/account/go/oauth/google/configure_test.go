package google

import (
	"strings"
	"testing"

	"github.com/emersonary/appkit/accounts"
	"github.com/emersonary/appkit/accounts/oauth"
)

func TestConfigure_disabledWithoutCredentials(t *testing.T) {
	svc, err := accounts.New(nil, accounts.Config{
		Schema:  "accounts",
		Tenancy: accounts.TenancyConfig{DefaultTenantID: "default"},
		OAuth: accounts.OAccountConfig{
			Google: accounts.GoogleConfig{
				Enabled:     true,
				RedirectURL: "http://localhost/account/google/callback",
			},
		},
	}, accounts.Secrets{JWTSecret: "test-secret"}, accounts.Options{})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	if Configure(svc) {
		t.Fatal("expected Configure to return false without credentials")
	}
	if _, ok := svc.OAccountProvider(oauth.ProviderGoogle); ok {
		t.Fatal("expected no google provider")
	}
}

func TestConfigure_registersGoogle(t *testing.T) {
	svc, err := accounts.New(nil, accounts.Config{
		Schema:  "accounts",
		Tenancy: accounts.TenancyConfig{DefaultTenantID: "default"},
		OAuth: accounts.OAccountConfig{
			Google: accounts.GoogleConfig{
				Enabled:     true,
				RedirectURL: "http://localhost/account/google/callback",
			},
		},
	}, accounts.Secrets{
		JWTSecret:          "test-secret",
		GoogleClientID:     "client-id",
		GoogleClientSecret: "client-secret",
	}, accounts.Options{})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	if !Configure(svc) {
		t.Fatal("expected Configure to return true")
	}
	provider, ok := svc.OAccountProvider(oauth.ProviderGoogle)
	if !ok {
		t.Fatal("expected google provider to be registered")
	}
	if provider.Name() != oauth.ProviderGoogle {
		t.Fatalf("provider name: got %q", provider.Name())
	}
}

func TestFirstNonEmptyCredential(t *testing.T) {
	if got := firstNonEmptyCredential("", "  ", "value"); got != "value" {
		t.Fatalf("got %q", got)
	}
	if got := firstNonEmptyCredential(strings.Repeat(" ", 3)); got != "" {
		t.Fatalf("got %q", got)
	}
}
