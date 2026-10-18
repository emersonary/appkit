//go:build integration

package migrate_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/emersonary/appkit/uuidv7"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Integration tests for platform.func_gen_uuid_v7(ts timestamptz default now).
//
// Run (Postgres migrated, e.g. via-jeri db reset):
//
//	VIA_JERI_DATABASE_PORT=15433 go test -tags=integration ./migrate/... -count=1
//
// Or set APPKIT_TEST_DATABASE_DSN to a full postgres URL.

func testDatabaseDSN() string {
	if dsn := os.Getenv("APPKIT_TEST_DATABASE_DSN"); dsn != "" {
		return dsn
	}
	port := os.Getenv("VIA_JERI_DATABASE_PORT")
	if port == "" {
		port = "5432"
	}
	user := os.Getenv("VIA_JERI_DATABASE_USER")
	if user == "" {
		user = "postgres"
	}
	pass := os.Getenv("VIA_JERI_DATABASE_PASSWORD")
	if pass == "" {
		pass = "postgres"
	}
	db := os.Getenv("VIA_JERI_DATABASE_NAME")
	if db == "" {
		db = "viajeri"
	}
	host := os.Getenv("VIA_JERI_DATABASE_HOST")
	if host == "" {
		host = "localhost"
	}
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", user, pass, host, port, db)
}

func openTestPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	t.Cleanup(cancel)

	pool, err := pgxpool.New(ctx, testDatabaseDSN())
	if err != nil {
		t.Skipf("postgres unavailable: %v", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		t.Skipf("postgres ping failed: %v", err)
	}
	if !platformUUIDFunctionReady(ctx, pool) {
		pool.Close()
		t.Skip("platform.func_gen_uuid_v7(timestamptz) not installed; run migrations (e.g. go run ./cmd/db reset)")
	}
	t.Cleanup(pool.Close)
	return pool
}

func platformUUIDFunctionReady(ctx context.Context, pool *pgxpool.Pool) bool {
	const q = `
SELECT EXISTS (
  SELECT 1
  FROM pg_proc p
  INNER JOIN pg_namespace n ON n.oid = p.pronamespace
  WHERE n.nspname = 'platform'
	  AND p.proname = 'func_gen_uuid_v7'
	  AND pg_get_function_identity_arguments(p.oid) = 'ts timestamp with time zone'
)`
	var ok bool
	return pool.QueryRow(ctx, q).Scan(&ok) == nil && ok
}

func queryUUIDAt(t *testing.T, ctx context.Context, pool *pgxpool.Pool, ts time.Time) uuid.UUID {
	t.Helper()
	var s string
	if err := pool.QueryRow(ctx, `SELECT platform.func_gen_uuid_v7($1::timestamptz)::text`, ts).Scan(&s); err != nil {
		t.Fatalf("func_gen_uuid_v7(ts): %v", err)
	}
	id, err := uuid.Parse(s)
	if err != nil {
		t.Fatalf("parse uuid: %v", err)
	}
	return id
}

func queryUUIDNow(t *testing.T, ctx context.Context, pool *pgxpool.Pool) uuid.UUID {
	t.Helper()
	var s string
	if err := pool.QueryRow(ctx, `SELECT platform.func_gen_uuid_v7()::text`).Scan(&s); err != nil {
		t.Fatalf("func_gen_uuid_v7(): %v", err)
	}
	id, err := uuid.Parse(s)
	if err != nil {
		t.Fatalf("parse uuid: %v", err)
	}
	return id
}

func TestPostgresFuncGenUUIDv7_FromTimeStampRoundTrip(t *testing.T) {
	pool := openTestPool(t)
	ctx := context.Background()

	ts := time.Date(2024, 11, 15, 12, 0, 0, 123456789, time.UTC)
	id := queryUUIDAt(t, ctx, pool, ts)
	got := uuidv7.TimeStampFromUUIDv7(id)
	if got.UnixMilli() != ts.UnixMilli() {
		t.Fatalf("millis mismatch: got %v want %v", got, ts)
	}
	if uuidv7.Version(id) != 7 {
		t.Fatalf("expected version 7, got %d", uuidv7.Version(id))
	}
}

func TestPostgresFuncGenUUIDv7_MillisMatchGoAtFixedTime(t *testing.T) {
	pool := openTestPool(t)
	ctx := context.Background()

	ts := time.Date(2024, 11, 15, 12, 0, 0, 987654321, time.UTC)
	goID := uuidv7.UUIDv7FromTimeStamp(ts)
	pgID := queryUUIDAt(t, ctx, pool, ts)

	if uuidv7.MillisFromUUIDv7(goID) != uuidv7.MillisFromUUIDv7(pgID) {
		t.Fatalf("millis: go=%d pg=%d", uuidv7.MillisFromUUIDv7(goID), uuidv7.MillisFromUUIDv7(pgID))
	}
	if uuidv7.TimeStampFromUUIDv7(goID).UnixMilli() != uuidv7.TimeStampFromUUIDv7(pgID).UnixMilli() {
		t.Fatalf("round-trip millis differ between go and pg")
	}
}

func TestPostgresFuncGenUUIDv7_VersionAndOrderByTimestamp(t *testing.T) {
	pool := openTestPool(t)
	ctx := context.Background()

	t1 := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	t2 := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	a := queryUUIDAt(t, ctx, pool, t1)
	b := queryUUIDAt(t, ctx, pool, t2)

	if uuidv7.Version(a) != 7 || uuidv7.Version(b) != 7 {
		t.Fatalf("expected version 7, got %d and %d", uuidv7.Version(a), uuidv7.Version(b))
	}
	if b.String() <= a.String() {
		t.Fatalf("expected lexicographic order: %s then %s", a, b)
	}
}

func TestPostgresFuncGenUUIDv7_DefaultArgNearNow(t *testing.T) {
	pool := openTestPool(t)
	ctx := context.Background()

	id := queryUUIDNow(t, ctx, pool)
	got := uuidv7.TimeStampFromUUIDv7(id)
	now := time.Now()
	slack := 2 * time.Second
	if got.Before(now.Add(-slack)) || got.After(now.Add(slack)) {
		t.Fatalf("embedded time out of range: got %v now %v", got, now)
	}
}
