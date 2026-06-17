package transport

import (
	"context"
	"testing"

	accountv1 "github.com/emersonary/appkit/accounts/gen/account/v1"
	"github.com/emersonary/appkit/accounts"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
)

type stubSessionResolver struct {
	session accounts.Session
	err     error
}

func (s stubSessionResolver) SessionFromToken(_ context.Context, _ string) (accounts.Session, error) {
	if s.err != nil {
		return accounts.Session{}, s.err
	}
	return s.session, nil
}

func TestGRPCUnaryInterceptor_publicAccountMethod(t *testing.T) {
	interceptor := GRPCUnaryInterceptor(mustTestService(t), "/member.v1.MembershipService/")
	called := false
	_, err := interceptor(
		context.Background(),
		nil,
		&grpc.UnaryServerInfo{FullMethod: accountv1.AccountService_Login_FullMethodName},
		func(ctx context.Context, req any) (any, error) {
			called = true
			return "ok", nil
		},
	)
	if err != nil {
		t.Fatalf("Login: %v", err)
	}
	if !called {
		t.Fatal("expected handler to run")
	}
}

func TestGRPCUnaryInterceptor_protectedPrefixRequiresSession(t *testing.T) {
	interceptor := GRPCUnaryInterceptor(mustTestService(t), "/member.v1.MembershipService/")
	_, err := interceptor(
		context.Background(),
		nil,
		&grpc.UnaryServerInfo{FullMethod: "/member.v1.MembershipService/Enroll"},
		func(context.Context, any) (any, error) { return "ok", nil },
	)
	assertGRPCCode(t, err, codes.Unauthenticated)
}

func TestGRPCUnaryInterceptor_protectedPrefixWithSession(t *testing.T) {
	resolver := stubSessionResolver{session: accounts.Session{
		Account: accounts.Account{ID: "acct-1"},
	}}
	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs(
		"authorization", "Bearer token-abc",
	))
	interceptor := GRPCUnaryInterceptor(resolver, "/member.v1.MembershipService/")

	var accountID string
	_, err := interceptor(
		ctx,
		nil,
		&grpc.UnaryServerInfo{FullMethod: "/member.v1.MembershipService/Enroll"},
		func(ctx context.Context, _ any) (any, error) {
			accountID, _ = AccountIDFromContext(ctx)
			return "ok", nil
		},
	)
	if err != nil {
		t.Fatalf("Enroll: %v", err)
	}
	if accountID != "acct-1" {
		t.Fatalf("account id: got %q", accountID)
	}
}
