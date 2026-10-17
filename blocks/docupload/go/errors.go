package docupload

import "github.com/emersonary/appkit/apperror"

var (
	ErrInvalidArgument = apperror.Error{
		Code: "DOCUPLOAD_INVALID_ARGUMENT", Domain: "docupload",
		Message: "invalid request argument", Kind: apperror.KindValidation,
	}
	ErrEmptyFile = apperror.Error{
		Code: "DOCUPLOAD_EMPTY_FILE", Domain: "docupload",
		Message: "file is empty", Kind: apperror.KindValidation,
	}
	ErrTooLarge = apperror.Error{
		Code: "DOCUPLOAD_TOO_LARGE", Domain: "docupload",
		Message: "file exceeds the upload size limit", Kind: apperror.KindValidation,
	}
	ErrNotFound = apperror.Error{
		Code: "DOCUPLOAD_NOT_FOUND", Domain: "docupload",
		Message: "upload session not found", Kind: apperror.KindValidation,
	}
	ErrExpired = apperror.Error{
		Code: "DOCUPLOAD_EXPIRED", Domain: "docupload",
		Message: "upload session has expired", Kind: apperror.KindValidation,
	}
	ErrIncomplete = apperror.Error{
		Code: "DOCUPLOAD_INCOMPLETE", Domain: "docupload",
		Message: "upload is missing blocks", Kind: apperror.KindValidation,
	}
	ErrChecksum = apperror.Error{
		Code: "DOCUPLOAD_CHECKSUM", Domain: "docupload",
		Message: "uploaded bytes do not match the declared digest", Kind: apperror.KindValidation,
	}
	ErrConflict = apperror.Error{
		Code: "DOCUPLOAD_CONFLICT", Domain: "docupload",
		Message: "block index already holds different bytes", Kind: apperror.KindValidation,
	}
	ErrUnauthenticated = apperror.Error{
		Code: "DOCUPLOAD_UNAUTHENTICATED", Domain: "docupload",
		Message: "authentication required", Kind: apperror.KindValidation,
	}
	ErrInternal = apperror.Error{
		Code: "DOCUPLOAD_INTERNAL", Domain: "docupload",
		Message: "upload store failed", Kind: apperror.KindInternal,
	}
)

func wrapErr(base apperror.Error, field string, err error) error {
	if err == nil {
		return nil
	}
	return base.With(field, err.Error())
}
