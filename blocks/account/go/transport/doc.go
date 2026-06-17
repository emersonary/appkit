// Package transport wires account HTTP, Connect, and gRPC around accounts.Service.
//
// # Block pattern (use in app composition roots)
//
//	svc, err := accounts.New(db, cfg, secrets, opts)     // createServices
//	transport.New(svc, &transport.Mount{                  // createConnectionHandlers
//	    HTTPMux:    mux,
//	    GRPCServer: grpcSrv,
//	})
//
// # Cross-service account session
//
// Use GRPCUnaryInterceptor and ConnectRequireSession so host apps do not duplicate
// session parsing or account RPC allowlists:
//
//	grpc.NewServer(grpc.UnaryInterceptor(
//	    transport.GRPCUnaryInterceptor(svc, "/member.v1.MembershipService/"),
//	))
//	connect.WithInterceptors(transport.ConnectRequireSession(svc, procedures...))
//
// Authenticated id: AccountIDFromContext(ctx).
package transport
