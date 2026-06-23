package language

import "github.com/emersonary/appkit/apperror"

var (
	ErrSchemaRequired = apperror.Error{
		Code:    "LANGUAGE_SCHEMA_REQUIRED",
		Domain:  "language",
		Message: "schema must be set in language configuration",
		Kind:    apperror.KindValidation,
	}

	ErrInvalidSchema = apperror.Error{
		Code:    "LANGUAGE_INVALID_SCHEMA",
		Domain:  "language",
		Message: "schema name is invalid",
		Kind:    apperror.KindValidation,
	}

	ErrApplySchema = apperror.Error{
		Code:    "LANGUAGE_APPLY_SCHEMA_FAILED",
		Domain:  "language",
		Message: "failed to apply language schema",
		Kind:    apperror.KindInternal,
	}

	ErrUnknownLanguage = apperror.Error{
		Code:    "LANGUAGE_NOT_ENABLED",
		Domain:  "language",
		Message: "language is not enabled in configuration",
		Kind:    apperror.KindValidation,
	}

	ErrUnknownCatalogCode = apperror.Error{
		Code:    "LANGUAGE_UNKNOWN_CODE",
		Domain:  "language",
		Message: "language code is not in the catalog",
		Kind:    apperror.KindValidation,
	}

	ErrLoadConfig = apperror.Error{
		Code:    "LANGUAGE_LOAD_CONFIG_FAILED",
		Domain:  "language",
		Message: "failed to load language config",
		Kind:    apperror.KindInternal,
	}

	ErrEmptyLanguages = apperror.Error{
		Code:    "LANGUAGE_LIST_EMPTY",
		Domain:  "language",
		Message: "languages must not be empty",
		Kind:    apperror.KindValidation,
	}

	ErrDuplicateLanguage = apperror.Error{
		Code:    "LANGUAGE_DUPLICATE_CODE",
		Domain:  "language",
		Message: "duplicate language code in configuration",
		Kind:    apperror.KindValidation,
	}

	ErrDefaultLanguageMissing = apperror.Error{
		Code:    "LANGUAGE_DEFAULT_MISSING",
		Domain:  "language",
		Message: "default language must appear in languages list",
		Kind:    apperror.KindValidation,
	}

	ErrDefaultLanguageRequired = apperror.Error{
		Code:    "LANGUAGE_DEFAULT_REQUIRED",
		Domain:  "language",
		Message: "default_language must be set",
		Kind:    apperror.KindValidation,
	}

	ErrLanguageNotFound = apperror.Error{
		Code:    "LANGUAGE_NOT_FOUND",
		Domain:  "language",
		Message: "language not found",
		Kind:    apperror.KindValidation,
	}
)

func wrapErr(base apperror.Error, field string, err error) error {
	if err == nil {
		return nil
	}

	return base.With(field, err.Error())
}
