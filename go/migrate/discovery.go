package migrate

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"strings"
)

func loadTables(ctx context.Context, db *sql.DB, cfg Config) ([]Table, []SkippedTable, error) {
	query := buildLoadTablesSQL(cfg)

	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	var tables []Table
	var skipped []SkippedTable

	for rows.Next() {
		var schemaName string
		var tableName string
		var tableComment string

		if err := rows.Scan(&schemaName, &tableName, &tableComment); err != nil {
			return nil, nil, err
		}

		pk, err := loadPrimaryKey(ctx, db, schemaName, tableName)
		if err != nil {
			return nil, nil, err
		}

		if pk == "" {
			skipped = append(skipped, SkippedTable{
				Schema: schemaName,
				Table:  tableName,
				Reason: "no primary key found",
			})
			continue
		}

		cols, err := loadColumns(ctx, db, schemaName, tableName)
		if err != nil {
			return nil, nil, err
		}

		tables = append(tables, Table{
			Schema:     schemaName,
			Name:       tableName,
			PrimaryKey: pk,
			Audit:      commentHasMarker(tableComment, auditCommentMarker),
			Repo:       commentHasMarker(tableComment, repoCommentMarker),
			Columns:    cols,
		})
	}

	return tables, skipped, rows.Err()
}

func loadForeignKeys(ctx context.Context, db *sql.DB, cfg Config) ([]ForeignKey, error) {
	query := buildLoadForeignKeysSQL(cfg)

	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var foreignKeys []ForeignKey

	for rows.Next() {
		var fk ForeignKey
		if err := rows.Scan(
			&fk.ChildSchema,
			&fk.ChildTable,
			&fk.ChildColumn,
			&fk.ParentSchema,
			&fk.ParentTable,
			&fk.ParentColumn,
		); err != nil {
			return nil, err
		}

		isUnique, err := isUniqueColumn(ctx, db, fk.ChildSchema, fk.ChildTable, fk.ChildColumn)
		if err != nil {
			return nil, err
		}

		fk.IsOneToOne = isUnique
		foreignKeys = append(foreignKeys, fk)
	}

	return foreignKeys, rows.Err()
}

func buildExcludePatternsSQL(columnExpr string, patterns []string) string {
	if len(patterns) == 0 {
		return ""
	}

	var b strings.Builder
	for _, pattern := range patterns {
		b.WriteString(fmt.Sprintf("\n		  AND %s NOT LIKE %s", columnExpr, quoteLiteral(pattern)))
	}

	return b.String()
}

func loadPrimaryKey(ctx context.Context, db *sql.DB, schemaName, tableName string) (string, error) {
	var pk string

	err := db.QueryRowContext(ctx, `
		SELECT kcu.column_name
		FROM information_schema.table_constraints tc
		JOIN information_schema.key_column_usage kcu
		  ON tc.constraint_name = kcu.constraint_name
		 AND tc.table_schema = kcu.table_schema
		WHERE tc.table_schema = $1
		  AND tc.table_name = $2
		  AND tc.constraint_type = 'PRIMARY KEY'
		ORDER BY kcu.ordinal_position
		LIMIT 1;
	`, schemaName, tableName).Scan(&pk)

	if err == sql.ErrNoRows {
		return "", nil
	}

	return pk, err
}

func loadColumns(ctx context.Context, db *sql.DB, schemaName, tableName string) ([]Column, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT column_name, data_type, udt_name, is_nullable, column_default
		FROM information_schema.columns
		WHERE table_schema = $1
		  AND table_name = $2
		ORDER BY ordinal_position;
	`, schemaName, tableName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cols []Column

	for rows.Next() {
		var col Column
		var isNullable string
		var defaultValue sql.NullString

		if err := rows.Scan(&col.Name, &col.Type, &col.UdtName, &isNullable, &defaultValue); err != nil {
			return nil, err
		}

		col.IsNullable = strings.EqualFold(isNullable, "YES")
		if defaultValue.Valid {
			col.DefaultValue = defaultValue.String
		}

		cols = append(cols, col)
	}

	return cols, rows.Err()
}

func isUniqueColumn(ctx context.Context, db *sql.DB, schemaName, tableName, columnName string) (bool, error) {
	var isUnique bool

	err := db.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM pg_index i
			JOIN pg_class t ON t.oid = i.indrelid
			JOIN pg_namespace n ON n.oid = t.relnamespace
			JOIN pg_attribute a
			  ON a.attrelid = t.oid
			 AND a.attnum = ANY(i.indkey)
			WHERE n.nspname = $1
			  AND t.relname = $2
			  AND a.attname = $3
			  AND i.indisunique
		);
	`, schemaName, tableName, columnName).Scan(&isUnique)

	return isUnique, err
}

func attachChildRelations(tables []Table, foreignKeys []ForeignKey) []Table {
	tableIndex := make(map[string]int, len(tables))
	for i, table := range tables {
		tableIndex[tableKey(table.Schema, table.Name)] = i
	}

	for _, fk := range foreignKeys {
		parentIndex, ok := tableIndex[tableKey(fk.ParentSchema, fk.ParentTable)]
		if !ok {
			continue
		}

		childIndex, ok := tableIndex[tableKey(fk.ChildSchema, fk.ChildTable)]
		if !ok {
			continue
		}

		child := tables[childIndex]
		isOneToOne := fk.IsOneToOne || prefersSingleObjectJSON(child.Name)
		tables[parentIndex].Children = append(tables[parentIndex].Children, ChildRelation{
			Schema:     child.Schema,
			Table:      child.Name,
			PrimaryKey: child.PrimaryKey,
			FKColumn:   fk.ChildColumn,
			JSONKey:    jsonKeyForChild(child.Name, isOneToOne),
			IsOneToOne: isOneToOne,
		})
	}

	for i := range tables {
		sort.Slice(tables[i].Children, func(a, b int) bool {
			left := tables[i].Children[a]
			right := tables[i].Children[b]
			if left.Schema != right.Schema {
				return left.Schema < right.Schema
			}

			return left.Table < right.Table
		})
	}

	return tables
}

func jsonKeyForChild(childTable string, oneToOne bool) string {
	key := strings.TrimPrefix(childTable, "tbl_")
	if oneToOne {
		return key
	}

	return pluralJSONKey(key)
}

func pluralJSONKey(key string) string {
	if strings.HasSuffix(key, "data") || strings.HasSuffix(key, "luggage") {
		return key
	}

	if strings.HasSuffix(key, "s") {
		return key
	}

	return key + "s"
}

func prefersSingleObjectJSON(tableName string) bool {
	switch tableName {
	case "tbl_trip_luggage":
		return true
	default:
		return false
	}
}
