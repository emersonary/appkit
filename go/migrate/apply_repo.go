package migrate

import (
	"context"
	"database/sql"
	"fmt"
)

func applyRepoFunctions(ctx context.Context, db *sql.DB, tables []Table) (created, reused int, err error) {
	if _, err := db.ExecContext(ctx, buildRepoRegistrySQL()); err != nil {
		return 0, 0, fmt.Errorf("create repo registry: %w", err)
	}

	ordered, err := sortTablesChildrenFirst(tables)
	if err != nil {
		return 0, 0, err
	}

	tableMap := indexTables(tables)
	names := repoNames{}
	resolved := map[funcKey]resolvedFunc{}

	for _, table := range ordered {
		for _, operation := range repoOperations() {
			snap, err := buildRepoSnapshot(table, operation, tableMap, resolved)
			if err != nil {
				return created, reused, fmt.Errorf("snapshot %s.%s.%s: %w", table.Schema, table.Name, operation, err)
			}

			hash, _, err := hashRepoSnapshot(snap)
			if err != nil {
				return created, reused, fmt.Errorf("hash %s.%s.%s: %w", table.Schema, table.Name, operation, err)
			}

			key := makeFuncKey(table.Schema, table.Name, operation)
			existing, found, err := lookupRepoFunctionByHash(ctx, db, table.Schema, table.Name, operation, hash)
			if err != nil {
				return created, reused, fmt.Errorf("lookup repo function: %w", err)
			}

			var res resolvedFunc
			if found {
				res = existing
				reused++
			} else {
				version, err := nextRepoFunctionVersion(ctx, db, table.Schema, table.Name, operation)
				if err != nil {
					return created, reused, fmt.Errorf("next version: %w", err)
				}
				res = resolvedFunc{
					Version: version,
					Name:    versionedRepoFunctionName(table.Name, operation, version),
					Hash:    hash,
				}
				created++
			}

			names[key] = res.Name
			resolved[key] = res

			definition := buildRepoFunctionSQL(table, operation, tableMap, names)
			if _, err := db.ExecContext(ctx, definition); err != nil {
				return created, reused, fmt.Errorf("create %s.%s: %w", table.Schema, res.Name, err)
			}

			audit, repo := registryFlags(table, operation)
			if found {
				if err := updateRepoFunctionFlags(ctx, db, table.Schema, table.Name, operation, res, audit, repo); err != nil {
					return created, reused, err
				}
			} else {
				if err := insertRepoFunction(ctx, db, table.Schema, table.Name, operation, res, definition, audit, repo); err != nil {
					return created, reused, err
				}
			}
		}
	}

	return created, reused, nil
}

// registryFlags returns boo_audit / boo_repo for a registry row.
// List is only "repo-active" when the table has child relations (nested list payoff).
func registryFlags(table Table, operation string) (audit, repo bool) {
	audit = table.Audit
	repo = table.Repo
	if operation == "list" && len(table.Children) == 0 {
		repo = false
	}
	return audit, repo
}

func buildRepoFunctionSQL(table Table, operation string, tableMap map[string]Table, names repoNames) string {
	switch operation {
	case "insert":
		return buildInsertFunctionSQL(table, names)
	case "update":
		return buildUpdateFunctionSQL(table, names)
	case "upsert":
		return buildUpsertFunctionSQL(table, names)
	case "delete":
		return buildDeleteFunctionSQL(table, tableMap, names)
	case "get":
		return buildGetFunctionSQL(table, tableMap, names)
	case "list":
		return buildListFunctionSQL(table, names)
	default:
		return ""
	}
}
