package transport

import (
	"context"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type contextKey string

const (
	accountIDKey       contextKey = "accountID"
	messengerUserIDKey contextKey = "messengerUserID"
	agentAuthKey       contextKey = "agentAuth"
)

func AccountIDFromContext(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(accountIDKey).(string)
	return id, ok && id != ""
}

func WithAccountID(ctx context.Context, accountID string) context.Context {
	return context.WithValue(ctx, accountIDKey, accountID)
}

func MessengerUserIDFromContext(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(messengerUserIDKey).(string)
	return id, ok && id != ""
}

func WithMessengerUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, messengerUserIDKey, userID)
}

func AgentAuthFromContext(ctx context.Context) bool {
	v, ok := ctx.Value(agentAuthKey).(bool)
	return ok && v
}

func WithAgentAuth(ctx context.Context) context.Context {
	return context.WithValue(ctx, agentAuthKey, true)
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

type AccountIDResolver func(ctx context.Context, accessToken string) (accountID string, err error)

func Authenticator(resolver AccountIDResolver) func(context.Context) (context.Context, error) {
	return func(ctx context.Context) (context.Context, error) {
		token, err := BearerToken(ctx)
		if err != nil {
			return nil, err
		}
		accountID, err := resolver(ctx, token)
		if err != nil {
			return nil, MapGRPCError(err)
		}
		if accountID == "" {
			return nil, status.Error(codes.Unauthenticated, "missing account id")
		}
		return WithAccountID(ctx, accountID), nil
	}
}
