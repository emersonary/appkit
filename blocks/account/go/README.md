# accounts (Go)

Config-driven identity layer: schema bootstrap, email/password auth, OAuth, JWT sessions.

## Config (main `config.yaml` `accounts` node)

See `accounts.example.yaml` for the full shape. Secrets (`jwt_secret`, Google `client_secret`) belong in env or local config overrides.

## Apply schema (migration time)

```go
if err := accounts.ApplySchemaFromAppConfig(ctx, sqlDB, appCfg.Accounts); err != nil {
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
