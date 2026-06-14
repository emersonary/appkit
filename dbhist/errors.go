package dbhist

import "github.com/emersonary/appkit/apperror"

var (
	ErrSchemasEmpty = apperror.Error{
		Code:    "DBHIST_SCHEMAS_EMPTY",
		Domain:  "dbhist",
		Message: "schemas must not be empty",
		Kind:    apperror.KindValidation,
	}

	ErrInvalidSchema = apperror.Error{
		Code:    "DBHIST_INVALID_SCHEMA",
		Domain:  "dbhist",
		Message: "schema name is invalid",
		Kind:    apperror.KindValidation,
	}

	ErrInvalidTablePattern = apperror.Error{
		Code:    "DBHIST_INVALID_TABLE_PATTERN",
		Domain:  "dbhist",
		Message: "table pattern contains invalid characters",
		Kind:    apperror.KindValidation,
	}

	ErrInvalidExcludePattern = apperror.Error{
		Code:    "DBHIST_INVALID_EXCLUDE_PATTERN",
		Domain:  "dbhist",
		Message: "exclude pattern contains invalid characters",
		Kind:    apperror.KindValidation,
	}

	ErrLoadConfig = apperror.Error{
		Code:    "DBHIST_LOAD_CONFIG_FAILED",
		Domain:  "dbhist",
		Message: "failed to load dbhist config",
		Kind:    apperror.KindInternal,
	}

	ErrLoadTables = apperror.Error{
		Code:    "DBHIST_LOAD_TABLES_FAILED",
		Domain:  "dbhist",
		Message: "failed to load tables",
		Kind:    apperror.KindInternal,
	}

	ErrLoadForeignKeys = apperror.Error{
		Code:    "DBHIST_LOAD_FOREIGN_KEYS_FAILED",
		Domain:  "dbhist",
		Message: "failed to load foreign keys",
		Kind:    apperror.KindInternal,
	}

	ErrApplyAudit = apperror.Error{
		Code:    "DBHIST_APPLY_AUDIT_FAILED",
		Domain:  "dbhist",
		Message: "failed to apply audit metadata",
		Kind:    apperror.KindInternal,
	}

	ErrLoadAuditIDs = apperror.Error{
		Code:    "DBHIST_LOAD_AUDIT_IDS_FAILED",
		Domain:  "dbhist",
		Message: "failed to load audit ids",
		Kind:    apperror.KindInternal,
	}

	ErrApplyHistory = apperror.Error{
		Code:    "DBHIST_APPLY_HISTORY_FAILED",
		Domain:  "dbhist",
		Message: "failed to apply history infrastructure",
		Kind:    apperror.KindInternal,
	}

	ErrApplyRepo = apperror.Error{
		Code:    "DBHIST_APPLY_REPO_FAILED",
		Domain:  "dbhist",
		Message: "failed to apply repository functions",
		Kind:    apperror.KindInternal,
	}
)

func wrapErr(base apperror.Error, field string, err error) error {
	if err == nil {
		return nil
	}

	return base.With(field, err.Error())
}
