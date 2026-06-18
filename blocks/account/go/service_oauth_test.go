package accounts

import (
	"context"
	"errors"
	"testing"

	"github.com/emersonary/appkit/apperror"
)

func TestLoginOAuth_registrationDisabled(t *testing.T) {
	disabled := false
	svc, err := New(nil, Config{
		Schema:  "accounts",
		Tenancy: TenancyConfig{DefaultTenantID: "default"},
		Features: FeaturesConfig{
			RegistrationEnabled: &disabled,
		},
	}, Secrets{JWTSecret: "test-secret"}, Options{})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	_, err = svc.LoginOAuth(context.Background(), "google", "user-1", "user@example.com", nil, nil, nil)
	if !errors.Is(err, ErrRegistrationDisabled) {
		t.Fatalf("got %v want ErrRegistrationDisabled", err)
	}

	var appErr apperror.Error
	if !errors.As(err, &appErr) || appErr.Code != ErrRegistrationDisabled.Code {
		t.Fatalf("unexpected error type: %v", err)
	}
}
