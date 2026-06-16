# account block

Cross-cutting **accounts** feature: Go service library, React UI, and protobuf contract shipped together.

```
blocks/account/
├── proto/v1/account.proto
├── go/     # module github.com/emersonary/appkit/accounts
│   ├── service.go, store.go, schema.go
│   ├── http/         # OAuth + verify-email REST routes
│   └── transport/    # Connect + gRPC AccountService, Block mount API
└── web/    # npm @emersonary/appkit-accounts
```

## Wiring (Go)

```go
svc, err := accounts.NewService(sqlDB, cfg, secrets, accounts.Options{Mailer: mailer})
block := transport.NewBlock(svc)

// HTTP: Connect RPC + default REST + your routes
block.MountHTTP(mux, transport.HTTPMount{
    OAuthProvider: googleProvider,
    RegisterExtra: func(mux *http.ServeMux) {
        mux.HandleFunc("GET /healthz", handleHealth)
        // REST or WebSocket handlers
    },
    ExtraRoutes: []transport.HTTPRoute{
        {Pattern: "/ws", Handler: wsHandler},
    },
})

// gRPC
block.RegisterGRPC(grpcServer)
```

`Service` methods (`Login`, `Register`, `SessionFromToken`, …) stay public for direct use in tests, jobs, and app-specific HTTP.

Proto contract: [`proto/v1/account.proto`](proto/v1/account.proto).

## Codegen

From appkit root:

```bash
make proto
```

## Consumers

**Go** — import paths unchanged:

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
