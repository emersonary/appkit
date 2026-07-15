package dbhist

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
)

type funcKey struct {
	Schema    string
	Table     string
	Operation string
}

type resolvedFunc struct {
	Version int
	Name    string
	Hash    string
}

type repoCallRef struct {
	Schema    string `json:"schema"`
	Table     string `json:"table"`
	Operation string `json:"operation"`
	Hash      string `json:"definition_hash"`
}

type repoColumnSnap struct {
	Name     string `json:"name"`
	UdtName  string `json:"udt"`
	Nullable bool   `json:"nullable"`
	Default  string `json:"default,omitempty"`
}

type repoChildSnap struct {
	Schema     string `json:"schema"`
	Table      string `json:"table"`
	PrimaryKey string `json:"pk"`
	FKColumn   string `json:"fk"`
	JSONKey    string `json:"json_key"`
	OneToOne   bool   `json:"one_to_one"`
}

type repoSnapshot struct {
	Generator string           `json:"generator"`
	Schema    string           `json:"schema"`
	Table     string           `json:"table"`
	Operation string           `json:"operation"`
	PK        repoColumnSnap   `json:"pk"`
	Columns   []repoColumnSnap `json:"columns"`
	RowStatus string           `json:"row_status,omitempty"`
	Children  []repoChildSnap  `json:"children,omitempty"`
	Calls     []repoCallRef    `json:"calls,omitempty"`
}

func makeFuncKey(schema, table, operation string) funcKey {
	return funcKey{Schema: schema, Table: table, Operation: operation}
}

func versionedRepoFunctionName(tableName, operation string, version int) string {
	return fmt.Sprintf("func_%s_%s_v%0*d", operation, tableName, repoVersionWidth, version)
}

func hashRepoSnapshot(snap repoSnapshot) (string, []byte, error) {
	body, err := json.Marshal(snap)
	if err != nil {
		return "", nil, err
	}
	sum := sha256.Sum256(body)
	return hex.EncodeToString(sum[:]), body, nil
}

func buildRepoSnapshot(table Table, operation string, tableMap map[string]Table, resolved map[funcKey]resolvedFunc) (repoSnapshot, error) {
	pkCol := columnByName(table, table.PrimaryKey)
	snap := repoSnapshot{
		Generator: repoGeneratorVersion,
		Schema:    table.Schema,
		Table:     table.Name,
		Operation: operation,
		PK: repoColumnSnap{
			Name:     table.PrimaryKey,
			UdtName:  pkCol.UdtName,
			Nullable: pkCol.IsNullable,
			Default:  pkCol.DefaultValue,
		},
		RowStatus: findRowStatusColumn(table),
	}

	for _, col := range table.Columns {
		if col.Name == table.PrimaryKey {
			continue
		}
		snap.Columns = append(snap.Columns, repoColumnSnap{
			Name:     col.Name,
			UdtName:  col.UdtName,
			Nullable: col.IsNullable,
			Default:  col.DefaultValue,
		})
	}

	for _, child := range table.Children {
		snap.Children = append(snap.Children, repoChildSnap{
			Schema:     child.Schema,
			Table:      child.Table,
			PrimaryKey: child.PrimaryKey,
			FKColumn:   child.FKColumn,
			JSONKey:    child.JSONKey,
			OneToOne:   child.IsOneToOne,
		})
	}

	calls, err := repoCallsForOperation(table, operation, tableMap, resolved)
	if err != nil {
		return repoSnapshot{}, err
	}
	snap.Calls = calls
	return snap, nil
}

func repoCallsForOperation(table Table, operation string, tableMap map[string]Table, resolved map[funcKey]resolvedFunc) ([]repoCallRef, error) {
	var refs []repoCallRef

	add := func(schema, name, op string) error {
		key := makeFuncKey(schema, name, op)
		res, ok := resolved[key]
		if !ok {
			return fmt.Errorf("missing resolved dependency %s.%s.%s", schema, name, op)
		}
		refs = append(refs, repoCallRef{
			Schema:    schema,
			Table:     name,
			Operation: op,
			Hash:      res.Hash,
		})
		return nil
	}

	switch operation {
	case "insert":
		for _, child := range table.Children {
			if err := add(child.Schema, child.Table, "insert"); err != nil {
				return nil, err
			}
		}
	case "update", "upsert":
		// update/upsert nest child upserts; upsert also depends on own insert/update.
		if operation == "upsert" {
			if err := add(table.Schema, table.Name, "insert"); err != nil {
				return nil, err
			}
			if err := add(table.Schema, table.Name, "update"); err != nil {
				return nil, err
			}
		}
		for _, child := range table.Children {
			if err := add(child.Schema, child.Table, "upsert"); err != nil {
				return nil, err
			}
		}
	case "delete":
		for _, child := range table.Children {
			if _, ok := tableMap[tableKey(child.Schema, child.Table)]; !ok {
				continue
			}
			if err := add(child.Schema, child.Table, "delete"); err != nil {
				return nil, err
			}
		}
	case "get":
		for _, child := range table.Children {
			if _, ok := tableMap[tableKey(child.Schema, child.Table)]; !ok {
				continue
			}
			if err := add(child.Schema, child.Table, "get"); err != nil {
				return nil, err
			}
		}
	}

	sort.Slice(refs, func(i, j int) bool {
		if refs[i].Schema != refs[j].Schema {
			return refs[i].Schema < refs[j].Schema
		}
		if refs[i].Table != refs[j].Table {
			return refs[i].Table < refs[j].Table
		}
		return refs[i].Operation < refs[j].Operation
	})
	return refs, nil
}

func columnByName(table Table, name string) Column {
	for _, col := range table.Columns {
		if col.Name == name {
			return col
		}
	}
	return Column{Name: name}
}

func repoOperations() []string {
	return []string{"insert", "update", "upsert", "delete", "get"}
}
