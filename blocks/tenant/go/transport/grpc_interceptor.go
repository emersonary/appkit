package transport

import (
	"context"
	"strings"

	"github.com/emersonary/appkit/tenants"
	"google.golang.org/grpc"
)

const tenantServicePrefix = "/tenant.v1.TenantService/"

// GRPCUnaryInterceptor authenticates tenant gRPC methods and attaches account id to context.
func GRPCUnaryInterceptor(resolver AccountIDResolver) grpc.UnaryServerInterceptor {
	authenticate := Authenticator(resolver)

	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		if !strings.HasPrefix(info.FullMethod, tenantServicePrefix) {
			return handler(ctx, req)
		}
		if resolver == nil {
			return nil, MapGRPCError(tenants.ErrUnauthenticated)
		}

		nextCtx, err := authenticate(ctx)
		if err != nil {
			return nil, err
		}
		return handler(nextCtx, req)
	}
}
