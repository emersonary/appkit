//go:build integration

package idempotency_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/emersonary/appkit/idempotency"
	_ "github.com/jackc/pgx/v5/stdlib"
)

// Integration tests require Postgres with idempotency migrations applied (via-jeri db reset).
//
//	VIA_JERI_DATABASE_PORT=15433 go test -tags=integration ./idempotency/... -count=1

func testDSN() string {
	if dsn := os.Getenv("APPKIT_TEST_DATABASE_DSN"); dsn != "" {
		return dsn
	}
	port := os.Getenv("VIA_JERI_DATABASE_PORT")
	if port == "" {
		port = "5432"
	}
	user := envOr("VIA_JERI_DATABASE_USER", "postgres")
	pass := envOr("VIA_JERI_DATABASE_PASSWORD", "postgres")
	db := envOr("VIA_JERI_DATABASE_NAME", "viajeri")
	host := envOr("VIA_JERI_DATABASE_HOST", "localhost")
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", user, pass, host, port, db)
}

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func openTestDB(t *testing.T) *sql.DB {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	t.Cleanup(cancel)
	db, err := sql.Open("pgx", testDSN())
	if err != nil {
		t.Skipf("postgres: %v", err)
	}
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		t.Skipf("postgres ping: %v", err)
	}
	var ready bool
	err = db.QueryRowContext(ctx, `
SELECT EXISTS (
  SELECT 1 FROM information_schema.tables
  WHERE table_schema = 'idempotency' AND table_name = 'tbl_idempotency_records'
)`).Scan(&ready)
	if err != nil || !ready {
		db.Close()
		t.Skip("idempotency.tbl_idempotency_records missing; run via-jeri: go run ./cmd/db reset")
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func ensureEffectsTable(t *testing.T, db *sql.DB) {
	t.Helper()
	ctx := context.Background()
	_, err := db.ExecContext(ctx, `
CREATE TABLE IF NOT EXISTS idempotency.test_side_effects (
  str_key text PRIMARY KEY,
  str_value text NOT NULL
)`)
	if err != nil {
		t.Fatal(err)
	}
}

func countEffects(t *testing.T, db *sql.DB, key string) int {
	t.Helper()
	var n int
	err := db.QueryRowContext(context.Background(),
		`SELECT COUNT(*) FROM idempotency.test_side_effects WHERE str_key = $1`, key).Scan(&n)
	if err != nil {
		t.Fatal(err)
	}
	return n
}

func countIdempotencyRows(t *testing.T, db *sql.DB, scope, key string) int {
	t.Helper()
	var n int
	err := db.QueryRowContext(context.Background(),
		`SELECT COUNT(*) FROM idempotency.tbl_idempotency_records WHERE str_scope = $1 AND str_key = $2`,
		scope, key).Scan(&n)
	if err != nil {
		t.Fatal(err)
	}
	return n
}

type booking struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

// TestExecuteTx_SuccessThenCached shows the main ExecuteTx win: one atomic commit, then duplicates read the stored result.
func TestExecuteTx_SuccessThenCached(t *testing.T) {
	db := openTestDB(t)
	ensureEffectsTable(t, db)
	engine, err := idempotency.New[booking](db, idempotency.Options{})
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	scope := "test:execute-tx"
	key := fmt.Sprintf("success-%d", time.Now().UnixNano())
	fp, err := idempotency.Fingerprint(map[string]string{"amount": "100"})
	if err != nil {
		t.Fatal(err)
	}
	req := idempotency.Request{Scope: scope, Key: key, Fingerprint: fp}
	effectKey := scope + ":" + key

	calls := 0
	first, err := engine.ExecuteTx(ctx, req, func(ctx context.Context, tx *sql.Tx) (booking, error) {
		calls++
		_, err := tx.ExecContext(ctx,
			`INSERT INTO idempotency.test_side_effects (str_key, str_value) VALUES ($1, $2)`,
			effectKey, "confirmed")
		if err != nil {
			return booking{}, err
		}
		return booking{ID: effectKey, Status: "CONFIRMED"}, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if first.Cached || calls != 1 {
		t.Fatalf("first: cached=%v calls=%d", first.Cached, calls)
	}
	if countEffects(t, db, effectKey) != 1 {
		t.Fatal("expected one side-effect row")
	}

	second, err := engine.ExecuteTx(ctx, req, func(ctx context.Context, tx *sql.Tx) (booking, error) {
		calls++
		t.Fatal("callback must not run on cache hit")
		return booking{}, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if !second.Cached || second.Value.ID != first.Value.ID {
		t.Fatalf("second: cached=%v value=%+v", second.Cached, second.Value)
	}
	if calls != 1 {
		t.Fatalf("expected 1 callback invocation, got %d", calls)
	}
}

// TestExecuteTx_FailureLeavesNoRow proves a failed attempt rolls back claim + business writes, so the same key can be retried.
func TestExecuteTx_FailureLeavesNoRow(t *testing.T) {
	db := openTestDB(t)
	ensureEffectsTable(t, db)
	engine, err := idempotency.New[booking](db, idempotency.Options{})
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	scope := "test:execute-tx"
	key := fmt.Sprintf("fail-retry-%d", time.Now().UnixNano())
	fp, _ := idempotency.Fingerprint("v1")
	req := idempotency.Request{Scope: scope, Key: key, Fingerprint: fp}
	effectKey := scope + ":" + key

	_, err = engine.ExecuteTx(ctx, req, func(ctx context.Context, tx *sql.Tx) (booking, error) {
		_, err := tx.ExecContext(ctx,
			`INSERT INTO idempotency.test_side_effects (str_key, str_value) VALUES ($1, $2)`,
			effectKey, "partial")
		if err != nil {
			return booking{}, err
		}
		return booking{}, errors.New("simulated business error")
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if countIdempotencyRows(t, db, scope, key) != 0 {
		t.Fatal("idempotency row should roll back with failed tx")
	}
	if countEffects(t, db, effectKey) != 0 {
		t.Fatal("side effect should roll back with failed tx")
	}

	ok, err := engine.ExecuteTx(ctx, req, func(ctx context.Context, tx *sql.Tx) (booking, error) {
		_, err := tx.ExecContext(ctx,
			`INSERT INTO idempotency.test_side_effects (str_key, str_value) VALUES ($1, $2)`,
			effectKey, "ok")
		return booking{ID: effectKey, Status: "CONFIRMED"}, err
	})
	if err != nil {
		t.Fatal(err)
	}
	if ok.Cached {
		t.Fatal("retry after rollback should run fresh, not cache")
	}
	if countEffects(t, db, effectKey) != 1 {
		t.Fatal("expected successful side effect after retry")
	}
}

// TestExecute_SuccessThenCached mirrors ExecuteTx for the happy path when work happens outside the transaction.
func TestExecute_SuccessThenCached(t *testing.T) {
	db := openTestDB(t)
	engine, err := idempotency.New[booking](db, idempotency.Options{Lease: time.Minute})
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	scope := "test:execute"
	key := fmt.Sprintf("ext-success-%d", time.Now().UnixNano())
	fp, _ := idempotency.Fingerprint(map[string]int{"n": 1})
	req := idempotency.Request{Scope: scope, Key: key, Fingerprint: fp}

	externalCalls := 0
	first, err := engine.Execute(ctx, req, func(ctx context.Context) (booking, error) {
		externalCalls++
		return booking{ID: "ext-1", Status: "CONFIRMED"}, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if first.Cached {
		t.Fatal("first call should not be cached")
	}

	second, err := engine.Execute(ctx, req, func(ctx context.Context) (booking, error) {
		externalCalls++
		t.Fatal("external op must not run again")
		return booking{}, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if !second.Cached || externalCalls != 1 {
		t.Fatalf("cached=%v externalCalls=%d", second.Cached, externalCalls)
	}
	st, err := engine.Inspect(ctx, req)
	if err != nil || st.Status != "COMPLETED" {
		t.Fatalf("inspect: %+v err=%v", st, err)
	}
}

// TestExecute_FailureMarksUncertain shows Execute persists state when the callback fails (side effect may have happened).
func TestExecute_FailureMarksUncertain(t *testing.T) {
	db := openTestDB(t)
	engine, err := idempotency.New[booking](db, idempotency.Options{Lease: time.Minute})
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	scope := "test:execute"
	key := fmt.Sprintf("ext-fail-%d", time.Now().UnixNano())
	fp, _ := idempotency.Fingerprint("fail-case")
	req := idempotency.Request{Scope: scope, Key: key, Fingerprint: fp}

	_, err = engine.Execute(ctx, req, func(ctx context.Context) (booking, error) {
		return booking{}, errors.New("provider timeout after charge")
	})
	if err == nil {
		t.Fatal("expected joined callback error")
	}
	st, err := engine.Inspect(ctx, req)
	if err != nil {
		t.Fatal(err)
	}
	if st.Status != "UNCERTAIN" {
		t.Fatalf("expected UNCERTAIN, got %q", st.Status)
	}
	if countIdempotencyRows(t, db, scope, key) != 1 {
		t.Fatal("record must remain for reconciliation")
	}

	_, err = engine.Execute(ctx, req, func(ctx context.Context) (booking, error) {
		return booking{}, nil
	})
	if !errors.Is(err, idempotency.ErrUncertain) {
		t.Fatalf("duplicate call should not auto-retry: %v", err)
	}
}
