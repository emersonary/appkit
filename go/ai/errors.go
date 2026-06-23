package ai

import (
	"fmt"

	"github.com/emersonary/appkit/apperror"
)

var (
	ErrLoadConfig = apperror.Error{
		Code:    "AI_LOAD_CONFIG_FAILED",
		Domain:  "ai",
		Message: "failed to load ai config",
		Kind:    apperror.KindInternal,
	}

	ErrInvalidConfig = apperror.Error{
		Code:    "AI_INVALID_CONFIG",
		Domain:  "ai",
		Message: "ai configuration is invalid",
		Kind:    apperror.KindValidation,
	}

	ErrProviderNotFound = apperror.Error{
		Code:    "AI_PROVIDER_NOT_FOUND",
		Domain:  "ai",
		Message: "ai provider not found",
		Kind:    apperror.KindValidation,
	}

	ErrServiceNotRouted = apperror.Error{
		Code:    "AI_SERVICE_NOT_ROUTED",
		Domain:  "ai",
		Message: "ai service type is not routed to a provider",
		Kind:    apperror.KindValidation,
	}

	ErrNotImplemented = apperror.Error{
		Code:    "AI_NOT_IMPLEMENTED",
		Domain:  "ai",
		Message: "ai operation is not implemented for this provider",
		Kind:    apperror.KindValidation,
	}

	ErrTranslateFailed = apperror.Error{
		Code:    "AI_TRANSLATE_FAILED",
		Domain:  "ai",
		Message: "translation request failed",
		Kind:    apperror.KindInternal,
	}

	ErrInvalidRequest = apperror.Error{
		Code:    "AI_INVALID_REQUEST",
		Domain:  "ai",
		Message: "ai request is invalid",
		Kind:    apperror.KindValidation,
	}
)

func wrapErr(base apperror.Error, field string, err error) error {
	if err == nil {
		return nil
	}
	return base.With(field, err.Error())
}

func invalidConfig(field, detail string) error {
	return ErrInvalidConfig.With(field, detail)
}

func invalidConfigf(field, format string, args ...any) error {
	return invalidConfig(field, fmt.Sprintf(format, args...))
}
