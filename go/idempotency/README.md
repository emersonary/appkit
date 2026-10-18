# PostgreSQL idempotency for Go

Concrete generic `Idempotency[T]` initialized with `*sql.DB`. No storage interface.

## Setup

1. Apply AppKit migrate instructions (`IdempotencyInstructions()`), e.g. via-jeri `go run ./cmd/db reset` after platform instructions.
2. Table: `idempotency.tbl_idempotency_records` (DDL from [`schema.yaml`](schema.yaml) via schemagen).
3. Pass a `*sql.DB` (stdlib from `pgxpool` is fine).

## Choose the right method

- **`ExecuteTx` (recommended for PostgreSQL-only writes):** the business callback must use the supplied `*sql.Tx` for *all* database writes. The claim, business effects, and cached result commit atomically. Failed callbacks roll back. No HTTP calls, Kafka publishing, or other nontransactional effects inside this callback.
- **`Execute` (external side effects):** persists an atomic claim before calling the callback, then stores the result. If the callback fails, the record becomes `UNCERTAIN` because the side effect may have happened. If the process crashes, the claim stays `PROCESSING`; after its lease expires, duplicate calls return `ErrUncertain`. Do not blindly retry. Reconcile using the provider's stable idempotency key or read-after-write lookup, then call `ResolveCompleted` or `AllowRetry` only after proving retry safety.

## API

```go
engine, err := idempotency.New[Booking](db, idempotency.Options{})
result, err := engine.Execute(ctx, request, func(ctx context.Context) (Booking, error) {
    return externalProvider.CreateBooking(ctx, booking, request.Key)
})
```

`Result[T]` includes `Value`, `StartedAt`, `CompletedAt`, `Cached`.

`Request` includes `Scope`, `Key`, `Fingerprint`. Scope MUST include tenant and operation namespace. Fingerprint MUST cover the semantically relevant immutable request fields. Keys must be stable across Temporal retries. Do not use a random key per attempt.

## Safety limitations

See the upstream reference design in project docs; behavior matches the original `go-idempotency` reference (leases, `UNCERTAIN`, no automatic retry).

## Testing

`go test ./idempotency/...` from `appkit/go`.
