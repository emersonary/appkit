package migrate

import (
	"fmt"
	"strings"
)

func buildHistSQL(tables []Table) string {
	var b strings.Builder

	for _, table := range tables {
		b.WriteString(buildHistTableSQL(table))
		b.WriteString(buildHistDetailViewSQL(table))
		b.WriteString(buildTriggerFunctionSQL(table))
		b.WriteString(buildTriggerSQL(table))
	}

	return b.String()
}

func buildHistTableSQL(table Table) string {
	histTableName := table.Name + "_hist"
	histDetailTableName := table.Name + "_hist_detail"
	idHistColumn := table.PrimaryKey + "_hist"
	idHistDetailColumn := table.PrimaryKey + "_hist_detail"

	return fmt.Sprintf(`
CREATE TABLE IF NOT EXISTS %s (
	%s BIGSERIAL PRIMARY KEY,
	id_audit_table BIGINT NOT NULL REFERENCES audit.tbl_audit_table(id_audit_table),
	%s TEXT NOT NULL,
	str_operation CHAR(1) NOT NULL,
	str_changed_by TEXT,
	dt_changed_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS %s (
	%s BIGSERIAL PRIMARY KEY,
	%s BIGINT NOT NULL REFERENCES %s(%s),
	id_audit_field BIGINT NOT NULL REFERENCES audit.tbl_audit_field(id_audit_field),
	str_old_value TEXT,
	str_new_value TEXT
);

CREATE INDEX IF NOT EXISTS %s
ON %s(%s);

CREATE INDEX IF NOT EXISTS %s
ON %s(dt_changed_at);

CREATE INDEX IF NOT EXISTS %s
ON %s(%s);
`,
		quoteQualifiedIdent(table.Schema, histTableName),
		quoteIdent(idHistColumn),
		quoteIdent(table.PrimaryKey),

		quoteQualifiedIdent(table.Schema, histDetailTableName),
		quoteIdent(idHistDetailColumn),
		quoteIdent(idHistColumn),
		quoteQualifiedIdent(table.Schema, histTableName),
		quoteIdent(idHistColumn),

		quoteIdent("idx_"+histTableName+"_record"),
		quoteQualifiedIdent(table.Schema, histTableName),
		quoteIdent(table.PrimaryKey),

		quoteIdent("idx_"+histTableName+"_changed_at"),
		quoteQualifiedIdent(table.Schema, histTableName),

		quoteIdent("idx_"+histDetailTableName+"_hist"),
		quoteQualifiedIdent(table.Schema, histDetailTableName),
		quoteIdent(idHistColumn),
	)
}

func buildHistDetailViewSQL(table Table) string {
	histTableName := table.Name + "_hist"
	histDetailTableName := table.Name + "_hist_detail"

	viewName := "v_" + strings.TrimPrefix(histDetailTableName, "tbl_")

	idHistColumn := table.PrimaryKey + "_hist"

	return fmt.Sprintf(`
CREATE OR REPLACE VIEW %s AS
SELECT
	h.%s,
	h.%s,
	h.str_operation,
	h.str_changed_by,
	h.dt_changed_at,

	d.id_audit_field,
	f.str_field_name,
	f.str_data_type,

	d.str_old_value,
	d.str_new_value

FROM %s h
LEFT JOIN %s d
  ON d.%s = h.%s
LEFT JOIN audit.tbl_audit_field f
  ON f.id_audit_field = d.id_audit_field

ORDER BY h.%s;
`,
		quoteQualifiedIdent(table.Schema, viewName),

		quoteIdent(idHistColumn),
		quoteIdent(table.PrimaryKey),

		quoteQualifiedIdent(table.Schema, histTableName),
		quoteQualifiedIdent(table.Schema, histDetailTableName),

		quoteIdent(idHistColumn),
		quoteIdent(idHistColumn),

		quoteIdent(idHistColumn),
	)
}

func buildFieldHistSQL(table Table, col Column) string {
	histDetailTableName := table.Name + "_hist_detail"
	idHistColumn := table.PrimaryKey + "_hist"
	colName := quoteIdent(col.Name)

	return fmt.Sprintf(`
	IF OLD.%s IS DISTINCT FROM NEW.%s THEN
		INSERT INTO %s (
			%s,
			id_audit_field,
			str_old_value,
			str_new_value
		)
		VALUES (
			v_id_hist,
			%d,
			OLD.%s::TEXT,
			NEW.%s::TEXT
		);
	END IF;
`,
		colName,
		colName,
		quoteQualifiedIdent(table.Schema, histDetailTableName),
		quoteIdent(idHistColumn),
		col.IDAuditField,
		colName,
		colName,
	)
}

func buildTriggerFunctionSQL(table Table) string {
	functionName := "fn_audit_" + table.Name
	histTableName := table.Name + "_hist"
	idHistColumn := table.PrimaryKey + "_hist"
	hasRowStatus := tableHasColumn(table, "str_RowStatus")

	var b strings.Builder

	b.WriteString(fmt.Sprintf(`
CREATE OR REPLACE FUNCTION %s()
RETURNS TRIGGER AS $$
DECLARE
	v_id_hist BIGINT;
	v_operation CHAR(1);
	v_changed_by TEXT;
BEGIN
	v_changed_by := current_setting('app.current_user_id', true);

	IF TG_OP = 'INSERT' THEN
		INSERT INTO %s (
			id_audit_table,
			%s,
			str_operation,
			str_changed_by
		)
		VALUES (
			%d,
			NEW.%s::TEXT,
			'I',
			v_changed_by
		);

		RETURN NEW;
	END IF;
`,
		quoteQualifiedIdent(table.Schema, functionName),
		quoteQualifiedIdent(table.Schema, histTableName),
		quoteIdent(table.PrimaryKey),
		table.IDAuditTable,
		quoteIdent(table.PrimaryKey),
	))

	if hasRowStatus {
		b.WriteString(fmt.Sprintf(`
	IF OLD.%s IS DISTINCT FROM 'D'
	   AND NEW.%s = 'D' THEN
		v_operation := 'D';
	ELSE
		v_operation := 'U';
	END IF;
`,
			quoteIdent("str_RowStatus"),
			quoteIdent("str_RowStatus"),
		))
	} else {
		b.WriteString(`
	v_operation := 'U';
`)
	}

	b.WriteString(fmt.Sprintf(`
	INSERT INTO %s (
		id_audit_table,
		%s,
		str_operation,
		str_changed_by
	)
	VALUES (
		%d,
		NEW.%s::TEXT,
		v_operation,
		v_changed_by
	)
	RETURNING %s INTO v_id_hist;
`,
		quoteQualifiedIdent(table.Schema, histTableName),
		quoteIdent(table.PrimaryKey),
		table.IDAuditTable,
		quoteIdent(table.PrimaryKey),
		quoteIdent(idHistColumn),
	))

	for _, col := range table.Columns {
		if col.Name == table.PrimaryKey {
			continue
		}

		b.WriteString(buildFieldHistSQL(table, col))
	}

	b.WriteString(`
	RETURN NEW;
END;
$$ LANGUAGE plpgsql;
`)

	return b.String()
}

func buildTriggerSQL(table Table) string {
	triggerName := "trg_audit_" + table.Name
	functionName := "fn_audit_" + table.Name

	return fmt.Sprintf(`
DROP TRIGGER IF EXISTS %s ON %s;

CREATE TRIGGER %s
AFTER INSERT OR UPDATE ON %s
FOR EACH ROW
EXECUTE FUNCTION %s();
`,
		quoteIdent(triggerName),
		quoteQualifiedIdent(table.Schema, table.Name),
		quoteIdent(triggerName),
		quoteQualifiedIdent(table.Schema, table.Name),
		quoteQualifiedIdent(table.Schema, functionName),
	)
}
