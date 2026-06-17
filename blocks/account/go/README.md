# accounts (Go)

Config-driven identity layer: schema bootstrap, email/password auth, OAuth, JWT sessions.

## Config (`accounts.yaml`)

```yaml
schema: account

tenancy:
  enabled: false
  default_tenant_id: sahar

session:
  access_token_ttl: 24h

urls:
  frontend_url: http://localhost:5174
  api_public_url: http://localhost:8082

oauth:
  state_cookie_name: sahar_oauth_state
  google:
    enabled: true
    redirect_url: http://localhost:8082/account/google/callback

features:
  admin_flag: true
```

Secrets (`JWTSecret`, Google credentials) are passed at runtime via `accounts.Secrets`, not YAML.

## Apply schema (migration time)

```go
cfg, err := accounts.LoadConfig("config/accounts.yaml")
if err != nil { return err }

if err := accounts.ApplySchema(ctx, sqlDB, cfg); err != nil {
    return err
}
```

`ApplySchema` is idempotent and creates **only** block-owned tables in the configured schema. Host-app legacy migrations stay in the consumer repo.

## Service

```go
svc, err := accounts.New(sqlDB, cfg, accounts.Secrets{
    JWTSecret:          os.Getenv("JWT_SECRET"),
    GoogleClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
    GoogleClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
}, accounts.Options{
    Mailer: myMailer,
})
```

## Transport mount

```go
accounttransport.New(svc, &accounttransport.Mount{
    HTTPMux:    mux,
    GRPCServer: grpcSrv,
})
```

Registers account REST (`/account/*`), Connect, and gRPC. OAuth is configured automatically when mount is non-nil.

## Cross-service account session

```go
// gRPC: built-in account method rules + optional protected prefixes
accounttransport.GRPCUnaryInterceptor(svc, "/member.v1.MembershipService/")

// Connect: require session for listed procedures
accounttransport.ConnectRequireSession(svc, procedureNames...)
```

## HTTP routes (also registered via Mount)

`GET /account/verify-email`, `GET /account/google`, `GET /account/google/callback`
