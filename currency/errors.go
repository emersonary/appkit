package currency

import "github.com/emersonary/appkit/apperror"

var (
	ErrSchemaRequired = apperror.Error{
		Code:    "CURRENCY_SCHEMA_REQUIRED",
		Domain:  "currency",
		Message: "schema must be set in currency configuration",
		Kind:    apperror.KindValidation,
	}

	ErrInvalidSchema = apperror.Error{
		Code:    "CURRENCY_INVALID_SCHEMA",
		Domain:  "currency",
		Message: "schema name is invalid",
		Kind:    apperror.KindValidation,
	}

	ErrApplySchema = apperror.Error{
		Code:    "CURRENCY_APPLY_SCHEMA_FAILED",
		Domain:  "currency",
		Message: "failed to apply currency schema",
		Kind:    apperror.KindInternal,
	}

	ErrFeedFetch = apperror.Error{
		Code:    "CURRENCY_FEED_FETCH_FAILED",
		Domain:  "currency",
		Message: "failed to fetch exchange rates",
		Kind:    apperror.KindInternal,
	}

	ErrSyncRates = apperror.Error{
		Code:    "CURRENCY_SYNC_RATES_FAILED",
		Domain:  "currency",
		Message: "failed to sync exchange rates",
		Kind:    apperror.KindInternal,
	}

	ErrRateNotFound = apperror.Error{
		Code:    "CURRENCY_RATE_NOT_FOUND",
		Domain:  "currency",
		Message: "exchange rate not found",
		Kind:    apperror.KindValidation,
	}

	ErrInvalidAmount = apperror.Error{
		Code:    "CURRENCY_INVALID_AMOUNT",
		Domain:  "currency",
		Message: "amount must be greater than zero",
		Kind:    apperror.KindValidation,
	}

	ErrSameCurrency = apperror.Error{
		Code:    "CURRENCY_SAME_CURRENCY",
		Domain:  "currency",
		Message: "from and to currency must differ",
		Kind:    apperror.KindValidation,
	}

	ErrUnknownCurrency = apperror.Error{
		Code:    "CURRENCY_NOT_ENABLED",
		Domain:  "currency",
		Message: "currency is not enabled in configuration",
		Kind:    apperror.KindValidation,
	}

	ErrUnknownISO4217 = apperror.Error{
		Code:    "CURRENCY_UNKNOWN_ISO4217",
		Domain:  "currency",
		Message: "currency code is not a known ISO 4217 code",
		Kind:    apperror.KindValidation,
	}

	ErrLoadConfig = apperror.Error{
		Code:    "CURRENCY_LOAD_CONFIG_FAILED",
		Domain:  "currency",
		Message: "failed to load currency config",
		Kind:    apperror.KindInternal,
	}

	ErrEmptyCurrencies = apperror.Error{
		Code:    "CURRENCY_LIST_EMPTY",
		Domain:  "currency",
		Message: "currencies must not be empty",
		Kind:    apperror.KindValidation,
	}

	ErrDuplicateCurrency = apperror.Error{
		Code:    "CURRENCY_DUPLICATE_CODE",
		Domain:  "currency",
		Message: "duplicate currency code in configuration",
		Kind:    apperror.KindValidation,
	}

	ErrInvalidBaseCurrency = apperror.Error{
		Code:    "CURRENCY_INVALID_BASE",
		Domain:  "currency",
		Message: "base currency must be USD for USD-based exchange feeds",
		Kind:    apperror.KindValidation,
	}

	ErrBaseCurrencyMissing = apperror.Error{
		Code:    "CURRENCY_BASE_MISSING",
		Domain:  "currency",
		Message: "base currency must appear in currencies list",
		Kind:    apperror.KindValidation,
	}
)

func wrapErr(base apperror.Error, field string, err error) error {
	if err == nil {
		return nil
	}

	return base.With(field, err.Error())
}
