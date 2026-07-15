package migrate

import (
	"context"
	"database/sql"
	"fmt"
	"sync"

	"go.uber.org/zap"
)

// Service owns applied dbhist infrastructure for an application.
type Service struct {
	db     *sql.DB
	cfg    Config
	logger *zap.Logger

	mu     sync.RWMutex
	result Result
}

// Config returns the resolved dbhist config.
func (s *Service) Config() Config {
	return s.cfg
}

// LastResult returns the most recent Apply result.
func (s *Service) LastResult() Result {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.result
}

// Apply runs UpdateHist (audit metadata, history, versioned repo functions).
func (s *Service) Apply(ctx context.Context) (Result, error) {
	result, err := UpdateHist(ctx, s.db, s.cfg, Options{Logger: s.logger})
	if err != nil {
		return result, err
	}
	s.mu.Lock()
	s.result = result
	s.mu.Unlock()
	return result, nil
}

// WireOptions configures optional dbhist wiring behavior.
type WireOptions struct {
	Logger *zap.Logger
	// WorkerCtx runs the LastExecuted sync loop until cancelled. Optional
	// (migrate CLI omits it); API runtime should pass a RegisterWorker context.
	WorkerCtx context.Context
}

// Wire resolves config and applies audit/history/repo functions when app.Enabled is true.
// Returns nil, nil when disabled. Does not close db (caller owns the pool-backed handle).
func Wire(ctx context.Context, db *sql.DB, app AppConfig, opts WireOptions) (*Service, error) {
	if !app.Enabled {
		return nil, nil
	}
	if db == nil {
		return nil, fmt.Errorf("migrate: database is required when enabled")
	}

	cfg, err := ResolveBlockConfig(app)
	if err != nil {
		return nil, fmt.Errorf("resolve dbhist config: %w", err)
	}

	logger := opts.Logger
	if logger == nil {
		logger = zap.NewNop()
	}

	svc := &Service{
		db:     db,
		cfg:    cfg,
		logger: logger,
	}

	if _, err := svc.Apply(ctx); err != nil {
		return nil, err
	}

	result := svc.LastResult()
	svc.logger.Info("migrate audit/repo enabled",
		zap.Int("tables_found", result.TablesFound),
		zap.Int("tables_skipped", len(result.SkippedTables)),
		zap.Bool("audit_applied", result.AuditApplied),
		zap.Bool("history_applied", result.HistoryApplied),
		zap.Bool("repo_applied", result.RepoApplied),
		zap.Int("repo_functions_created", result.RepoFunctionsCreated),
		zap.Int("repo_functions_unchanged", result.RepoFunctionsUnchanged),
		zap.Int("repo_function_tables", len(SnapshotFunctions())),
	)

	svc.startLastExecutedSync(opts.WorkerCtx)
	return svc, nil
}
