package migrate

import (
	"context"
	"database/sql"
	"fmt"
)

func buildRepoRegistrySQL() string {
	return `
CREATE SCHEMA IF NOT EXISTS audit;

CREATE TABLE IF NOT EXISTS audit.tbl_repo_function (
	id_repo_function BIGSERIAL PRIMARY KEY,
	str_schema_name TEXT NOT NULL,
	str_table_name TEXT NOT NULL,
	str_operation TEXT NOT NULL,
	int_version INT NOT NULL,
	str_function_name TEXT NOT NULL,
	str_definition_hash TEXT NOT NULL,
	txt_definition TEXT NOT NULL,
	boo_audit BOOLEAN NOT NULL DEFAULT false,
	boo_repo BOOLEAN NOT NULL DEFAULT false,
	dt_created TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	UNIQUE (str_schema_name, str_table_name, str_operation, int_version),
	UNIQUE (str_schema_name, str_table_name, str_operation, str_definition_hash),
	UNIQUE (str_schema_name, str_function_name)
);

ALTER TABLE audit.tbl_repo_function
	ADD COLUMN IF NOT EXISTS boo_audit BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE audit.tbl_repo_function
	ADD COLUMN IF NOT EXISTS boo_repo BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE audit.tbl_repo_function
	DROP COLUMN IF EXISTS dt_last_executed;
ALTER TABLE audit.tbl_repo_function
	DROP COLUMN IF EXISTS boo_dirty;

CREATE TABLE IF NOT EXISTS audit.tbl_repo_function_exec (
	str_schema_name TEXT NOT NULL,
	str_table_name TEXT NOT NULL,
	str_function_name TEXT NOT NULL,
	dt_last_executed TIMESTAMPTZ NOT NULL,
	PRIMARY KEY (str_schema_name, str_table_name, str_function_name)
);
`
}

func lookupRepoFunctionByHash(ctx context.Context, db *sql.DB, schema, table, operation, hash string) (resolvedFunc, bool, error) {
	var out resolvedFunc
	err := db.QueryRowContext(ctx, `
SELECT int_version, str_function_name, str_definition_hash
FROM audit.tbl_repo_function
WHERE str_schema_name = $1
  AND str_table_name = $2
  AND str_operation = $3
  AND str_definition_hash = $4
`, schema, table, operation, hash).Scan(&out.Version, &out.Name, &out.Hash)
	if err == sql.ErrNoRows {
		return resolvedFunc{}, false, nil
	}
	if err != nil {
		return resolvedFunc{}, false, err
	}
	return out, true, nil
}

func nextRepoFunctionVersion(ctx context.Context, db *sql.DB, schema, table, operation string) (int, error) {
	var max sql.NullInt64
	err := db.QueryRowContext(ctx, `
SELECT MAX(int_version)
FROM audit.tbl_repo_function
WHERE str_schema_name = $1
  AND str_table_name = $2
  AND str_operation = $3
`, schema, table, operation).Scan(&max)
	if err != nil {
		return 0, err
	}
	if !max.Valid {
		return 1, nil
	}
	return int(max.Int64) + 1, nil
}

func insertRepoFunction(ctx context.Context, db *sql.DB, schema, table, operation string, res resolvedFunc, definition string, audit, repo bool) error {
	_, err := db.ExecContext(ctx, `
INSERT INTO audit.tbl_repo_function (
	str_schema_name, str_table_name, str_operation,
	int_version, str_function_name, str_definition_hash, txt_definition,
	boo_audit, boo_repo
) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
ON CONFLICT (str_schema_name, str_table_name, str_operation, str_definition_hash) DO UPDATE
SET boo_audit = EXCLUDED.boo_audit,
    boo_repo = EXCLUDED.boo_repo
`, schema, table, operation, res.Version, res.Name, res.Hash, definition, audit, repo)
	if err != nil {
		return fmt.Errorf("insert repo function registry: %w", err)
	}
	return nil
}

func updateRepoFunctionFlags(ctx context.Context, db *sql.DB, schema, table, operation string, res resolvedFunc, audit, repo bool) error {
	_, err := db.ExecContext(ctx, `
UPDATE audit.tbl_repo_function
SET boo_audit = $5,
    boo_repo = $6
WHERE str_schema_name = $1
  AND str_table_name = $2
  AND str_operation = $3
  AND str_function_name = $4
`, schema, table, operation, res.Name, audit, repo)
	if err != nil {
		return fmt.Errorf("update repo function flags: %w", err)
	}
	return nil
}
