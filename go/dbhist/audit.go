package dbhist

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

func buildAuditMetadataSQL(tables []Table) string {
	var b strings.Builder

	b.WriteString(`
CREATE SCHEMA IF NOT EXISTS audit;

CREATE TABLE IF NOT EXISTS audit.tbl_audit_table (
	id_audit_table BIGSERIAL PRIMARY KEY,
	str_schema_name TEXT NOT NULL,
	str_table_name TEXT NOT NULL,
	UNIQUE (str_schema_name, str_table_name)
);

CREATE TABLE IF NOT EXISTS audit.tbl_audit_field (
	id_audit_field BIGSERIAL PRIMARY KEY,
	id_audit_table BIGINT NOT NULL REFERENCES audit.tbl_audit_table(id_audit_table),
	str_field_name TEXT NOT NULL,
	str_data_type TEXT NOT NULL,
	UNIQUE (id_audit_table, str_field_name)
);
`)

	for _, table := range tables {
		b.WriteString(buildTableRegistrationSQL(table))
	}

	return b.String()
}

func loadAuditIDs(ctx context.Context, db *sql.DB, tables []Table) ([]Table, error) {
	for ti := range tables {
		err := db.QueryRowContext(ctx, `
			SELECT id_audit_table
			FROM audit.tbl_audit_table
			WHERE str_schema_name = $1
			  AND str_table_name = $2;
		`, tables[ti].Schema, tables[ti].Name).Scan(&tables[ti].IDAuditTable)
		if err != nil {
			return nil, err
		}

		for ci := range tables[ti].Columns {
			err := db.QueryRowContext(ctx, `
				SELECT f.id_audit_field
				FROM audit.tbl_audit_field f
				WHERE f.id_audit_table = $1
				  AND f.str_field_name = $2;
			`, tables[ti].IDAuditTable, tables[ti].Columns[ci].Name).Scan(&tables[ti].Columns[ci].IDAuditField)
			if err != nil {
				return nil, err
			}
		}
	}

	return tables, nil
}

func buildTableRegistrationSQL(table Table) string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf(`
INSERT INTO audit.tbl_audit_table (
	str_schema_name,
	str_table_name
)
VALUES (
	%s,
	%s
)
ON CONFLICT (str_schema_name, str_table_name) DO NOTHING;
`, quoteLiteral(table.Schema), quoteLiteral(table.Name)))

	for _, col := range table.Columns {
		b.WriteString(fmt.Sprintf(`
INSERT INTO audit.tbl_audit_field (
	id_audit_table,
	str_field_name,
	str_data_type
)
SELECT
	id_audit_table,
	%s,
	%s
FROM audit.tbl_audit_table
WHERE str_schema_name = %s
  AND str_table_name = %s
ON CONFLICT (id_audit_table, str_field_name)
DO UPDATE SET str_data_type = EXCLUDED.str_data_type;
`,
			quoteLiteral(col.Name),
			quoteLiteral(col.Type),
			quoteLiteral(table.Schema),
			quoteLiteral(table.Name),
		))
	}

	return b.String()
}
