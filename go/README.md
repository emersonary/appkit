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
| `dbhist` | Generate audit/history SQL and repo functions from live PostgreSQL schemas |
| `currency` | ISO 4217 catalog, config-driven enabled currencies, schema bootstrap, rate sync, conversion — see [currency/README.md](currency/README.md) |
| `accounts` | Multi-tenant auth (accounts schema, OAuth, JWT sessions) — see [accounts/accounts.example.yaml](accounts/accounts.example.yaml) |
| `heapedcache` | Fixed-size thread-safe in-memory cache with LRU-style eviction — see [heapedcache/README.md](heapedcache/README.md) |

---

## dbhist

Generates PostgreSQL audit columns, history tables/triggers, and JSON repo functions from existing tables. Intended to run after normal schema migrations.

Config file example (`dbhist.yaml`):

```yaml
schemas:
  - core
  - catalog

table_pattern: tbl_%

exclude_patterns:
  - tmp_%

modules:
  audit: true
  history: true
  repo_functions: true
```

Usage:

```go
cfg, err := dbhist.LoadConfig("dbhist.yaml")
if err != nil {
    return err
}

result, err := dbhist.UpdateHist(ctx, db, cfg, dbhist.Options{
    Logger: logger,
})
```

Override the config path with `DBHIST_CONFIG`.

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

- **via-jeri** — `backend/cmd/db/main.go` uses `dbhist` and `log`
- **sahar** — `internal/db/currency.go`, `cmd/server/main.go`
