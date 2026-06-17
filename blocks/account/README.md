# account block

Cross-cutting **accounts** feature: Go service library, React UI, and protobuf contract shipped together.

```
blocks/account/
├── proto/v1/account.proto
├── go/     # module github.com/emersonary/appkit/accounts
│   ├── service.go, store.go, schema.go
│   ├── http/         # OAuth + verify-email REST routes
│   └── transport/    # Connect + gRPC, Mount, account interceptors
└── web/    # npm @emersonary/appkit-accounts
```

See also [`../BLOCK.md`](../BLOCK.md) for the block template used by future blocks.

## Wiring (Go)

```go
svc, err := accounts.New(sqlDB, cfg, secrets, accounts.Options{Mailer: mailer})

mux := http.NewServeMux()
grpcSrv := grpc.NewServer(grpc.UnaryInterceptor(
    accounttransport.GRPCUnaryInterceptor(svc, "/member.v1.MembershipService/"),
))

accounttransport.New(svc, &accounttransport.Mount{
    HTTPMux:    mux,
    GRPCServer: grpcSrv,
})

// App-owned routes (not accounts)
mux.HandleFunc("GET /healthz", handleHealth)
```

`transport.New` configures Google OAuth when `Mount` is non-nil.

### Protecting other services (Connect)

```go
connect.WithInterceptors(
    accounttransport.ConnectRequireSession(svc,
        memberv1connect.MembershipServiceGetMembershipProcedure,
        memberv1connect.MembershipServiceEnrollProcedure,
    ),
)
```

Account id in handlers: `accounttransport.AccountIDFromContext(ctx)`.

## Codegen

From appkit root:

```bash
make proto
```

## Consumers

**Go**:

```go
import (
    "github.com/emersonary/appkit/accounts"
    accounttransport "github.com/emersonary/appkit/accounts/transport"
)
```

Local replace (Sahar / monorepo):

```go
replace github.com/emersonary/appkit/accounts => ../../via-jeri/appkit/blocks/account/go
```

**npm**:

```json
"@emersonary/appkit-accounts": "file:../../via-jeri/appkit/blocks/account/web"
```

See [`web/README.md`](web/README.md) and [`web/BINDING.md`](web/BINDING.md).
