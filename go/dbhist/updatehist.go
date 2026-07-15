package dbhist

import (
	"context"
	"database/sql"

	"go.uber.org/zap"
)

func UpdateHist(ctx context.Context, db *sql.DB, cfg Config, opts Options) (Result, error) {
	if err := cfg.Validate(); err != nil {
		return Result{}, err
	}

	result := Result{}

	tables, skipped, err := loadTables(ctx, db, cfg)
	if err != nil {
		return Result{}, wrapErr(ErrLoadTables, "query", err)
	}

	result.SkippedTables = skipped

	foreignKeys, err := loadForeignKeys(ctx, db, cfg)
	if err != nil {
		return Result{}, wrapErr(ErrLoadForeignKeys, "query", err)
	}

	tables = attachChildRelations(tables, foreignKeys)
	result.TablesFound = len(tables)

	logSkippedTables(opts.Logger, result.SkippedTables)

	auditTables := filterTables(tables, func(t Table) bool { return t.Audit })
	if len(auditTables) > 0 {
		auditTables = filterChildren(auditTables, func(t Table) bool { return t.Audit })

		if _, err := db.ExecContext(ctx, buildAuditMetadataSQL(auditTables)); err != nil {
			return result, wrapErr(ErrApplyAudit, "exec", err)
		}
		result.AuditApplied = true

		auditTables, err = loadAuditIDs(ctx, db, auditTables)
		if err != nil {
			return result, wrapErr(ErrLoadAuditIDs, "query", err)
		}

		if _, err := db.ExecContext(ctx, buildHistSQL(auditTables)); err != nil {
			return result, wrapErr(ErrApplyHistory, "exec", err)
		}
		result.HistoryApplied = true
	}

	repoTables := filterTables(tables, func(t Table) bool { return t.Repo })
	if len(repoTables) > 0 {
		repoTables = filterChildren(repoTables, func(t Table) bool { return t.Repo })

		created, reused, err := applyRepoFunctions(ctx, db, repoTables)
		if err != nil {
			return result, wrapErr(ErrApplyRepo, "exec", err)
		}

		result.RepoApplied = true
		result.RepoFunctionsCreated = created
		result.RepoFunctionsUnchanged = reused
	}

	return result, nil
}

func filterTables(tables []Table, keep func(Table) bool) []Table {
	out := make([]Table, 0, len(tables))
	index := make(map[string]Table, len(tables))
	for _, t := range tables {
		index[tableKey(t.Schema, t.Name)] = t
	}
	for _, t := range tables {
		if !keep(t) {
			continue
		}
		// Re-resolve from index so filtered sets still carry full Table metadata.
		out = append(out, index[tableKey(t.Schema, t.Name)])
	}
	return out
}

// filterChildren keeps only child relations whose target table satisfies keep.
func filterChildren(tables []Table, keep func(Table) bool) []Table {
	index := make(map[string]Table, len(tables))
	for _, t := range tables {
		index[tableKey(t.Schema, t.Name)] = t
	}

	out := make([]Table, len(tables))
	copy(out, tables)
	for i := range out {
		children := make([]ChildRelation, 0, len(out[i].Children))
		for _, child := range out[i].Children {
			target, ok := index[tableKey(child.Schema, child.Table)]
			if !ok || !keep(target) {
				continue
			}
			children = append(children, child)
		}
		out[i].Children = children
	}
	return out
}

func logSkippedTables(logger *zap.Logger, skipped []SkippedTable) {
	if logger == nil {
		return
	}

	for _, item := range skipped {
		logger.Warn("table skipped",
			zap.String("schema", item.Schema),
			zap.String("table", item.Table),
			zap.String("reason", item.Reason),
		)
	}
}
