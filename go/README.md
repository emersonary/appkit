# appkit (Go)

Shared Go libraries for backend services.

```
github.com/emersonary/appkit
```

Module root is this directory (`go/`). Import paths are unchanged, e.g. `github.com/emersonary/appkit/currency`.

## Installation

See [PUBLISHING.md](PUBLISHING.md) for GitHub setup, tags, and CI.

```bash
go env -w GOPRIVATE=github.com/emersonary/appkit
go get github.com/emersonary/appkit@v0.1.0
```

### Local development

```go
require github.com/emersonary/appkit v0.0.0
replace github.com/emersonary/appkit => ../appkit/go
```

Adjust the `replace` path relative to your consumer module.

## Packages

| Package | Purpose |
|---------|---------|
| `apperror` | Structured application errors with gRPC mapping and zap fields |
| `log` | Zap logger setup and consistent error logging |
| `migrate` | SQL instruction runner + audit/history/repo functions (`UpdateHist`) from live PostgreSQL schemas |
| `currency` | ISO 4217 catalog, config-driven enabled currencies, schema bootstrap, rate sync, conversion — see [currency/README.md](currency/README.md) |
| `accounts` | Multi-tenant auth (accounts schema, OAuth, JWT sessions) — see [accounts/accounts.example.yaml](accounts/accounts.example.yaml) |
| `heapedcache` | Fixed-size thread-safe in-memory cache with LRU-style eviction — see [heapedcache/README.md](heapedcache/README.md) |
| `resource` | Reflect struct `resource` tags into UI schema + value maps for schema-driven edit forms — see [resource/README.md](resource/README.md) |

---

## migrate

Owns schema lifecycle in one package:

- **Instructions** — up-only SQL registered in code (`Runner`, `platform.schema_instructions`)
- **Audit / repo** — `UpdateHist` generates history triggers and versioned JSON repo functions from table comments (`AUDIT=true` / `REPO=true`)

YAML block on the app config:

```yaml
dbhist:
  enabled: true
```

Usage:

```go
runner, db, err := migrate.OpenRunner(ctx, dsn, migrate.ApplyConfig{Instructions: instr})
if err != nil {
    return err
}
if err := runner.Apply(ctx); err != nil {
    return err
}
if _, err := migrate.Wire(ctx, db, migrate.AppConfig{Enabled: true}, migrate.WireOptions{Logger: logger}); err != nil {
    return err
}
fns, err := migrate.MustFunctions("tenant", "tenant_locations")
```

The former `dbhist` import path remains as a thin compatibility shim over `migrate`.

---

## apperror / log

See package docs and tests under `apperror/` and `log/`.

---

## Development

From this directory (`go/`):

```bash
go test ./...
go test ./currency/...
```

## Reference integrations

- **via-jeri** — `backend/cmd/db/main.go` uses `migrate` (audit/repo) and `log`
- **sahar** — `internal/db/currency.go`, `cmd/server/main.go`
