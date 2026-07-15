package migrate

import "strings"

func quoteIdent(s string) string {
	return `"` + strings.ReplaceAll(s, `"`, `""`) + `"`
}

func quoteQualifiedIdent(schemaName, objectName string) string {
	return quoteIdent(schemaName) + "." + quoteIdent(objectName)
}

func quoteLiteral(s string) string {
	return `'` + strings.ReplaceAll(s, `'`, `''`) + `'`
}

func tableKey(schemaName, tableName string) string {
	return schemaName + "." + tableName
}

func tableHasColumn(table Table, columnName string) bool {
	for _, col := range table.Columns {
		if col.Name == columnName {
			return true
		}
	}

	return false
}
