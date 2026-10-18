package messenger

import (
	"errors"
	"fmt"

	"github.com/emersonary/appkit/apperror"
)

var (
	ErrNotFound = apperror.Error{
		Code: "MESSENGER_NOT_FOUND", Domain: "messenger",
		Message: "resource not found", Kind: apperror.KindValidation,
	}
	ErrForbidden = apperror.Error{
		Code: "MESSENGER_FORBIDDEN", Domain: "messenger",
		Message: "forbidden", Kind: apperror.KindValidation,
	}
	ErrUnauthenticated = apperror.Error{
		Code: "MESSENGER_UNAUTH", Domain: "messenger",
		Message: "authentication required", Kind: apperror.KindValidation,
	}
	ErrInvalidInput = apperror.Error{
		Code: "MESSENGER_INVALID", Domain: "messenger",
		Message: "invalid input", Kind: apperror.KindValidation,
	}
	ErrApplySchema = apperror.Error{
		Code: "MESSENGER_APPLY_SCHEMA", Domain: "messenger",
		Message: "failed to apply messenger schema", Kind: apperror.KindInternal,
	}
	ErrConflict = apperror.Error{
		Code: "MESSENGER_CONFLICT", Domain: "messenger",
		Message: "conflict", Kind: apperror.KindValidation,
	}
	ErrSystemMessageEdit = apperror.Error{
		Code: "MESSENGER_SYSTEM_EDIT", Domain: "messenger",
		Message: "system messages cannot be edited", Kind: apperror.KindValidation,
	}
)

func wrapErr(base apperror.Error, step string, err error) error {
	if err == nil {
		return nil
	}
	return base.With("step", step, "cause", err.Error())
}

func isNotFound(err error) bool {
	var appErr apperror.Error
	return errors.As(err, &appErr) && appErr.Code == ErrNotFound.Code
}

func quoteIdent(ident string) string {
	return `"` + ident + `"`
}

func qualifiedName(schema, table string) string {
	return quoteIdent(schema) + "." + quoteIdent(table)
}

func validateIdent(ident string) error {
	if ident == "" {
		return fmt.Errorf("empty identifier")
	}
	for _, r := range ident {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' {
			continue
		}
		return fmt.Errorf("invalid identifier %q", ident)
	}
	return nil
}
