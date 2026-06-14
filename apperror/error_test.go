package apperror

import (
	"errors"
	"testing"

	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestErrorWithPreservesKind(t *testing.T) {
	base := Error{
		Code:    "TEST",
		Domain:  "trip",
		Message: "test message",
		Kind:    KindValidation,
	}

	err := base.With("field_name", "extra detail")

	var appErr Error
	if !errors.As(err, &appErr) {
		t.Fatal("expected apperror.Error")
	}

	if appErr.Kind != KindValidation {
		t.Fatalf("expected kind %s, got %s", KindValidation, appErr.Kind)
	}

	if appErr.Field != "field_name" {
		t.Fatalf("expected field field_name, got %s", appErr.Field)
	}
}

func TestZapFields(t *testing.T) {
	err := Error{
		Code:    "TRIP_TYPE_REQUIRED",
		Domain:  "trip",
		Field:   "trip_type",
		Message: "trip type is required",
		Detail:  "received value: FOO",
		Kind:    KindValidation,
	}

	fields := err.ZapFields()
	if len(fields) != 6 {
		t.Fatalf("expected 6 zap fields, got %d", len(fields))
	}
}

func TestToGRPCStatus(t *testing.T) {
	validationErr := Error{
		Code:    "TRIP_TYPE_REQUIRED",
		Domain:  "trip",
		Message: "trip type is required",
		Kind:    KindValidation,
	}.With("trip_type")

	st := ToGRPCStatus(validationErr)
	if status.Code(st) != codes.InvalidArgument {
		t.Fatalf("expected InvalidArgument, got %v", status.Code(st))
	}

	internalErr := Error{
		Code:    "PUBLISH_FAILED",
		Domain:  "trip",
		Message: "failed to publish event",
		Kind:    KindInternal,
	}

	st = ToGRPCStatus(internalErr)
	if status.Code(st) != codes.Internal {
		t.Fatalf("expected Internal, got %v", status.Code(st))
	}

	st = ToGRPCStatus(errors.New("plain error"))
	if status.Code(st) != codes.Internal {
		t.Fatalf("expected Internal for plain error, got %v", status.Code(st))
	}
}

func TestIsValidation(t *testing.T) {
	if !IsValidation(Error{Kind: KindValidation}) {
		t.Fatal("expected validation error")
	}

	if IsValidation(Error{Kind: KindInternal}) {
		t.Fatal("expected non-validation error")
	}

	if IsValidation(errors.New("plain")) {
		t.Fatal("expected plain error to be non-validation")
	}
}

func TestDefaultKindIsValidation(t *testing.T) {
	err := Error{Code: "TEST", Message: "message"}

	if err.effectiveKind() != KindValidation {
		t.Fatalf("expected default kind validation, got %s", err.effectiveKind())
	}

	if err.LogLevel() != zapcore.WarnLevel {
		t.Fatalf("expected warn level for default validation error")
	}
}
