# Block template

Use this when adding a new directory under `blocks/<name>/`. Blocks bundle **proto + Go backend + React UI**. Libraries without UI belong in `go/` instead (e.g. currency).

## Directory structure

```
blocks/<name>/
├── proto/v1/<name>.proto
├── go/
│   ├── service.go          # domain API
│   ├── store.go            # persistence (private to block)
│   ├── schema.go           # ApplySchema — block tables only
│   ├── config.go           # YAML + validation
│   ├── transport/
│   │   ├── block.go        # New(svc, *Mount) — register routes
│   │   ├── auth.go         # cross-service account session helpers (if block owns sessions)
│   │   └── ...
│   └── http/               # optional REST owned by block
├── web/
│   ├── src/index.ts
│   └── package.json
└── README.md
```

## Backend wiring (consumer app)

Order in the app composition root:

1. **Migrations** — app goose SQL for app tables; call block `ApplySchema` for block tables.
2. **createServices** — `block.New(db, cfg, secrets, opts)`.
3. **createConnectionHandlers** — `transport.New(svc, &transport.Mount{HTTPMux, GRPCServer})`.

### Mount

`Mount` registers all transport the block owns. App keeps health checks, CORS, and app-specific routes outside the block.

### Cross-service account sessionentication

When other services (membership, blog, …) require a logged-in account, **do not** reimplement session parsing in the host app. Use the account block helpers:

```go
// gRPC server shared by account + app services
grpc.NewServer(grpc.UnaryInterceptor(
    accounttransport.GRPCUnaryInterceptor(accountsSvc,
        "/member.v1.MembershipService/",
        "/other.v1.OtherService/", // add prefixes for protected app RPCs
    ),
))

// Connect handler for an app service
connect.WithInterceptors(
    accounttransport.ConnectRequireSession(accountsSvc,
        memberv1connect.MembershipServiceGetMembershipProcedure,
        memberv1connect.MembershipServiceEnrollProcedure,
    ),
)
```

Authenticated account id is on the context: `transport.AccountIDFromContext(ctx)`.

Account RPC public/bearer rules are **built into** `GRPCUnaryInterceptor` — consumers must not maintain method allowlists.

### Schema rules

- `ApplySchema` creates **only** block-owned tables in the block YAML schema.
- **Never** migrate host-app tables or rebind host FKs inside the block.
- Host apps that replace legacy tables add a **goose migration** in the app repo (see Sahar `000026_migrate_legacy_public_to_account_to_block.sql`).

### Required adapters

Document what the host must supply (mailer, secrets, logger). Blocks should not import app packages.

## Frontend wiring (consumer app)

Document explicitly — not zero-config today:

1. npm `file:` or published package dependency
2. Vite alias + `server.fs.allow`
3. Dev proxy for Connect paths and block REST prefixes
4. App routes (login, callback, …)
5. Glue context (membership, i18n) wrapping block `Provider`

Prefer exporting route path constants from the web package where possible.

## Deploy checklist (per block README)

- Env vars and secrets table
- nginx location snippets for REST + Connect paths
- OAuth redirect URIs (if applicable)
- Example YAML config

## Proto / codegen

- Source of truth: `blocks/<name>/proto/`
- Go gen: `blocks/<name>/go/gen/`
- TS gen: `blocks/<name>/web/src/gen/`
- Consumers either import block gen directly or run a sync script — document one approach and CI-check stub drift.

## Versioning

When publishing, tag Go module and npm package together. Document minimum compatible appkit version in block README.

## Reference implementation

See [`account/`](account/) for the first complete block.
