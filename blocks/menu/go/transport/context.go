package transport

import (
	"context"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type contextKey string

const accountIDKey contextKey = "accountID"

func WithAccountID(ctx context.Context, accountID string) context.Context {
	return context.WithValue(ctx, accountIDKey, accountID)
}

func AccountIDFromContext(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(accountIDKey).(string)
	return id, ok && id != ""
}

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

type TokenAccountResolver interface {
	ResolveAccountID(ctx context.Context, accessToken string) (string, error)
}

func Authenticator(resolver TokenAccountResolver) func(context.Context) (context.Context, error) {
	return func(ctx context.Context) (context.Context, error) {
		token, err := BearerToken(ctx)
		if err != nil {
			return nil, err
		}
		accountID, err := resolver.ResolveAccountID(ctx, token)
		if err != nil {
			return nil, MapGRPCError(err)
		}
		return WithAccountID(ctx, accountID), nil
	}
}
