package accounts

import "github.com/emersonary/appkit/apperror"

var (
	ErrSchemaRequired = apperror.Error{
		Code:    "ACCOUNTS_SCHEMA_REQUIRED",
		Domain:  "accounts",
		Message: "schema must be set in accounts configuration",
		Kind:    apperror.KindValidation,
	}
	ErrInvalidSchema = apperror.Error{
		Code:    "ACCOUNTS_INVALID_SCHEMA",
		Domain:  "accounts",
		Message: "schema name is invalid",
		Kind:    apperror.KindValidation,
	}
	ErrDefaultTenantRequired = apperror.Error{
		Code:    "ACCOUNTS_DEFAULT_TENANT_REQUIRED",
		Domain:  "accounts",
		Message: "tenancy.default_tenant_id is required when tenancy is disabled",
		Kind:    apperror.KindValidation,
	}
	ErrApplySchema = apperror.Error{
		Code:    "ACCOUNTS_APPLY_SCHEMA",
		Domain:  "accounts",
		Message: "failed to apply accounts schema",
		Kind:    apperror.KindInternal,
	}
	ErrLoadConfig = apperror.Error{
		Code:    "ACCOUNTS_LOAD_CONFIG",
		Domain:  "accounts",
		Message: "failed to load accounts configuration",
		Kind:    apperror.KindInternal,
	}
	ErrNotFound = apperror.Error{
		Code:    "ACCOUNTS_NOT_FOUND",
		Domain:  "accounts",
		Message: "account not found",
		Kind:    apperror.KindValidation,
	}
	ErrUnauthenticated = apperror.Error{
		Code:    "ACCOUNTS_UNAUTHENTICATED",
		Domain:  "accounts",
		Message: "invalid credentials or session",
		Kind:    apperror.KindValidation,
	}
	ErrAlreadyExists = apperror.Error{
		Code:    "ACCOUNTS_ALREADY_EXISTS",
		Domain:  "accounts",
		Message: "account already exists",
		Kind:    apperror.KindValidation,
	}
	ErrEmailNotVerified = apperror.Error{
		Code:    "ACCOUNTS_EMAIL_NOT_VERIFIED",
		Domain:  "accounts",
		Message: "email address is not verified",
		Kind:    apperror.KindValidation,
	}
	ErrInvalidArgument = apperror.Error{
		Code:    "ACCOUNTS_INVALID_ARGUMENT",
		Domain:  "accounts",
		Message: "invalid request argument",
		Kind:    apperror.KindValidation,
	}
	ErrInvalidToken = apperror.Error{
		Code:    "ACCOUNTS_INVALID_TOKEN",
		Domain:  "accounts",
		Message: "invalid or expired token",
		Kind:    apperror.KindValidation,
	}
	ErrOAuthUnavailable = apperror.Error{
		Code:    "ACCOUNTS_OAUTH_UNAVAILABLE",
		Domain:  "accounts",
		Message: "oauth provider is not configured",
		Kind:    apperror.KindValidation,
	}
)

func wrapErr(base apperror.Error, field string, err error) error {
	if err == nil {
		return nil
	}
	return base.With(field, err.Error())
}
