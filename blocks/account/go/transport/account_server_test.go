package transport

import (
	"context"
	"testing"
	"time"

	accountv1 "github.com/emersonary/appkit/accounts/gen/account/v1"
	"github.com/emersonary/appkit/accounts"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestToProtoAccount(t *testing.T) {
	firstName := "Ada"
	lastName := "Lovelace"
	avatarURL := "https://example.com/a.png"
	verifiedAt := time.Now()

	account := accounts.Account{
		ID:              "acct-1",
		Email:           "ada@example.com",
		FirstName:       &firstName,
		LastName:        &lastName,
		AvatarURL:       &avatarURL,
		EmailVerifiedAt: &verifiedAt,
	}

	pb := ToProtoAccount(account)
	if pb.GetId() != account.ID {
		t.Fatalf("id: got %q", pb.GetId())
	}
	if pb.GetEmail() != account.Email {
		t.Fatalf("email: got %q", pb.GetEmail())
	}
	if !pb.GetEmailVerified() {
		t.Fatal("expected email verified true")
	}
}

func TestAccountServer_Login_InvalidArgument(t *testing.T) {
	s := NewAccountServer(mustTestService(t))

	_, err := s.Login(context.Background(), &accountv1.LoginRequest{
		Email:    "not-an-email",
		Password: "secret123",
	})
	assertGRPCCode(t, err, codes.InvalidArgument)
}

func TestBearerToken_FromConnectAuthorizationHeader(t *testing.T) {
	ctx := ContextWithAuthorizationHeader(context.Background(), "Bearer token-abc")
	token, err := BearerToken(ctx)
	if err != nil {
		t.Fatalf("BearerToken: %v", err)
	}
	if token != "token-abc" {
		t.Fatalf("token: got %q", token)
	}
}

func TestAccountServer_GetSession_MissingAuthorization(t *testing.T) {
	s := NewAccountServer(mustTestService(t))

	_, err := s.GetSession(context.Background(), &accountv1.GetSessionRequest{})
	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("expected gRPC status, got %v", err)
	}
	if st.Code() != codes.Unauthenticated {
		t.Fatalf("code: got %v want %v", st.Code(), codes.Unauthenticated)
	}
}

func mustTestService(t *testing.T) *accounts.Service {
	t.Helper()
	svc, err := accounts.New(nil, accounts.Config{
		Schema: "accounts",
		Tenancy: accounts.TenancyConfig{DefaultTenantID: "default"},
	}, accounts.Secrets{JWTSecret: "test-secret"}, accounts.Options{})
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	return svc
}

func assertGRPCCode(t *testing.T, err error, want codes.Code) {
	t.Helper()
	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("expected gRPC status, got %v", err)
	}
	if st.Code() != want {
		t.Fatalf("code: got %v want %v", st.Code(), want)
	}
}
