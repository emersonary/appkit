package migrate

import "go.uber.org/zap"

type Options struct {
	Logger *zap.Logger
}

type Result struct {
	TablesFound             int
	SkippedTables           []SkippedTable
	AuditApplied            bool
	HistoryApplied          bool
	RepoApplied             bool
	RepoFunctionsCreated    int
	RepoFunctionsUnchanged  int
}

type SkippedTable struct {
	Schema string
	Table  string
	Reason string
}
