package transport

import (
	"context"
	"strings"

	"connectrpc.com/connect"
	accountv1 "github.com/emersonary/appkit/accounts/gen/account/v1"
	"github.com/emersonary/appkit/accounts"
	"google.golang.org/grpc"
)

var accountPublicGRPCMethods = map[string]bool{
	accountv1.AccountService_Login_FullMethodName:                   true,
	accountv1.AccountService_AdminLogin_FullMethodName:              true,
	accountv1.AccountService_Register_FullMethodName:                true,
	accountv1.AccountService_GetVerificationStatus_FullMethodName:   true,
	accountv1.AccountService_ResendVerificationEmail_FullMethodName: true,
	accountv1.AccountService_RequestPasswordReset_FullMethodName:    true,
	accountv1.AccountService_ResetPassword_FullMethodName:           true,
}

var accountBearerGRPCMethods = map[string]bool{
	accountv1.AccountService_GetSession_FullMethodName: true,
	accountv1.AccountService_Logout_FullMethodName:     true,
}

// GRPCUnaryInterceptor applies account-service session rules and authenticates RPCs whose
// full method name starts with one of protectedPrefixes (e.g. "/member.v1.MembershipService/").
// Authenticated account id is attached via WithAccountID for downstream handlers.
func GRPCUnaryInterceptor(resolver SessionResolver, protectedPrefixes ...string) grpc.UnaryServerInterceptor {
	authenticate := Authenticator(resolver)

	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		method := info.FullMethod

		if accountPublicGRPCMethods[method] {
			return handler(ctx, req)
		}

		if accountBearerGRPCMethods[method] {
			if _, err := BearerToken(ctx); err != nil {
				return nil, err
			}
			return handler(ctx, req)
		}

		for _, prefix := range protectedPrefixes {
			if strings.HasPrefix(method, prefix) {
				next, err := authenticate(ctx)
				if err != nil {
					return nil, err
				}
				return handler(next, req)
			}
		}

		return handler(ctx, req)
	}
}

// ConnectRequireSession requires a valid account session for the given Connect procedures.
// Authorization is read from the Connect request header; account id is on the context for handlers.
func ConnectRequireSession(resolver SessionResolver, procedures ...string) connect.UnaryInterceptorFunc {
	required := make(map[string]struct{}, len(procedures))
	for _, procedure := range procedures {
		required[procedure] = struct{}{}
	}

	authenticate := Authenticator(resolver)

	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			if _, ok := required[req.Spec().Procedure]; !ok {
				return next(ctx, req)
			}

			ctx = ContextWithAuthorizationHeader(ctx, req.Header().Get("Authorization"))
			nextCtx, err := authenticate(ctx)
			if err != nil {
				return nil, ToConnectError(err)
			}
			return next(nextCtx, req)
		}
	}
}

// AdminAuthenticator validates bearer tokens and requires is_admin on the account.
func AdminAuthenticator(resolver SessionResolver) func(context.Context) (context.Context, error) {
	return func(ctx context.Context) (context.Context, error) {
		token, err := BearerToken(ctx)
		if err != nil {
			return nil, err
		}
		session, err := resolver.SessionFromToken(ctx, token)
		if err != nil {
			return nil, MapGRPCError(err)
		}
		if !session.Account.IsAdmin {
			return nil, MapGRPCError(accounts.ErrUnauthenticated)
		}
		return WithAccountID(ctx, session.Account.ID), nil
	}
}

// GRPCUnaryAdminInterceptor authenticates RPCs whose full method name starts with one of
// adminPrefixes and requires the account to have is_admin set.
func GRPCUnaryAdminInterceptor(resolver SessionResolver, adminPrefixes ...string) grpc.UnaryServerInterceptor {
	authenticate := AdminAuthenticator(resolver)

	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		for _, prefix := range adminPrefixes {
			if strings.HasPrefix(info.FullMethod, prefix) {
				next, err := authenticate(ctx)
				if err != nil {
					return nil, err
				}
				return handler(next, req)
			}
		}
		return handler(ctx, req)
	}
}

// ConnectRequireAdmin requires a valid admin session for the given Connect procedures.
func ConnectRequireAdmin(resolver SessionResolver, procedures ...string) connect.UnaryInterceptorFunc {
	required := make(map[string]struct{}, len(procedures))
	for _, procedure := range procedures {
		required[procedure] = struct{}{}
	}

	authenticate := AdminAuthenticator(resolver)

	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			if _, ok := required[req.Spec().Procedure]; !ok {
				return next(ctx, req)
			}

			ctx = ContextWithAuthorizationHeader(ctx, req.Header().Get("Authorization"))
			nextCtx, err := authenticate(ctx)
			if err != nil {
				return nil, ToConnectError(err)
			}
			return next(nextCtx, req)
		}
	}
}
