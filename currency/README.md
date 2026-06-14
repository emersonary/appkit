# currency

PostgreSQL-backed currency conversion for Go applications. The package ships a full ISO 4217 catalog, a YAML config for the currencies your project supports, schema bootstrap, exchange-rate sync, and amount conversion through USD.

```
github.com/emersonary/appkit/currency
```

## What to run at startup

There are two moments to wire up: **once when the database is prepared**, and **every time your application process starts**.

### One-time setup (before first run)

1. Add the config file to your project:

```bash
cp currency.example.yaml config/currency.yaml
```

2. Edit `config/currency.yaml` — set `schema` and list your `currencies`.

### At migration time (deploy / `goose up`)

After your normal SQL migrations, call **`ApplySchema`**. This creates the PostgreSQL schema (if missing), creates the currency tables, and seeds currency rows from the config.

```go
cfg, err := currency.LoadConfig("config/currency.yaml")
if err != nil {
    return err
}

if err := currency.ApplySchema(ctx, db, cfg); err != nil {
    return err
}
```

`ApplySchema` is idempotent — safe to run on every deploy. Most projects call it from the migration step, not from every HTTP request.

You need a `*database/sql.DB` connected to PostgreSQL. With pgx:

```go
sqlDB := stdlib.OpenDBFromPool(pool)
defer sqlDB.Close()

_ = currency.ApplySchema(ctx, sqlDB, cfg)
```

### Every application start (server `main`)

When your process boots, you do **not** need to call `ApplySchema` again if migrations already ran — unless you prefer to keep schema apply in startup instead of migrations.

On every start, run these three steps:

| Step | Function | Purpose |
|------|----------|---------|
| 1 | `currency.LoadConfig(path)` | Load and validate `currency.yaml` |
| 2 | `currency.NewService(db, cfg, opts)` | Build the service (validates config again) |
| 3 | `go svc.RunExchangeRateUpdater(ctx, interval)` | Fetch rates now, then on a timer |

```go
cfg, err := currency.LoadConfig("config/currency.yaml")
if err != nil {
    return err
}

sqlDB := stdlib.OpenDBFromPool(pool) // reuse for the life of the process

svc, err := currency.NewService(sqlDB, cfg, currency.Options{
    Logger: logger,
})
if err != nil {
    return err
}

go svc.RunExchangeRateUpdater(ctx, time.Hour)
```

Keep the `*currency.Service` in your app (inject it into handlers or domain services). Use it for the lifetime of the process:

```go
amount, err := svc.Convert(ctx, 100, "EUR", "BRL")
if err := svc.ValidateCode(req.Currency); err != nil { ... }
```

`RunExchangeRateUpdater` calls `SyncExchangeRates` immediately on start, then repeats every `interval`. Without it, conversions fail with `CURRENCY_RATE_NOT_FOUND` until something else calls `SyncExchangeRates`.

### Startup checklist

```
[ ] config/currency.yaml exists (schema + currencies listed)
[ ] ApplySchema ran during migrations (tables and seed data exist)
[ ] LoadConfig on app start
[ ] NewService on app start
[ ] RunExchangeRateUpdater in a background goroutine (recommended)
[ ] Pass *currency.Service into code that converts or validates currencies
```

### Full example (`main`)

```go
func main() {
    ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
    defer cancel()

    pool := openPostgres() // *pgxpool.Pool
    defer pool.Close()

    // Migration step (often a separate command — shown here for clarity):
    if err := runMigrations(ctx, pool); err != nil {
        log.Fatal(err)
    }

    cfg, err := currency.LoadConfig("config/currency.yaml")
    if err != nil {
        log.Fatal(err)
    }

    sqlDB := stdlib.OpenDBFromPool(pool)

    svc, err := currency.NewService(sqlDB, cfg, currency.Options{Logger: logger})
    if err != nil {
        log.Fatal(err)
    }

    go svc.RunExchangeRateUpdater(ctx, time.Hour)

    startHTTPServer(svc) // handlers call svc.Convert / svc.ValidateCode
    <-ctx.Done()
}

func runMigrations(ctx context.Context, pool *pgxpool.Pool) error {
    sqlDB := stdlib.OpenDBFromPool(pool)
    defer sqlDB.Close()

    if err := goose.UpContext(ctx, sqlDB, "migrations"); err != nil {
        return err
    }

    cfg, err := currency.LoadConfig("config/currency.yaml")
    if err != nil {
        return err
    }

    return currency.ApplySchema(ctx, sqlDB, cfg)
}
```

Override the config path with the `CURRENCY_CONFIG` environment variable if needed.

## When to use it

Use this package when your app needs to:

- store exchange rates in PostgreSQL
- convert amounts between currencies
- restrict API operations to a fixed, project-specific currency list
- seed currency metadata (name, symbol, website languages) from config

## Configuration file

The YAML file drives both the database schema and runtime validation.

```yaml
# PostgreSQL schema where currency tables are created.
# Created automatically if it does not exist.
schema: public

# Exchange feed is USD-based; this must stay USD.
base_currency: USD

# Only these currencies are seeded, synced, and accepted at runtime.
currencies:
  - USD
  - EUR
  - BRL

# Set to true to skip INSERT/UPDATE seed data on ApplySchema.
skip_seed: false
```

### Validation rules

- `schema` is required and must match `^[a-z][a-z0-9_]*$`
- `base_currency` must be `USD`
- `currencies` must be non-empty, contain no duplicates, and include `USD`
- every listed code must exist in the built-in ISO 4217 catalog

See `currency.example.yaml` for a starting template.

### Website languages (built-in)

Locale tags per currency are hardcoded in `website_languages.go` (not in YAML). The map includes every ISO 4217 code from `ISO4217Catalog`; non-country codes such as `XAU`, `XTS`, and `XXX` use an empty list.

When seeding the database, `ApplySchema` looks up languages from `CurrencyWebsiteLanguages` and stores them in the `website_languages` column.

```go
langs := currency.WebsiteLanguagesForCurrency("BRL") // ["pt"]
```

To add or change mappings, edit `website_languages.go` in this package.

## Database schema

`ApplySchema` is a **migration-time** call (see [What to run at startup](#what-to-run-at-startup)). It:

1. runs `CREATE SCHEMA IF NOT EXISTS "<schema>"`
2. creates tables, triggers, and indexes inside that schema:
   - `"<schema>"."currencies"`
   - `"<schema>"."currency_exchange_rates"`
   - `"<schema>"."currency_exchange_rate_history"`
3. seeds currency rows from the config (unless `skip_seed: true`)

This package does not replace goose or your migration tool — call `ApplySchema` from that step after `goose up`.

Use a dedicated schema when you want currency tables isolated from `public`:

```yaml
schema: billing
```

## ISO 4217 catalog vs enabled currencies

The package includes every ISO 4217 currency code as a read-only catalog (`catalog_iso4217.go`). Your config file defines which codes the application actually supports.

```go
entry, ok := currency.LookupISO4217("BRL")
codes := currency.AllISO4217Codes()
normalized := currency.NormalizeCode(" eur ") // "EUR"
```

The catalog validates config entries and supplies name/symbol values when seeding the database. It is not the runtime allow-list — that comes from `currencies` in your YAML file.

## Service API

Create the service once and reuse it:

```go
svc, err := currency.NewService(db, cfg, currency.Options{
    Logger: logger,
    APIURL: currency.DefaultAPIURL, // optional override
})
```

| Method | Description |
|--------|-------------|
| `ValidateCode(code)` | Checks ISO 4217 + enabled config list |
| `Convert(ctx, amount, from, to)` | Converts using stored USD-based rates |
| `SyncExchangeRates(ctx)` | Fetches feed and upserts rates for enabled currencies |
| `RunExchangeRateUpdater(ctx, interval)` | Background sync loop |
| `ListCurrencies(ctx)` | Lists seeded currencies filtered by config |
| `ListExchangeRates(ctx)` | Lists stored rates for enabled currencies |
| `GetExchangeRate(ctx, code)` | Returns one rate (not available for USD base) |
| `Config()` | Returns the loaded configuration |
| `Store()` | Low-level database access if needed |

Every method that accepts a currency code validates it against the configured list. Codes that exist in ISO 4217 but are not in your config return `ErrUnknownCurrency`.

### Convert amounts

```go
amount, err := svc.Convert(ctx, 100, "EUR", "BRL")
if err != nil {
    // validation error, missing rate, etc.
}
```

Conversion always goes through USD internally.

### Validate in handlers

```go
if err := svc.ValidateCode(req.Currency); err != nil {
    return apperror.ToGRPCStatus(err)
}
```

## Exchange rate feed

By default, rates are fetched from:

```
https://open.er-api.com/v6/latest/USD
```

`SyncExchangeRates` stores one rate per enabled non-USD currency. Override the URL with `Options.APIURL` if needed.

## Errors

Errors are structured `apperror.Error` values from `github.com/emersonary/appkit/apperror`:

```go
if appErr, ok := apperror.As(err); ok {
    switch appErr.Code {
    case currency.ErrUnknownCurrency.Code:
        // valid ISO code, not enabled in config
    case currency.ErrRateNotFound.Code:
        // no stored rate yet
    }
}
```

| Code | Meaning |
|------|---------|
| `CURRENCY_SCHEMA_REQUIRED` | `schema` missing from config |
| `CURRENCY_NOT_ENABLED` | valid ISO code, not in config |
| `CURRENCY_UNKNOWN_ISO4217` | unknown currency code |
| `CURRENCY_RATE_NOT_FOUND` | no exchange rate in database |
| `CURRENCY_INVALID_AMOUNT` | amount must be > 0 |
| `CURRENCY_SAME_CURRENCY` | `from` and `to` are identical |

Use `apperror.IsValidation(err)` to distinguish client errors from internal failures.

## Tests

From the `appkit` module root:

```bash
go test ./currency/...
```

## Reference integration (sahar)

| When | What runs |
|------|-----------|
| Migrations (`db.RunMigrations`) | `ApplySchema` after goose up |
| Server start (`cmd/server/main.go`) | `LoadConfig` → `NewService` → `RunExchangeRateUpdater` |

See `internal/db/migrate.go`, `internal/db/currency.go`, and `cmd/server/main.go`.
