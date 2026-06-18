# tenant block

Organization and membership for multi-tenant apps. Tenants does **not** import the accounts block; the host wires auth via `ResolveAccountID`.

```
blocks/tenant/
├── proto/v1/tenant.proto
├── go/     # module github.com/emersonary/appkit/tenants
└── web/    # npm @emersonary/appkit-tenants
```

## Go wiring

```yaml
# config.yaml
tenants:
  enabled: true
  config_path: config/tenants.yaml
  jwt_secret: "" # optional; prefer wiring ResolveAccountID from accounts
```

```go
svc, err := tenants.Wire(ctx, sqlDB, appCfg.Tenants)

tenanttransport.New(svc, &tenanttransport.Mount{
    HTTPMux: mux,
    GRPCServer: grpcSrv,
    ResolveAccountID: func(ctx context.Context, token string) (string, error) {
        session, err := accountsSvc.SessionFromToken(ctx, token)
        return session.Account.ID, err
    },
})
```

After `tenants.CreateTenant`, re-issue the session with the active tenant:

```go
session, _ := accountsSvc.IssueSession(ctx, accountID, tenantID)
```

## RPCs

- `CreateTenant` — caller becomes `owner`
- `ListMyTenants`
- `GetTenant`
- `InviteMember` — owner/admin only
- `AcceptInvite`

## Schema (`tenant` schema by default)

- `tenants` — slug, name, timezone, status
- `tenant_accounts` — membership + role
- `tenant_invites` — email invites

## Fixed mode (`mode: fixed`)

Declare a catalog of feeds in `tenants.yaml`; runtime upserts them on wire and rejects `CreateTenant`.

```yaml
schema: tenant
mode: fixed
feed:
  - id: sahar
    name: Sahar
    timezone: America/New_York
    metadata:
      languages: [pt, en]
      default_language: pt
      public_url: https://example.com
      callback_url: https://admin.example.com
```

Feed `id` values must be lowercase alphanumeric (no dashes). Product apps may read `metadata` for domain-specific fields.

## Web

```tsx
<TenantsProvider tenantClient={tenantClient} onActiveTenantChange={setTenantId}>
  <CreateTenantForm />
  <TenantSwitcher />
</TenantsProvider>
```

Generate TS stubs: `npm run generate:tenant-proto` in `blocks/tenant/web`.
