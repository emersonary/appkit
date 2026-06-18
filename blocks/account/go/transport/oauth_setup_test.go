package transport

import (
	"testing"

	"github.com/emersonary/appkit/accounts"
)

func TestConfigureOAuthProviders_registrationDisabled(t *testing.T) {
	disabled := false
	svc := mustOAuthTestService(t, accounts.Config{
		Features: accounts.FeaturesConfig{RegistrationEnabled: &disabled},
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
	})

	enabled := configureOAuthProviders(svc)
	if enabled.Google {
		t.Fatal("expected oauth providers not configured when registration is disabled")
	}
}

func TestConfigureOAuthProviders_parentDisabled(t *testing.T) {
	svc := mustOAuthTestService(t, accounts.Config{
		OAuth: accounts.OAccountConfig{
			Enabled: boolPtr(false),
			Google: accounts.GoogleConfig{
				Enabled:     true,
				RedirectURL: "http://localhost/account/google/callback",
			},
		},
	}, accounts.Secrets{
		JWTSecret:          "test-secret",
		GoogleClientID:     "client-id",
		GoogleClientSecret: "client-secret",
	})

	enabled := configureOAuthProviders(svc)
	if enabled.Google {
		t.Fatal("expected google disabled when oauth parent is disabled")
	}
}

func TestConfigureOAuthProviders_googleEnabled(t *testing.T) {
	svc := mustOAuthTestService(t, accounts.Config{
		OAuth: accounts.OAccountConfig{
			Google: accounts.GoogleConfig{
				Enabled:     true,
				RedirectURL: "http://localhost/account/google/callback",
			},
			Facebook: accounts.ProviderToggle{Enabled: true},
			Apple:    accounts.ProviderToggle{Enabled: true},
		},
	}, accounts.Secrets{
		JWTSecret:          "test-secret",
		GoogleClientID:     "client-id",
		GoogleClientSecret: "client-secret",
	})

	enabled := configureOAuthProviders(svc)
	if !enabled.Google {
		t.Fatal("expected google enabled")
	}
	if enabled.Facebook {
		t.Fatal("facebook is not implemented yet")
	}
	if enabled.Apple {
		t.Fatal("apple is not implemented yet")
	}
}

func mustOAuthTestService(t *testing.T, cfg accounts.Config, secrets accounts.Secrets) *accounts.Service {
	t.Helper()
	cfg.Schema = "accounts"
	cfg.Tenancy.DefaultTenantID = "default"
	svc, err := accounts.New(nil, cfg, secrets, accounts.Options{})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return svc
}

func boolPtr(v bool) *bool {
	return &v
}
