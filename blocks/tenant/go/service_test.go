package tenants

import (
	"context"
	"testing"
)

func TestValidateSlug(t *testing.T) {
	if err := validateSlug("my-org"); err != nil {
		t.Fatalf("expected valid slug: %v", err)
	}
	if err := validateSlug("Bad Slug"); err == nil {
		t.Fatal("expected invalid slug")
	}
}

func TestService_CreateTenantRequiresAccount(t *testing.T) {
	svc := &Service{cfg: Config{Schema: "tenant"}}
	_, err := svc.CreateTenant(context.Background(), "", "Acme", "acme", "UTC")
	if err != ErrInvalidArgument {
		t.Fatalf("got %v want ErrInvalidArgument", err)
	}
}
