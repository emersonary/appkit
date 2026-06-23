package menu

import (
	"fmt"

	"github.com/emersonary/appkit/apperror"
)

var (
	ErrLoadSetup = apperror.Error{
		Code:    "MENU_LOAD_SETUP_FAILED",
		Domain:  "menu",
		Message: "failed to load menu setup",
		Kind:    apperror.KindInternal,
	}

	ErrInvalidSetup = apperror.Error{
		Code:    "MENU_INVALID_SETUP",
		Domain:  "menu",
		Message: "menu setup is invalid",
		Kind:    apperror.KindValidation,
	}

	ErrUnauthenticated = apperror.Error{
		Code:    "MENU_UNAUTHENTICATED",
		Domain:  "menu",
		Message: "authentication required",
		Kind:    apperror.KindValidation,
	}

	ErrPermissionsRequired = apperror.Error{
		Code:    "MENU_PERMISSIONS_REQUIRED",
		Domain:  "menu",
		Message: "permissions service is required",
		Kind:    apperror.KindValidation,
	}
)

func wrapErr(base apperror.Error, field string, err error) error {
	if err == nil {
		return nil
	}
	return base.With(field, err.Error())
}

func invalidSetup(field, detail string) error {
	return ErrInvalidSetup.With(field, detail)
}

func invalidSetupf(field, format string, args ...any) error {
	return invalidSetup(field, fmt.Sprintf(format, args...))
}
