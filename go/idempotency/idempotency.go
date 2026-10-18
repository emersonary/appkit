package idempotency

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

const recordsTable = "idempotency.tbl_idempotency_records"

var (
	ErrInProgress    = errors.New("idempotency: operation in progress")
	ErrConflict      = errors.New("idempotency: key reused with different fingerprint")
	ErrUncertain     = errors.New("idempotency: previous execution outcome uncertain; reconcile before retry")
	ErrFailed        = errors.New("idempotency: previous execution failed; explicit recovery required")
	ErrLostOwnership = errors.New("idempotency: operation ownership lost")
	ErrInvalid       = errors.New("idempotency: invalid request")
)

type Request struct {
	Scope       string
	Key         string
	Fingerprint string
}

type Result[T any] struct {
	Value       T         `json:"value"`
	StartedAt   time.Time `json:"started_at"`
	CompletedAt time.Time `json:"completed_at"`
	Cached      bool      `json:"cached"`
}

type Options struct {
	Lease          time.Duration
	CleanupTimeout time.Duration
}

type Idempotency[T any] struct {
	db             *sql.DB
	lease          time.Duration
	cleanupTimeout time.Duration
}

func New[T any](db *sql.DB, opts Options) (*Idempotency[T], error) {
	if db == nil {
		return nil, errors.New("idempotency: nil database")
	}
	if opts.Lease == 0 {
		opts.Lease = 2 * time.Minute
	}
	if opts.CleanupTimeout == 0 {
		opts.CleanupTimeout = 10 * time.Second
	}
	if opts.Lease <= 0 || opts.CleanupTimeout <= 0 {
		return nil, ErrInvalid
	}
	return &Idempotency[T]{db: db, lease: opts.Lease, cleanupTimeout: opts.CleanupTimeout}, nil
}

func Fingerprint(v any) (string, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:]), nil
}

func validate(req Request) error {
	if req.Scope == "" || req.Key == "" || req.Fingerprint == "" {
		return ErrInvalid
	}
	return nil
}

func token() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

type record struct {
	fingerprint, status string
	owner               sql.NullString
	result              []byte
	started             time.Time
	completed           sql.NullTime
	lease               sql.NullTime
}

func scanRecord(row *sql.Row) (record, error) {
	var r record
	err := row.Scan(&r.fingerprint, &r.status, &r.owner, &r.result, &r.started, &r.completed, &r.lease)
	return r, err
}

const selectRecord = `SELECT str_fingerprint, str_status, str_owner_token, byte_result, dt_started, dt_completed, dt_lease_until
 FROM ` + recordsTable + ` WHERE str_scope=$1 AND str_key=$2`

func (i *Idempotency[T]) existing(ctx context.Context, req Request) (Result[T], error) {
	var zero Result[T]
	r, err := scanRecord(i.db.QueryRowContext(ctx, selectRecord, req.Scope, req.Key))
	if err != nil {
		return zero, err
	}
	return decode[T](req, r)
}

func decode[T any](req Request, r record) (Result[T], error) {
	var zero Result[T]
	if r.fingerprint != req.Fingerprint {
		return zero, ErrConflict
	}
	switch r.status {
	case "COMPLETED":
		if !r.completed.Valid {
			return zero, errors.New("idempotency: completed record missing timestamp")
		}
		var value T
		if err := json.Unmarshal(r.result, &value); err != nil {
			return zero, fmt.Errorf("decode cached result: %w", err)
		}
		return Result[T]{Value: value, StartedAt: r.started, CompletedAt: r.completed.Time, Cached: true}, nil
	case "PROCESSING":
		if r.lease.Valid && !time.Now().Before(r.lease.Time) {
			return zero, ErrUncertain
		}
		return zero, ErrInProgress
	case "UNCERTAIN":
		return zero, ErrUncertain
	case "FAILED":
		return zero, ErrFailed
	default:
		return zero, fmt.Errorf("idempotency: unknown status %q", r.status)
	}
}

// Execute coordinates calls across processes. External side effects MUST themselves be
// idempotent or reconcilable. Expired leases NEVER trigger automatic re-execution.
func (i *Idempotency[T]) Execute(ctx context.Context, req Request, op func(context.Context) (T, error)) (Result[T], error) {
	var zero Result[T]
	if err := validate(req); err != nil {
		return zero, err
	}
	if op == nil {
		return zero, ErrInvalid
	}
	owner, err := token()
	if err != nil {
		return zero, err
	}
	var started time.Time
	err = i.db.QueryRowContext(ctx, `INSERT INTO `+recordsTable+`
  (str_scope, str_key, str_fingerprint, str_status, str_owner_token, dt_started, dt_lease_until)
  VALUES ($1,$2,$3,'PROCESSING',$4,clock_timestamp(),clock_timestamp()+($5 * interval '1 second'))
  ON CONFLICT (str_scope, str_key) DO NOTHING RETURNING dt_started`, req.Scope, req.Key, req.Fingerprint, owner, i.lease.Seconds()).Scan(&started)
	if errors.Is(err, sql.ErrNoRows) {
		return i.existing(ctx, req)
	}
	if err != nil {
		return zero, err
	}

	value, opErr := op(ctx)
	cleanupCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), i.cleanupTimeout)
	defer cancel()
	if opErr != nil {
		_, markErr := i.db.ExecContext(cleanupCtx, `UPDATE `+recordsTable+`
   SET str_status='UNCERTAIN', str_error_message=$4, str_owner_token=NULL, dt_updated=clock_timestamp()
   WHERE str_scope=$1 AND str_key=$2 AND str_owner_token=$3 AND str_status='PROCESSING'`, req.Scope, req.Key, owner, opErr.Error())
		return zero, errors.Join(opErr, markErr)
	}
	data, err := json.Marshal(value)
	if err != nil {
		return zero, fmt.Errorf("encode operation result (outcome uncertain): %w", err)
	}
	var completed time.Time
	err = i.db.QueryRowContext(cleanupCtx, `UPDATE `+recordsTable+`
  SET str_status='COMPLETED', byte_result=$4, dt_completed=clock_timestamp(), str_owner_token=NULL,
      dt_lease_until=NULL, dt_updated=clock_timestamp()
  WHERE str_scope=$1 AND str_key=$2 AND str_owner_token=$3 AND str_status='PROCESSING'
  RETURNING dt_completed`, req.Scope, req.Key, owner, data).Scan(&completed)
	if errors.Is(err, sql.ErrNoRows) {
		return zero, ErrLostOwnership
	}
	if err != nil {
		return zero, fmt.Errorf("persist completion (outcome uncertain): %w", err)
	}
	return Result[T]{Value: value, StartedAt: started, CompletedAt: completed, Cached: false}, nil
}

// State is the current record for diagnostics and reconciliation.
type State struct {
	Status      string
	StartedAt   time.Time
	CompletedAt *time.Time
	LeaseUntil  *time.Time
}

func (i *Idempotency[T]) Inspect(ctx context.Context, req Request) (State, error) {
	if err := validate(req); err != nil {
		return State{}, err
	}
	r, err := scanRecord(i.db.QueryRowContext(ctx, selectRecord, req.Scope, req.Key))
	if err != nil {
		return State{}, err
	}
	if r.fingerprint != req.Fingerprint {
		return State{}, ErrConflict
	}
	s := State{Status: r.status, StartedAt: r.started}
	if r.completed.Valid {
		v := r.completed.Time
		s.CompletedAt = &v
	}
	if r.lease.Valid {
		v := r.lease.Time
		s.LeaseUntil = &v
	}
	return s, nil
}

// ResolveCompleted records a verified external outcome. Never call without verification.
func (i *Idempotency[T]) ResolveCompleted(ctx context.Context, req Request, value T) (Result[T], error) {
	var zero Result[T]
	if err := validate(req); err != nil {
		return zero, err
	}
	data, err := json.Marshal(value)
	if err != nil {
		return zero, err
	}
	var started, completed time.Time
	err = i.db.QueryRowContext(ctx, `UPDATE `+recordsTable+` SET
  str_status='COMPLETED', byte_result=$4, dt_completed=clock_timestamp(), str_owner_token=NULL,
  dt_lease_until=NULL, dt_updated=clock_timestamp()
  WHERE str_scope=$1 AND str_key=$2 AND str_fingerprint=$3
  AND (str_status='UNCERTAIN' OR (str_status='PROCESSING' AND dt_lease_until < clock_timestamp()))
  RETURNING dt_started, dt_completed`, req.Scope, req.Key, req.Fingerprint, data).Scan(&started, &completed)
	if errors.Is(err, sql.ErrNoRows) {
		return zero, ErrLostOwnership
	}
	if err != nil {
		return zero, err
	}
	return Result[T]{Value: value, StartedAt: started, CompletedAt: completed, Cached: false}, nil
}

// AllowRetry requires independent proof that no side effect occurred, or that retry
// is safe because the downstream system enforces the same stable idempotency key.
func (i *Idempotency[T]) AllowRetry(ctx context.Context, req Request) error {
	if err := validate(req); err != nil {
		return err
	}
	res, err := i.db.ExecContext(ctx, `DELETE FROM `+recordsTable+`
  WHERE str_scope=$1 AND str_key=$2 AND str_fingerprint=$3
  AND (str_status='UNCERTAIN' OR (str_status='PROCESSING' AND dt_lease_until < clock_timestamp()))`, req.Scope, req.Key, req.Fingerprint)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n != 1 {
		return ErrLostOwnership
	}
	return nil
}

// DeleteExpiredCompleted only removes completed records with expires_at set.
func (i *Idempotency[T]) DeleteExpiredCompleted(ctx context.Context) (int64, error) {
	r, err := i.db.ExecContext(ctx, `DELETE FROM `+recordsTable+`
 WHERE str_status='COMPLETED' AND dt_expires_at IS NOT NULL AND dt_expires_at < clock_timestamp()`)
	if err != nil {
		return 0, err
	}
	return r.RowsAffected()
}
