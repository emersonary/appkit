package log

import (
	"errors"
	"testing"

	"github.com/emersonary/appkit/apperror"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func TestLogAppErrorUsesStructuredFields(t *testing.T) {
	core, recorded := observer.New(zap.DebugLevel)
	logger := zap.New(core)

	appErr := apperror.Error{
		Code:    "TRIP_TYPE_REQUIRED",
		Domain:  "trip",
		Field:   "trip_type",
		Message: "trip type is required",
		Kind:    apperror.KindValidation,
	}

	Log(logger, "create trip failed", appErr)

	if recorded.Len() != 1 {
		t.Fatalf("expected 1 log entry, got %d", recorded.Len())
	}

	entry := recorded.All()[0]
	if entry.Level != zap.WarnLevel {
		t.Fatalf("expected warn level, got %s", entry.Level)
	}

	contextMap := entry.ContextMap()
	if contextMap["error.code"] != "TRIP_TYPE_REQUIRED" {
		t.Fatalf("expected structured error.code, got %v", contextMap["error.code"])
	}
}

func TestLogPlainErrorUsesErrorLevel(t *testing.T) {
	core, recorded := observer.New(zap.DebugLevel)
	logger := zap.New(core)

	Log(logger, "database connection failed", errors.New("connection refused"))

	if recorded.Len() != 1 {
		t.Fatalf("expected 1 log entry, got %d", recorded.Len())
	}

	entry := recorded.All()[0]
	if entry.Level != zap.ErrorLevel {
		t.Fatalf("expected error level, got %s", entry.Level)
	}
}

func TestParseLevel(t *testing.T) {
	tests := map[string]zapcore.Level{
		"debug":   zapcore.DebugLevel,
		"info":    zapcore.InfoLevel,
		"warn":    zapcore.WarnLevel,
		"error":   zapcore.ErrorLevel,
		"warning": zapcore.WarnLevel,
	}

	for input, want := range tests {
		got, err := parseLevel(input)
		if err != nil {
			t.Fatalf("parseLevel(%q) returned error: %v", input, err)
		}

		if got != want {
			t.Fatalf("parseLevel(%q) = %v, want %v", input, got, want)
		}
	}
}

func TestNewLoggerJSON(t *testing.T) {
	logger, err := New(Config{Level: "info", Format: "json"})
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}

	logger.Info("test message")
}

func TestNewLoggerText(t *testing.T) {
	logger, err := New(Config{Level: "debug", Format: "text"})
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}

	logger.Debug("test message")
}
