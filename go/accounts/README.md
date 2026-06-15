# accounts (Go)

Config-driven identity layer: schema bootstrap, email/password auth, OAuth, JWT sessions.

## Config (`accounts.yaml`)

```yaml
schema: auth

tenancy:
  enabled: false          # true creates tenants + account_tenants tables
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
    redirect_url: http://localhost:8082/auth/google/callback

features:
  admin_flag: true
```

## Apply schema (migration time)

```go
cfg, err := accounts.LoadConfig("config/accounts.yaml")
if err != nil { return err }

if err := accounts.ApplySchema(ctx, sqlDB, cfg); err != nil {
    return err
}
```

`ApplySchema` is idempotent. It also migrates legacy `public.users` → `auth.accounts` when present.

## Service

```go
svc, err := accounts.NewService(sqlDB, cfg, accounts.Secrets{
    JWTSecret:          os.Getenv("JWT_SECRET"),
    GoogleClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
    GoogleClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
}, accounts.Options{
    Mailer: myMailer,
})

svc.RegisterProvider(googleoauth.New(googleoauth.Config{...}))
```

## HTTP routes

```go
accounthttp.New(svc, googleProvider).RegisterRoutes(mux)
```

Registers: `GET /auth/google`, `GET /auth/google/callback`, `GET /auth/verify-email`.
