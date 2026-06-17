package tenants

import "github.com/emersonary/appkit/apperror"

var (
	ErrSchemaRequired = apperror.Error{
		Code: "TENANTS_SCHEMA_REQUIRED", Domain: "tenants",
		Message: "schema must be set in tenants configuration", Kind: apperror.KindValidation,
	}
	ErrInvalidSchema = apperror.Error{
		Code: "TENANTS_INVALID_SCHEMA", Domain: "tenants",
		Message: "schema name is invalid", Kind: apperror.KindValidation,
	}
	ErrApplySchema = apperror.Error{
		Code: "TENANTS_APPLY_SCHEMA", Domain: "tenants",
		Message: "failed to apply tenants schema", Kind: apperror.KindInternal,
	}
	ErrLoadConfig = apperror.Error{
		Code: "TENANTS_LOAD_CONFIG", Domain: "tenants",
		Message: "failed to load tenants configuration", Kind: apperror.KindInternal,
	}
	ErrNotFound = apperror.Error{
		Code: "TENANTS_NOT_FOUND", Domain: "tenants",
		Message: "tenant not found", Kind: apperror.KindValidation,
	}
	ErrForbidden = apperror.Error{
		Code: "TENANTS_FORBIDDEN", Domain: "tenants",
		Message: "forbidden", Kind: apperror.KindValidation,
	}
	ErrUnauthenticated = apperror.Error{
		Code: "TENANTS_UNAUTHENTICATED", Domain: "tenants",
		Message: "authentication required", Kind: apperror.KindValidation,
	}
	ErrAlreadyExists = apperror.Error{
		Code: "TENANTS_ALREADY_EXISTS", Domain: "tenants",
		Message: "tenant already exists", Kind: apperror.KindValidation,
	}
	ErrInvalidArgument = apperror.Error{
		Code: "TENANTS_INVALID_ARGUMENT", Domain: "tenants",
		Message: "invalid request argument", Kind: apperror.KindValidation,
	}
	ErrInvalidToken = apperror.Error{
		Code: "TENANTS_INVALID_TOKEN", Domain: "tenants",
		Message: "invalid or expired invite token", Kind: apperror.KindValidation,
	}
)

func wrapErr(base apperror.Error, field string, err error) error {
	if err == nil {
		return nil
	}
	return base.With(field, err.Error())
}
