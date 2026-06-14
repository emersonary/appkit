package dbhist

import (
	"context"
	"database/sql"

	"go.uber.org/zap"
)

func UpdateHist(ctx context.Context, db *sql.DB, cfg Config, opts Options) (Result, error) {
	cfg.applyDefaults()
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

	needsAuditMetadata := cfg.Modules.Audit || cfg.Modules.History
	if needsAuditMetadata {
		if _, err := db.ExecContext(ctx, buildAuditMetadataSQL(tables)); err != nil {
			return result, wrapErr(ErrApplyAudit, "exec", err)
		}

		result.AuditApplied = true
	}

	if cfg.Modules.History {
		tables, err = loadAuditIDs(ctx, db, tables)
		if err != nil {
			return result, wrapErr(ErrLoadAuditIDs, "query", err)
		}

		if _, err := db.ExecContext(ctx, buildHistSQL(tables)); err != nil {
			return result, wrapErr(ErrApplyHistory, "exec", err)
		}

		result.HistoryApplied = true
	}

	if cfg.Modules.RepoFunctions {
		if _, err := db.ExecContext(ctx, buildRepoFunctionsSQL(tables)); err != nil {
			return result, wrapErr(ErrApplyRepo, "exec", err)
		}

		result.RepoApplied = true
	}

	return result, nil
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
