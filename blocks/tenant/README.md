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

## Web

```tsx
<TenantsProvider tenantClient={tenantClient} onActiveTenantChange={setTenantId}>
  <CreateTenantForm />
  <TenantSwitcher />
</TenantsProvider>
```

Generate TS stubs: `npm run generate:tenant-proto` in `blocks/tenant/web`.
