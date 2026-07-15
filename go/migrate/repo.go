package migrate

import (
	"fmt"
	"sort"
	"strings"
)

type repoNames map[funcKey]string

func (n repoNames) qual(schema, table, operation string) string {
	name, ok := n[makeFuncKey(schema, table, operation)]
	if !ok {
		name = versionedRepoFunctionName(table, operation, 1)
	}
	return quoteQualifiedIdent(schema, name)
}

func buildRepoFunctionsSQL(tables []Table, names repoNames) (string, error) {
	if names == nil {
		names = repoNames{}
	}
	ordered, err := sortTablesChildrenFirst(tables)
	if err != nil {
		return "", err
	}

	var b strings.Builder
	tableMap := indexTables(tables)

	for _, table := range ordered {
		b.WriteString(buildInsertFunctionSQL(table, names))
		b.WriteString(buildUpdateFunctionSQL(table, names))
		b.WriteString(buildUpsertFunctionSQL(table, names))
		b.WriteString(buildDeleteFunctionSQL(table, tableMap, names))
		b.WriteString(buildGetFunctionSQL(table, tableMap, names))
		b.WriteString(buildListFunctionSQL(table, names))
	}

	return b.String(), nil
}

func indexTables(tables []Table) map[string]Table {
	tableMap := make(map[string]Table, len(tables))
	for _, table := range tables {
		tableMap[tableKey(table.Schema, table.Name)] = table
	}

	return tableMap
}

func sortTablesChildrenFirst(tables []Table) ([]Table, error) {
	const (
		unvisited = 0
		visiting  = 1
		visited   = 2
	)

	index := make(map[string]int, len(tables))
	depth := make(map[string]int, len(tables))
	state := make(map[string]int, len(tables))

	for i, table := range tables {
		key := tableKey(table.Schema, table.Name)
		index[key] = i
		depth[key] = 0
	}

	var visit func(table Table) error
	visit = func(table Table) error {
		key := tableKey(table.Schema, table.Name)
		switch state[key] {
		case visiting:
			return ErrRepoCycle.With("table", key)
		case visited:
			return nil
		}

		state[key] = visiting
		maxChildDepth := 0
		for _, child := range table.Children {
			childKey := tableKey(child.Schema, child.Table)
			childIndex, ok := index[childKey]
			if !ok {
				continue
			}

			if err := visit(tables[childIndex]); err != nil {
				return err
			}
			if depth[childKey] > maxChildDepth {
				maxChildDepth = depth[childKey]
			}
		}

		depth[key] = maxChildDepth + 1
		state[key] = visited
		return nil
	}

	for _, table := range tables {
		if err := visit(table); err != nil {
			return nil, err
		}
	}

	sorted := append([]Table(nil), tables...)
	sort.Slice(sorted, func(i, j int) bool {
		left := depth[tableKey(sorted[i].Schema, sorted[i].Name)]
		right := depth[tableKey(sorted[j].Schema, sorted[j].Name)]
		if left != right {
			return left < right
		}

		if sorted[i].Schema != sorted[j].Schema {
			return sorted[i].Schema < sorted[j].Schema
		}

		return sorted[i].Name < sorted[j].Name
	})

	return sorted, nil
}


func buildInsertFunctionSQL(table Table, names repoNames) string {
	rowStatusColumn := findRowStatusColumn(table)
	insertColumns := insertableColumns(table, rowStatusColumn)
	pkVar := "v_" + strings.ToLower(table.PrimaryKey)

	var b strings.Builder

	b.WriteString(fmt.Sprintf(`
CREATE OR REPLACE FUNCTION %s(payload JSONB)
RETURNS JSONB
LANGUAGE plpgsql
AS $$
DECLARE
	%s UUID;
	v_result JSONB;
	v_child JSONB;
	v_item JSONB;
BEGIN
`,
		names.qual(table.Schema, table.Name, "insert"),
		pkVar,
	))

	if isUUIDColumn(table, table.PrimaryKey) {
		b.WriteString(fmt.Sprintf(
			"\t%s := COALESCE((payload->>%s)::UUID, gen_random_uuid());\n\n",
			pkVar,
			quoteLiteral(table.PrimaryKey),
		))
	} else {
		b.WriteString(fmt.Sprintf(
			"\tIF NOT (payload ? %s) THEN\n\t\tRAISE EXCEPTION 'missing primary key %%', %s;\n\tEND IF;\n\n",
			quoteLiteral(table.PrimaryKey),
			quoteLiteral(table.PrimaryKey),
		))
	}

	b.WriteString("\tINSERT INTO ")
	b.WriteString(quoteQualifiedIdent(table.Schema, table.Name))
	b.WriteString(" (\n\t\t")
	b.WriteString(quoteIdent(table.PrimaryKey))
	for _, column := range insertColumns {
		b.WriteString(",\n\t\t")
		b.WriteString(quoteIdent(column.Name))
	}
	b.WriteString("\n\t) VALUES (\n\t\t")
	b.WriteString(pkVar)
	for _, column := range insertColumns {
		b.WriteString(",\n\t\t")
		b.WriteString(jsonValueExpr(column, "payload"))
	}
	b.WriteString("\n\t);\n\n")

	b.WriteString(fmt.Sprintf(
		"\tv_result := jsonb_build_object(%s, %s);\n\n",
		quoteLiteral(table.PrimaryKey),
		pkVar,
	))

	b.WriteString(buildChildInsertCallsSQL(table, pkVar, names))

	b.WriteString("\n\tRETURN v_result;\nEND;\n$$;\n")

	return b.String()
}

func buildChildInsertCallsSQL(table Table, pkVar string, names repoNames) string {
	var b strings.Builder

	for _, child := range table.Children {
		functionName := names.qual(child.Schema, child.Table, "insert")

		if child.IsOneToOne {
			b.WriteString(fmt.Sprintf(`
	IF payload ? %s AND jsonb_typeof(payload->%s) = 'object' THEN
		v_child := (payload->%s) || jsonb_build_object(%s, %s);
		PERFORM %s(v_child);
	END IF;
`,
				quoteLiteral(child.JSONKey),
				quoteLiteral(child.JSONKey),
				quoteLiteral(child.JSONKey),
				quoteLiteral(child.FKColumn),
				pkVar,
				functionName,
			))
			continue
		}

		b.WriteString(fmt.Sprintf(`
	IF payload ? %s THEN
		IF jsonb_typeof(payload->%s) = 'array' THEN
			FOR v_item IN
				SELECT value
				FROM jsonb_array_elements(payload->%s) AS t(value)
			LOOP
				v_item := v_item || jsonb_build_object(%s, %s);
				PERFORM %s(v_item);
			END LOOP;
		ELSIF jsonb_typeof(payload->%s) = 'object' THEN
			v_item := (payload->%s) || jsonb_build_object(%s, %s);
			PERFORM %s(v_item);
		END IF;
	END IF;
`,
			quoteLiteral(child.JSONKey),
			quoteLiteral(child.JSONKey),
			quoteLiteral(child.JSONKey),
			quoteLiteral(child.FKColumn),
			pkVar,
			functionName,
			quoteLiteral(child.JSONKey),
			quoteLiteral(child.JSONKey),
			quoteLiteral(child.FKColumn),
			pkVar,
			functionName,
		))
	}

	return b.String()
}

func buildUpdateFunctionSQL(table Table, names repoNames) string {
	rowStatusColumn := findRowStatusColumn(table)
	updatableCols := updatableColumns(table, rowStatusColumn)
	pkVar := "v_" + strings.ToLower(table.PrimaryKey)

	var setClauses []string
	for _, column := range updatableCols {
		setClauses = append(setClauses, fmt.Sprintf(
			"%s = %s",
			quoteIdent(column.Name),
			updateValueExpr(column, "payload"),
		))
	}

	var b strings.Builder

	b.WriteString(fmt.Sprintf(`
CREATE OR REPLACE FUNCTION %s(payload JSONB)
RETURNS JSONB
LANGUAGE plpgsql
AS $$
DECLARE
	%s UUID;
	v_result JSONB;
	v_child JSONB;
	v_item JSONB;
BEGIN
	IF NOT (payload ? %s) THEN
		RAISE EXCEPTION 'missing primary key %%', %s;
	END IF;

	%s := (payload->>%s)::UUID;

	UPDATE %s
	SET
		%s
	WHERE %s = %s
	  AND COALESCE(%s, 'A') <> 'D';

	IF NOT FOUND THEN
		RAISE EXCEPTION 'record not found for update in %%.%%',
			%s,
			%s;
	END IF;

	v_result := jsonb_build_object(%s, %s);
`,
		names.qual(table.Schema, table.Name, "update"),
		pkVar,
		quoteLiteral(table.PrimaryKey),
		quoteLiteral(table.PrimaryKey),
		pkVar,
		quoteLiteral(table.PrimaryKey),
		quoteQualifiedIdent(table.Schema, table.Name),
		strings.Join(setClauses, ",\n\t\t"),
		quoteIdent(table.PrimaryKey),
		pkVar,
		rowStatusSQLRef(table, rowStatusColumn),
		quoteLiteral(table.Schema),
		quoteLiteral(table.Name),
		quoteLiteral(table.PrimaryKey),
		pkVar,
	))

	b.WriteString(buildChildUpsertCallsSQL(table, pkVar, names))
	b.WriteString("\n\tRETURN v_result;\nEND;\n$$;\n")

	return b.String()
}

func buildChildUpsertCallsSQL(table Table, pkVar string, names repoNames) string {
	var b strings.Builder

	for _, child := range table.Children {
		upsertFunction := names.qual(child.Schema, child.Table, "upsert")

		if child.IsOneToOne {
			b.WriteString(fmt.Sprintf(`
	IF payload ? %s AND jsonb_typeof(payload->%s) = 'object' THEN
		v_child := (payload->%s) || jsonb_build_object(%s, %s);
		PERFORM %s(v_child);
	END IF;
`,
				quoteLiteral(child.JSONKey),
				quoteLiteral(child.JSONKey),
				quoteLiteral(child.JSONKey),
				quoteLiteral(child.FKColumn),
				pkVar,
				upsertFunction,
			))
			continue
		}

		b.WriteString(fmt.Sprintf(`
	IF payload ? %s THEN
		IF jsonb_typeof(payload->%s) = 'array' THEN
			FOR v_item IN
				SELECT value
				FROM jsonb_array_elements(payload->%s) AS t(value)
			LOOP
				v_item := v_item || jsonb_build_object(%s, %s);
				PERFORM %s(v_item);
			END LOOP;
		ELSIF jsonb_typeof(payload->%s) = 'object' THEN
			v_item := (payload->%s) || jsonb_build_object(%s, %s);
			PERFORM %s(v_item);
		END IF;
	END IF;
`,
			quoteLiteral(child.JSONKey),
			quoteLiteral(child.JSONKey),
			quoteLiteral(child.JSONKey),
			quoteLiteral(child.FKColumn),
			pkVar,
			upsertFunction,
			quoteLiteral(child.JSONKey),
			quoteLiteral(child.JSONKey),
			quoteLiteral(child.FKColumn),
			pkVar,
			upsertFunction,
		))
	}

	return b.String()
}

func buildUpsertFunctionSQL(table Table, names repoNames) string {
	insertFunction := names.qual(table.Schema, table.Name, "insert")
	updateFunction := names.qual(table.Schema, table.Name, "update")
	rowStatusColumn := findRowStatusColumn(table)

	var existsCondition string
	if rowStatusColumn != "" {
		existsCondition = fmt.Sprintf(
			"COALESCE(%s, 'A') <> 'D'",
			rowStatusSQLRef(table, rowStatusColumn),
		)
	} else {
		existsCondition = "TRUE"
	}

	return fmt.Sprintf(`
CREATE OR REPLACE FUNCTION %s(payload JSONB)
RETURNS JSONB
LANGUAGE plpgsql
AS $$
BEGIN
	IF payload ? %s
	   AND EXISTS (
			SELECT 1
			FROM %s
			WHERE %s = (payload->>%s)::UUID
			  AND %s
	   ) THEN
		RETURN %s(payload);
	END IF;

	RETURN %s(payload);
END;
$$;
`,
		names.qual(table.Schema, table.Name, "upsert"),
		quoteLiteral(table.PrimaryKey),
		quoteQualifiedIdent(table.Schema, table.Name),
		quoteIdent(table.PrimaryKey),
		quoteLiteral(table.PrimaryKey),
		existsCondition,
		updateFunction,
		insertFunction,
	)
}

func buildDeleteFunctionSQL(table Table, tableMap map[string]Table, names repoNames) string {
	rowStatusColumn := findRowStatusColumn(table)
	pkVar := "v_" + strings.ToLower(table.PrimaryKey)

	var b strings.Builder

	b.WriteString(fmt.Sprintf(`
CREATE OR REPLACE FUNCTION %s(payload JSONB)
RETURNS JSONB
LANGUAGE plpgsql
AS $$
DECLARE
	%s UUID;
	v_child_id UUID;
BEGIN
	IF NOT (payload ? %s) THEN
		RAISE EXCEPTION 'missing primary key %%', %s;
	END IF;

	%s := (payload->>%s)::UUID;
`,
		names.qual(table.Schema, table.Name, "delete"),
		pkVar,
		quoteLiteral(table.PrimaryKey),
		quoteLiteral(table.PrimaryKey),
		pkVar,
		quoteLiteral(table.PrimaryKey),
	))

	b.WriteString(buildChildDeleteCallsSQL(table, tableMap, pkVar, names))

	if rowStatusColumn != "" {
		b.WriteString(fmt.Sprintf(`
	UPDATE %s
	SET %s = 'D'
	WHERE %s = %s
	  AND COALESCE(%s, 'A') <> 'D';
`,
			quoteQualifiedIdent(table.Schema, table.Name),
			quoteIdent(rowStatusColumn),
			quoteIdent(table.PrimaryKey),
			pkVar,
			quoteIdent(rowStatusColumn),
		))
	} else {
		b.WriteString(fmt.Sprintf(`
	DELETE FROM %s
	WHERE %s = %s;
`,
			quoteQualifiedIdent(table.Schema, table.Name),
			quoteIdent(table.PrimaryKey),
			pkVar,
		))
	}

	b.WriteString(fmt.Sprintf(`
	RETURN jsonb_build_object(%s, %s, 'deleted', TRUE);
END;
$$;
`, quoteLiteral(table.PrimaryKey), pkVar))

	return b.String()
}

func buildChildDeleteCallsSQL(table Table, tableMap map[string]Table, pkVar string, names repoNames) string {
	var b strings.Builder

	for _, child := range table.Children {
		childTable := tableMap[tableKey(child.Schema, child.Table)]
		childRowStatusColumn := findRowStatusColumn(childTable)
		deleteFunction := names.qual(child.Schema, child.Table, "delete")

		activeFilter := "TRUE"
		if childRowStatusColumn != "" {
			activeFilter = fmt.Sprintf("COALESCE(%s, 'A') <> 'D'", quoteIdent(childRowStatusColumn))
		}

		b.WriteString(fmt.Sprintf(`
	FOR v_child_id IN
		SELECT %s
		FROM %s
		WHERE %s = %s
		  AND %s
	LOOP
		PERFORM %s(jsonb_build_object(%s, v_child_id));
	END LOOP;
`,
			quoteIdent(child.PrimaryKey),
			quoteQualifiedIdent(child.Schema, child.Table),
			quoteIdent(child.FKColumn),
			pkVar,
			activeFilter,
			deleteFunction,
			quoteLiteral(child.PrimaryKey),
		))
	}

	return b.String()
}

func buildGetFunctionSQL(table Table, tableMap map[string]Table, names repoNames) string {
	rowStatusColumn := findRowStatusColumn(table)
	pkVar := "v_" + strings.ToLower(table.PrimaryKey)
	rowVar := "v_row"

	var b strings.Builder

	b.WriteString(fmt.Sprintf(`
CREATE OR REPLACE FUNCTION %s(payload JSONB)
RETURNS JSONB
LANGUAGE plpgsql
AS $$
DECLARE
	%s UUID;
	%s %s;
	v_result JSONB;
	v_child JSONB;
BEGIN
	IF NOT (payload ? %s) THEN
		RAISE EXCEPTION 'missing primary key %%', %s;
	END IF;

	%s := (payload->>%s)::UUID;

	SELECT *
	INTO %s
	FROM %s
	WHERE %s = %s
	  AND %s;

	IF NOT FOUND THEN
		RAISE EXCEPTION 'record not found for get in %%.%%',
			%s,
			%s;
	END IF;

	v_result := %s;
`,
		names.qual(table.Schema, table.Name, "get"),
		pkVar,
		rowVar,
		quoteQualifiedIdent(table.Schema, table.Name),
		quoteLiteral(table.PrimaryKey),
		quoteLiteral(table.PrimaryKey),
		pkVar,
		quoteLiteral(table.PrimaryKey),
		rowVar,
		quoteQualifiedIdent(table.Schema, table.Name),
		quoteIdent(table.PrimaryKey),
		pkVar,
		activeRowFilter(table, rowVar, rowStatusColumn),
		quoteLiteral(table.Schema),
		quoteLiteral(table.Name),
		buildRowToJSONObject(table, rowVar),
	))

	b.WriteString(buildChildGetCallsSQL(table, tableMap, pkVar, names))
	b.WriteString("\n\tRETURN v_result;\nEND;\n$$;\n")

	return b.String()
}

func buildListFunctionSQL(table Table, names repoNames) string {
	rowStatusColumn := findRowStatusColumn(table)
	getFunction := names.qual(table.Schema, table.Name, "get")

	return fmt.Sprintf(`
CREATE OR REPLACE FUNCTION %s(payload JSONB)
RETURNS JSONB
LANGUAGE plpgsql
AS $$
DECLARE
	v_result JSONB;
BEGIN
	IF payload IS NULL THEN
		payload := '{}'::jsonb;
	END IF;

	SELECT COALESCE(
		jsonb_agg(
			%s(jsonb_build_object(%s, r.%s))
			ORDER BY r.%s
		),
		'[]'::jsonb
	)
	INTO v_result
	FROM %s r
	WHERE %s
	  AND %s;

	RETURN v_result;
END;
$$;
`,
		names.qual(table.Schema, table.Name, "list"),
		getFunction,
		quoteLiteral(table.PrimaryKey),
		quoteIdent(table.PrimaryKey),
		quoteIdent(table.PrimaryKey),
		quoteQualifiedIdent(table.Schema, table.Name),
		activeRowFilter(table, "r", rowStatusColumn),
		buildListFilterSQL(table, "payload"),
	)
}

func buildListFilterSQL(table Table, payloadVar string) string {
	parts := make([]string, 0, len(table.Columns))
	for _, column := range table.Columns {
		parts = append(parts, fmt.Sprintf(
			`(NOT (%s ? %s) OR r.%s IS NOT DISTINCT FROM %s)`,
			payloadVar,
			quoteLiteral(column.Name),
			quoteIdent(column.Name),
			jsonCastExpr(column, payloadVar),
		))
	}
	if len(parts) == 0 {
		return "TRUE"
	}
	return strings.Join(parts, "\n\t  AND ")
}

func buildChildGetCallsSQL(table Table, tableMap map[string]Table, pkVar string, names repoNames) string {
	var b strings.Builder

	for _, child := range table.Children {
		childTable := tableMap[tableKey(child.Schema, child.Table)]
		childRowStatusColumn := findRowStatusColumn(childTable)
		getFunction := names.qual(child.Schema, child.Table, "get")

		activeFilter := "TRUE"
		if childRowStatusColumn != "" {
			activeFilter = fmt.Sprintf(
				"COALESCE(c.%s, 'A') <> 'D'",
				quoteIdent(childRowStatusColumn),
			)
		}

		if child.IsOneToOne {
			b.WriteString(fmt.Sprintf(`
	SELECT %s(jsonb_build_object(%s, c.%s))
	INTO v_child
	FROM %s c
	WHERE c.%s = %s
	  AND %s
	LIMIT 1;

	IF v_child IS NOT NULL THEN
		v_result := v_result || jsonb_build_object(%s, v_child);
	END IF;
`,
				getFunction,
				quoteLiteral(child.PrimaryKey),
				quoteIdent(child.PrimaryKey),
				quoteQualifiedIdent(child.Schema, child.Table),
				quoteIdent(child.FKColumn),
				pkVar,
				activeFilter,
				quoteLiteral(child.JSONKey),
			))
			continue
		}

		b.WriteString(fmt.Sprintf(`
	v_result := v_result || jsonb_build_object(
		%s,
		COALESCE((
			SELECT jsonb_agg(%s(jsonb_build_object(%s, c.%s)) ORDER BY c.%s)
			FROM %s c
			WHERE c.%s = %s
			  AND %s
		), '[]'::jsonb)
	);
`,
			quoteLiteral(child.JSONKey),
			getFunction,
			quoteLiteral(child.PrimaryKey),
			quoteIdent(child.PrimaryKey),
			quoteIdent(child.PrimaryKey),
			quoteQualifiedIdent(child.Schema, child.Table),
			quoteIdent(child.FKColumn),
			pkVar,
			activeFilter,
		))
	}

	return b.String()
}

func buildRowToJSONObject(table Table, rowVar string) string {
	parts := make([]string, 0, len(table.Columns)*2)

	for _, column := range table.Columns {
		parts = append(parts,
			quoteLiteral(column.Name),
			fmt.Sprintf("%s.%s", rowVar, quoteIdent(column.Name)),
		)
	}

	return "jsonb_build_object(" + strings.Join(parts, ", ") + ")"
}

func activeRowFilter(table Table, rowAlias string, rowStatusColumn string) string {
	if rowStatusColumn == "" {
		return "TRUE"
	}

	return fmt.Sprintf(
		"COALESCE(%s.%s, 'A') <> 'D'",
		rowAlias,
		quoteIdent(rowStatusColumn),
	)
}

func insertableColumns(table Table, rowStatusColumn string) []Column {
	var columns []Column

	for _, column := range table.Columns {
		if column.Name == table.PrimaryKey {
			continue
		}

		if rowStatusColumn != "" && strings.EqualFold(column.Name, rowStatusColumn) {
			continue
		}

		columns = append(columns, column)
	}

	return columns
}

func updatableColumns(table Table, rowStatusColumn string) []Column {
	return insertableColumns(table, rowStatusColumn)
}

func findRowStatusColumn(table Table) string {
	for _, column := range table.Columns {
		if strings.EqualFold(column.Name, "str_rowstatus") {
			return column.Name
		}
	}

	return ""
}

func rowStatusSQLRef(table Table, rowStatusColumn string) string {
	if rowStatusColumn == "" {
		return "'A'"
	}

	return quoteIdent(rowStatusColumn)
}

func isUUIDColumn(table Table, columnName string) bool {
	for _, column := range table.Columns {
		if column.Name == columnName {
			return strings.EqualFold(column.Type, "uuid") || strings.EqualFold(column.UdtName, "uuid")
		}
	}

	return strings.HasPrefix(strings.ToLower(columnName), "id_")
}

func jsonValueExpr(column Column, payloadVar string) string {
	if column.DefaultValue != "" && !column.IsNullable {
		return fmt.Sprintf(
			"COALESCE(%s, %s)",
			jsonCastExpr(column, payloadVar),
			sqlDefaultLiteral(column),
		)
	}

	return jsonCastExpr(column, payloadVar)
}

func updateValueExpr(column Column, payloadVar string) string {
	return fmt.Sprintf(
		"CASE WHEN %s ? %s THEN %s ELSE %s END",
		payloadVar,
		quoteLiteral(column.Name),
		jsonCastExpr(column, payloadVar),
		quoteIdent(column.Name),
	)
}

func jsonCastExpr(column Column, payloadVar string) string {
	payloadAccess := fmt.Sprintf("%s->>%s", payloadVar, quoteLiteral(column.Name))

	switch strings.ToLower(column.Type) {
	case "uuid":
		return fmt.Sprintf("(%s)::UUID", payloadAccess)
	case "bigint":
		return fmt.Sprintf("(%s)::BIGINT", payloadAccess)
	case "integer", "smallint":
		return fmt.Sprintf("(%s)::INTEGER", payloadAccess)
	case "boolean":
		return fmt.Sprintf("(%s)::BOOLEAN", payloadAccess)
	case "double precision", "numeric", "real":
		return fmt.Sprintf("(%s)::DOUBLE PRECISION", payloadAccess)
	case "timestamp with time zone":
		return fmt.Sprintf("(%s)::TIMESTAMPTZ", payloadAccess)
	case "timestamp without time zone":
		return fmt.Sprintf("(%s)::TIMESTAMP", payloadAccess)
	case "json", "jsonb":
		return fmt.Sprintf("(%s)::JSONB", fmt.Sprintf("%s->%s", payloadVar, quoteLiteral(column.Name)))
	default:
		return payloadAccess
	}
}

func sqlDefaultLiteral(column Column) string {
	defaultValue := strings.TrimSpace(column.DefaultValue)
	if defaultValue == "" {
		if column.IsNullable {
			return "NULL"
		}

		return "''"
	}

	if strings.HasPrefix(defaultValue, "nextval(") {
		return "NULL"
	}

	if strings.Contains(defaultValue, "::") {
		return defaultValue
	}

	if strings.EqualFold(defaultValue, "true") || strings.EqualFold(defaultValue, "false") {
		return defaultValue
	}

	if defaultValue == "now()" || strings.HasPrefix(defaultValue, "now(") {
		return "now()"
	}

	if strings.EqualFold(column.Type, "boolean") {
		return defaultValue
	}

	if strings.EqualFold(column.Type, "integer") ||
		strings.EqualFold(column.Type, "bigint") ||
		strings.EqualFold(column.Type, "smallint") ||
		strings.EqualFold(column.Type, "numeric") ||
		strings.EqualFold(column.Type, "double precision") {
		return defaultValue
	}

	return quoteLiteral(strings.Trim(defaultValue, "'"))
}
