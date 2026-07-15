package dbhist

import (
	"fmt"
	"strings"
)

func buildLoadTablesSQL(cfg Config) string {
	exclude := buildExcludePatternsSQL("c.relname", cfg.allExcludePatterns())
	audit := strings.ToLower(strings.TrimSpace(auditCommentMarker))
	repo := strings.ToLower(strings.TrimSpace(repoCommentMarker))
	return fmt.Sprintf(`
		SELECT
			n.nspname AS table_schema,
			c.relname AS table_name,
			lower(coalesce(obj_description(c.oid, 'pg_class'), '')) AS table_comment
		FROM pg_class c
		JOIN pg_namespace n ON n.oid = c.relnamespace
		WHERE c.relkind = 'r'
		  AND n.nspname NOT IN ('pg_catalog', 'information_schema', 'pg_toast')
		  AND n.nspname NOT LIKE 'pg_temp_%%'
		  AND n.nspname NOT LIKE 'pg_toast_temp_%%'
		  %s
		  AND (
		    strpos(lower(coalesce(obj_description(c.oid, 'pg_class'), '')), %s) > 0
		    OR strpos(lower(coalesce(obj_description(c.oid, 'pg_class'), '')), %s) > 0
		  )
		ORDER BY n.nspname, c.relname;
	`, exclude, quoteLiteral(audit), quoteLiteral(repo))
}

func buildLoadForeignKeysSQL(cfg Config) string {
	exclude := buildExcludePatternsSQL("src.relname", cfg.allExcludePatterns())
	audit := strings.ToLower(strings.TrimSpace(auditCommentMarker))
	repo := strings.ToLower(strings.TrimSpace(repoCommentMarker))
	commentPred := fmt.Sprintf(`(
		strpos(lower(coalesce(obj_description(src.oid, 'pg_class'), '')), %s) > 0
		OR strpos(lower(coalesce(obj_description(src.oid, 'pg_class'), '')), %s) > 0
	)`, quoteLiteral(audit), quoteLiteral(repo))
	return fmt.Sprintf(`
		SELECT
			src_ns.nspname AS child_schema,
			src.relname AS child_table,
			src_att.attname AS child_column,
			tgt_ns.nspname AS parent_schema,
			tgt.relname AS parent_table,
			tgt_att.attname AS parent_column
		FROM pg_constraint con
		JOIN pg_class src ON src.oid = con.conrelid
		JOIN pg_namespace src_ns ON src_ns.oid = src.relnamespace
		JOIN pg_class tgt ON tgt.oid = con.confrelid
		JOIN pg_namespace tgt_ns ON tgt_ns.oid = tgt.relnamespace
		JOIN unnest(con.conkey) WITH ORDINALITY AS src_cols(attnum, ord) ON TRUE
		JOIN unnest(con.confkey) WITH ORDINALITY AS tgt_cols(attnum, ord)
		  ON src_cols.ord = tgt_cols.ord
		JOIN pg_attribute src_att
		  ON src_att.attrelid = src.oid
		 AND src_att.attnum = src_cols.attnum
		JOIN pg_attribute tgt_att
		  ON tgt_att.attrelid = tgt.oid
		 AND tgt_att.attnum = tgt_cols.attnum
		WHERE con.contype = 'f'
		  AND src.relkind = 'r'
		  AND src_ns.nspname NOT IN ('pg_catalog', 'information_schema', 'pg_toast')
		  %s
		  AND %s
		ORDER BY child_schema, child_table, child_column;
	`, exclude, commentPred)
}
