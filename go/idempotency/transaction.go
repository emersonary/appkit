package idempotency

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
)

// ExecuteTx runs the business operation and idempotency completion in ONE PostgreSQL
// transaction. The callback MUST use tx for all writes; no external side effects.
func (i *Idempotency[T]) ExecuteTx(ctx context.Context, req Request, op func(context.Context, *sql.Tx) (T, error)) (Result[T], error) {
	var zero Result[T]
	if err := validate(req); err != nil {
		return zero, err
	}
	if op == nil {
		return zero, ErrInvalid
	}
	tx, err := i.db.BeginTx(ctx, nil)
	if err != nil {
		return zero, err
	}
	defer tx.Rollback()
	owner, err := token()
	if err != nil {
		return zero, err
	}
	var started sql.NullTime
	err = tx.QueryRowContext(ctx, `INSERT INTO `+recordsTable+`
  (str_scope, str_key, str_fingerprint, str_status, str_owner_token, dt_started)
  VALUES($1,$2,$3,'PROCESSING',$4,clock_timestamp())
  ON CONFLICT (str_scope, str_key) DO NOTHING RETURNING dt_started`, req.Scope, req.Key, req.Fingerprint, owner).Scan(&started)
	if errors.Is(err, sql.ErrNoRows) {
		var r record
		err = tx.QueryRowContext(ctx, selectRecord, req.Scope, req.Key).Scan(
			&r.fingerprint, &r.status, &r.owner, &r.result, &r.started, &r.completed, &r.lease)
		if err != nil {
			return zero, err
		}
		cached, err := decode[T](req, r)
		if err != nil {
			return zero, err
		}
		if err := tx.Commit(); err != nil {
			return zero, err
		}
		return cached, nil
	}
	if err != nil {
		return zero, err
	}
	value, err := op(ctx, tx)
	if err != nil {
		return zero, err
	}
	data, err := json.Marshal(value)
	if err != nil {
		return zero, fmt.Errorf("marshal result: %w", err)
	}
	var completed sql.NullTime
	err = tx.QueryRowContext(ctx, `UPDATE `+recordsTable+`
  SET str_status='COMPLETED', byte_result=$4, dt_completed=clock_timestamp(),
  str_owner_token=NULL, dt_updated=clock_timestamp()
  WHERE str_scope=$1 AND str_key=$2 AND str_owner_token=$3 AND str_status='PROCESSING'
  RETURNING dt_completed`, req.Scope, req.Key, owner, data).Scan(&completed)
	if err != nil {
		return zero, err
	}
	if err := tx.Commit(); err != nil {
		return zero, fmt.Errorf("commit outcome uncertain: %w", err)
	}
	return Result[T]{Value: value, StartedAt: started.Time, CompletedAt: completed.Time, Cached: false}, nil
}
