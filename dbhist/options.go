package dbhist

import "go.uber.org/zap"

type Options struct {
	Logger *zap.Logger
}

type Result struct {
	TablesFound     int
	SkippedTables   []SkippedTable
	AuditApplied    bool
	HistoryApplied  bool
	RepoApplied     bool
}

type SkippedTable struct {
	Schema string
	Table  string
	Reason string
}
