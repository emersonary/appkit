package permissions

import (
	"database/sql"
	"errors"

	"github.com/emersonary/appkit/apperror"
)

var (
	ErrSchemaRequired = apperror.Error{
		Code:    "PERMISSIONS_SCHEMA_REQUIRED",
		Domain:  "permissions",
		Message: "schema must be set in permissions configuration",
		Kind:    apperror.KindValidation,
	}

	ErrInvalidSchema = apperror.Error{
		Code:    "PERMISSIONS_INVALID_SCHEMA",
		Domain:  "permissions",
		Message: "schema name is invalid",
		Kind:    apperror.KindValidation,
	}

	ErrApplySchema = apperror.Error{
		Code:    "PERMISSIONS_APPLY_SCHEMA_FAILED",
		Domain:  "permissions",
		Message: "failed to apply permissions schema",
		Kind:    apperror.KindInternal,
	}

	ErrLoadSetup = apperror.Error{
		Code:    "PERMISSIONS_LOAD_SETUP_FAILED",
		Domain:  "permissions",
		Message: "failed to load permissions setup",
		Kind:    apperror.KindInternal,
	}

	ErrDefaultProfileRequired = apperror.Error{
		Code:    "PERMISSIONS_DEFAULT_PROFILE_REQUIRED",
		Domain:  "permissions",
		Message: "default_profile must be set",
		Kind:    apperror.KindValidation,
	}

	ErrProfileNotFound = apperror.Error{
		Code:    "PERMISSIONS_PROFILE_NOT_FOUND",
		Domain:  "permissions",
		Message: "profile not found",
		Kind:    apperror.KindValidation,
	}

	ErrPermissionNotFound = apperror.Error{
		Code:    "PERMISSIONS_PERMISSION_NOT_FOUND",
		Domain:  "permissions",
		Message: "permission not found",
		Kind:    apperror.KindValidation,
	}

	ErrPermissionDenied = apperror.Error{
		Code:    "PERMISSIONS_DENIED",
		Domain:  "permissions",
		Message: "permission denied",
		Kind:    apperror.KindValidation,
	}

	ErrInvalidAction = apperror.Error{
		Code:    "PERMISSIONS_INVALID_ACTION",
		Domain:  "permissions",
		Message: "invalid action bit",
		Kind:    apperror.KindValidation,
	}
)

func wrapErr(base apperror.Error, field string, err error) error {
	if err == nil {
		return nil
	}
	return base.With(field, err.Error())
}

func IsNotFound(err error) bool {
	return errors.Is(err, sql.ErrNoRows)
}
