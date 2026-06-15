package log

import (
	"errors"
	"os"
	"strings"

	"github.com/emersonary/appkit/apperror"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Config struct {
	Level  string
	Format string
}

func New(cfg Config) (*zap.Logger, error) {
	level, err := parseLevel(cfg.Level)
	if err != nil {
		return nil, err
	}

	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.TimeKey = "timestamp"
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderCfg.MessageKey = "message"
	encoderCfg.LevelKey = "level"

	var encoder zapcore.Encoder
	if strings.EqualFold(cfg.Format, "text") {
		encoder = zapcore.NewConsoleEncoder(encoderCfg)
	} else {
		encoder = zapcore.NewJSONEncoder(encoderCfg)
	}

	core := zapcore.NewCore(
		encoder,
		zapcore.AddSync(os.Stdout),
		level,
	)

	return zap.New(core), nil
}

func Log(logger *zap.Logger, msg string, err error, fields ...zap.Field) {
	if logger == nil {
		return
	}

	var appErr apperror.Error
	if errors.As(err, &appErr) {
		fields = append(fields, appErr.ZapFields()...)
		logger.Log(appErr.LogLevel(), msg, fields...)
		return
	}

	fields = append(fields, zap.Error(err))
	logger.Error(msg, fields...)
}

func parseLevel(value string) (zapcore.Level, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", "info":
		return zapcore.InfoLevel, nil
	case "debug":
		return zapcore.DebugLevel, nil
	case "warn", "warning":
		return zapcore.WarnLevel, nil
	case "error":
		return zapcore.ErrorLevel, nil
	default:
		return zapcore.InfoLevel, nil
	}
}
