// Package dbhist is a compatibility shim for github.com/emersonary/appkit/migrate.
//
// Schema instructions (migrate) and audit/repo (UpdateHist) now live in the same
// package: migrate. Prefer importing migrate directly.
package dbhist

import "github.com/emersonary/appkit/migrate"

type (
	AppConfig      = migrate.AppConfig
	Config         = migrate.Config
	Options        = migrate.Options
	Result         = migrate.Result
	Service        = migrate.Service
	TableFunctions = migrate.TableFunctions
	WireOptions    = migrate.WireOptions
	Table          = migrate.Table
	SkippedTable   = migrate.SkippedTable
)

var (
	Wire               = migrate.Wire
	UpdateHist         = migrate.UpdateHist
	LoadTableFunctions = migrate.LoadTableFunctions
	Functions          = migrate.Functions
	MustFunctions      = migrate.MustFunctions
	SnapshotFunctions  = migrate.SnapshotFunctions
	RepoEnabled        = migrate.RepoEnabled
	AuditEnabled       = migrate.AuditEnabled
	ResolveBlockConfig = migrate.ResolveBlockConfig
)
