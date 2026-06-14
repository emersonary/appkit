# appkit

Shared Go libraries for backend services. The module provides structured errors, logging, PostgreSQL history/audit generation, currency conversion, and an in-memory heaped cache.

```
github.com/emersonary/appkit
```

## Installation

Module path:

```
github.com/emersonary/appkit
```

### Private repository

appkit is intended to be published as a **private GitHub module**. See [PUBLISHING.md](PUBLISHING.md) for creating the repo, tagging releases, and CI setup.

On each machine that builds consumers (via-jeri, sahar, etc.):

```bash
go env -w GOPRIVATE=github.com/emersonary/appkit
```

In a consumer project:

```bash
go get github.com/emersonary/appkit@v0.1.0
```

```go
// go.mod
require github.com/emersonary/appkit v0.1.0
```

### Local development (before publishing, or while iterating)

Point at a checkout with a `replace` directive:

```go
// go.mod
require github.com/emersonary/appkit v0.0.0

replace github.com/emersonary/appkit => ../appkit
```

## Packages

| Package | Purpose |
|---------|---------|
| `apperror` | Structured application errors with gRPC mapping and zap fields |
| `log` | Zap logger setup and consistent error logging |
| `dbhist` | Generate audit/history SQL and repo functions from live PostgreSQL schemas |
| `currency` | ISO 4217 catalog, config-driven enabled currencies, schema bootstrap, rate sync, conversion — see [currency/README.md](currency/README.md) |
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

`UpdateHist` discovers matching tables, applies audit metadata, creates `%_hist` / `%_hist_detail` structures, and optionally installs repo functions. It does not create search indexes.

---

## apperror

Structured errors for services and APIs:

```go
return currency.ErrUnknownCurrency.With("code", "XYZ")

if appErr, ok := apperror.As(err); ok {
    _ = appErr.ZapFields()
    _ = appErr.GRPCCode()
}

return apperror.ToGRPCStatus(err)
```

Use `apperror.IsValidation(err)` to distinguish client errors from internal failures.

---

## log

Zap logger setup with JSON (default) or text output:

```go
logger, err := log.New(log.Config{
    Level:  "info",
    Format: "json",
})
if err != nil {
    return err
}

log.Log(logger, "operation failed", err)
```

When the error is an `apperror.Error`, `log.Log` emits structured fields automatically.

---

## Development

Run tests from the module root:

```bash
go test ./...
```

Package-specific tests:

```bash
go test ./currency/...
go test ./dbhist/...
go test ./heapedcache/...
```

## Reference integrations

- **via-jeri** — `backend/cmd/db/main.go` uses `dbhist` and `log`
- **sahar** — `internal/db/currency.go` loads `config/currency.yaml` and calls `currency.ApplySchema`; `cmd/server/main.go` wires `currency.NewService`
