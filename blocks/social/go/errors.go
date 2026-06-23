package social

import (
	"fmt"

	"github.com/emersonary/appkit/apperror"
)

var (
	ErrLoadConfig = apperror.Error{
		Code:    "SOCIAL_LOAD_CONFIG_FAILED",
		Domain:  "social",
		Message: "failed to load social config",
		Kind:    apperror.KindInternal,
	}

	ErrInvalidConfig = apperror.Error{
		Code:    "SOCIAL_INVALID_CONFIG",
		Domain:  "social",
		Message: "social configuration is invalid",
		Kind:    apperror.KindValidation,
	}

	ErrPlatformNotFound = apperror.Error{
		Code:    "SOCIAL_PLATFORM_NOT_FOUND",
		Domain:  "social",
		Message: "social platform not found",
		Kind:    apperror.KindValidation,
	}

	ErrPlatformDisabled = apperror.Error{
		Code:    "SOCIAL_PLATFORM_DISABLED",
		Domain:  "social",
		Message: "social platform is disabled",
		Kind:    apperror.KindValidation,
	}

	ErrInvalidRequest = apperror.Error{
		Code:    "SOCIAL_INVALID_REQUEST",
		Domain:  "social",
		Message: "social request is invalid",
		Kind:    apperror.KindValidation,
	}

	ErrPublishFailed = apperror.Error{
		Code:    "SOCIAL_PUBLISH_FAILED",
		Domain:  "social",
		Message: "social publish request failed",
		Kind:    apperror.KindInternal,
	}

	ErrAPIFailed = apperror.Error{
		Code:    "SOCIAL_API_FAILED",
		Domain:  "social",
		Message: "social platform API request failed",
		Kind:    apperror.KindInternal,
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
