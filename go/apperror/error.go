package apperror

import (
	"errors"
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Kind string

const (
	KindValidation Kind = "validation"
	KindInternal   Kind = "internal"
)

type Error struct {
	Code    string
	Domain  string
	Field   string
	Message string
	Detail  string
	Kind    Kind
}

func (e Error) Error() string {
	if e.Detail != "" {
		return fmt.Sprintf(
			"%s: %s.%s: %s (%s)",
			e.Code,
			e.Domain,
			e.Field,
			e.Message,
			e.Detail,
		)
	}

	if e.Field != "" {
		return fmt.Sprintf(
			"%s: %s.%s: %s",
			e.Code,
			e.Domain,
			e.Field,
			e.Message,
		)
	}

	if e.Domain != "" {
		return fmt.Sprintf(
			"%s: %s: %s",
			e.Code,
			e.Domain,
			e.Message,
		)
	}

	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e Error) With(field string, detail ...string) error {
	err := Error{
		Code:    e.Code,
		Domain:  e.Domain,
		Field:   field,
		Message: e.Message,
		Kind:    e.Kind,
	}

	if len(detail) > 0 {
		err.Detail = detail[0]
	}

	return err
}

func (e Error) ZapFields() []zap.Field {
	fields := []zap.Field{
		zap.String("error.code", e.Code),
		zap.String("error.kind", string(e.effectiveKind())),
	}

	if e.Domain != "" {
		fields = append(fields, zap.String("error.domain", e.Domain))
	}

	if e.Field != "" {
		fields = append(fields, zap.String("error.field", e.Field))
	}

	if e.Message != "" {
		fields = append(fields, zap.String("error.message", e.Message))
	}

	if e.Detail != "" {
		fields = append(fields, zap.String("error.detail", e.Detail))
	}

	return fields
}

func (e Error) LogLevel() zapcore.Level {
	switch e.effectiveKind() {
	case KindValidation:
		return zapcore.WarnLevel
	default:
		return zapcore.ErrorLevel
	}
}

func (e Error) GRPCCode() codes.Code {
	switch e.effectiveKind() {
	case KindValidation:
		return codes.InvalidArgument
	default:
		return codes.Internal
	}
}

func (e Error) effectiveKind() Kind {
	if e.Kind != "" {
		return e.Kind
	}

	return KindValidation
}

func As(err error) (Error, bool) {
	var appErr Error
	ok := errors.As(err, &appErr)
	return appErr, ok
}

func IsValidation(err error) bool {
	appErr, ok := As(err)
	if !ok {
		return false
	}

	return appErr.effectiveKind() == KindValidation
}

func ToGRPCStatus(err error) error {
	if err == nil {
		return nil
	}

	appErr, ok := As(err)
	if ok {
		return status.Error(appErr.GRPCCode(), appErr.Error())
	}

	return status.Error(codes.Internal, err.Error())
}
