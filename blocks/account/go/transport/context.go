package transport

import (
	"context"
	"strings"

	"github.com/emersonary/appkit/accounts"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type contextKey string

const accountIDKey contextKey = "accountID"

func AccountIDFromContext(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(accountIDKey).(string)
	return id, ok && id != ""
}

// WithAccountID attaches the authenticated account id for downstream handlers.
func WithAccountID(ctx context.Context, accountID string) context.Context {
	return context.WithValue(ctx, accountIDKey, accountID)
}

// ContextWithAuthorizationHeader maps a Connect/HTTP Authorization header into gRPC metadata.
func ContextWithAuthorizationHeader(ctx context.Context, authorization string) context.Context {
	if authorization == "" {
		return ctx
	}
	return metadata.NewIncomingContext(ctx, metadata.Pairs("authorization", authorization))
}

func BearerToken(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", status.Error(codes.Unauthenticated, "missing metadata")
	}
	values := md.Get("authorization")
	if len(values) == 0 {
		return "", status.Error(codes.Unauthenticated, "missing authorization")
	}
	const prefix = "Bearer "
	for _, v := range values {
		if strings.HasPrefix(v, prefix) {
			return strings.TrimSpace(strings.TrimPrefix(v, prefix)), nil
		}
	}
	return "", status.Error(codes.Unauthenticated, "invalid authorization")
}

type SessionResolver interface {
	SessionFromToken(ctx context.Context, accessToken string) (accounts.Session, error)
}

func Authenticator(resolver SessionResolver) func(context.Context) (context.Context, error) {
	return func(ctx context.Context) (context.Context, error) {
		token, err := BearerToken(ctx)
		if err != nil {
			return nil, err
		}
		session, err := resolver.SessionFromToken(ctx, token)
		if err != nil {
			return nil, MapGRPCError(err)
		}
		return WithAccountID(ctx, session.Account.ID), nil
	}
}
