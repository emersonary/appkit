package migrate

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

const bootstrapSQL = `
CREATE SCHEMA IF NOT EXISTS platform;

CREATE TABLE IF NOT EXISTS platform.schema_instructions (
	id           TEXT PRIMARY KEY,
	source       TEXT NOT NULL DEFAULT '',
	sql_checksum TEXT NOT NULL DEFAULT '',
	status       TEXT NOT NULL CHECK (status IN ('applied', 'failed')),
	applied_at   TIMESTAMPTZ,
	error        TEXT,
	duration_ms  BIGINT
);
`

type store struct {
	db *sql.DB
}

func newStore(db *sql.DB) *store {
	return &store{db: db}
}

func (s *store) ensureMeta(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx, bootstrapSQL)
	return err
}

func (s *store) getRecord(ctx context.Context, id string) (*Record, error) {
	var rec Record
	var appliedAt sql.NullTime
	err := s.db.QueryRowContext(ctx, `
SELECT id, source, sql_checksum, status, applied_at, COALESCE(error, ''), COALESCE(duration_ms, 0)
FROM platform.schema_instructions
WHERE id = $1`, id).Scan(
		&rec.ID,
		&rec.Source,
		&rec.SQLChecksum,
		&rec.Status,
		&appliedAt,
		&rec.Error,
		&rec.DurationMS,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if appliedAt.Valid {
		v := appliedAt.Time.UTC().Format(time.RFC3339Nano)
		rec.AppliedAt = &v
	}
	return &rec, nil
}

func (s *store) listRecords(ctx context.Context) ([]Record, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT id, source, sql_checksum, status, applied_at, COALESCE(error, ''), COALESCE(duration_ms, 0)
FROM platform.schema_instructions
ORDER BY applied_at NULLS LAST, id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Record
	for rows.Next() {
		var rec Record
		var appliedAt sql.NullTime
		if err := rows.Scan(
			&rec.ID,
			&rec.Source,
			&rec.SQLChecksum,
			&rec.Status,
			&appliedAt,
			&rec.Error,
			&rec.DurationMS,
		); err != nil {
			return nil, err
		}
		if appliedAt.Valid {
			v := appliedAt.Time.UTC().Format(time.RFC3339Nano)
			rec.AppliedAt = &v
		}
		out = append(out, rec)
	}
	return out, rows.Err()
}

func (s *store) markApplied(ctx context.Context, inst Instruction, duration time.Duration) error {
	_, err := s.db.ExecContext(ctx, `
INSERT INTO platform.schema_instructions (id, source, sql_checksum, status, applied_at, error, duration_ms)
VALUES ($1, $2, $3, 'applied', now(), NULL, $4)
ON CONFLICT (id) DO UPDATE SET
	source = EXCLUDED.source,
	sql_checksum = EXCLUDED.sql_checksum,
	status = 'applied',
	applied_at = now(),
	error = NULL,
	duration_ms = EXCLUDED.duration_ms`,
		inst.ID,
		inst.Source,
		checksumSQL(inst.SQL),
		duration.Milliseconds(),
	)
	return err
}

func (s *store) markFailed(ctx context.Context, inst Instruction, execErr error) error {
	_, err := s.db.ExecContext(ctx, `
INSERT INTO platform.schema_instructions (id, source, sql_checksum, status, applied_at, error, duration_ms)
VALUES ($1, $2, $3, 'failed', now(), $4, NULL)
ON CONFLICT (id) DO UPDATE SET
	source = EXCLUDED.source,
	sql_checksum = EXCLUDED.sql_checksum,
	status = 'failed',
	applied_at = now(),
	error = EXCLUDED.error,
	duration_ms = NULL`,
		inst.ID,
		inst.Source,
		checksumSQL(inst.SQL),
		execErr.Error(),
	)
	return err
}
