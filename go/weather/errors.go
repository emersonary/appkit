package weather

import (
	"fmt"

	"github.com/emersonary/appkit/apperror"
)

var (
	ErrLoadConfig = apperror.Error{
		Code:    "WEATHER_LOAD_CONFIG_FAILED",
		Domain:  "weather",
		Message: "failed to load weather config",
		Kind:    apperror.KindInternal,
	}

	ErrInvalidConfig = apperror.Error{
		Code:    "WEATHER_INVALID_CONFIG",
		Domain:  "weather",
		Message: "weather configuration is invalid",
		Kind:    apperror.KindValidation,
	}

	ErrRedisRequired = apperror.Error{
		Code:    "WEATHER_REDIS_REQUIRED",
		Domain:  "weather",
		Message: "redis client is required when weather is enabled",
		Kind:    apperror.KindValidation,
	}

	ErrForecastNotFound = apperror.Error{
		Code:    "WEATHER_FORECAST_NOT_FOUND",
		Domain:  "weather",
		Message: "weather forecast not found",
		Kind:    apperror.KindValidation,
	}

	ErrFetchForecast = apperror.Error{
		Code:    "WEATHER_FETCH_FORECAST_FAILED",
		Domain:  "weather",
		Message: "failed to fetch weather forecast",
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
