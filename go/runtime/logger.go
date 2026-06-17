package runtime

import (
	appkitlog "github.com/emersonary/appkit/log"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func (a *Application[T]) createLogger(cfg appkitlog.Config) error {
	logger, err := appkitlog.New(cfg)
	if err != nil {
		return err
	}

	a.Logger = logger.WithOptions(
		zap.AddCaller(),
		zap.AddStacktrace(zapcore.ErrorLevel),
	)
	return nil
}

func (a *Application[T]) LogError(msg string, err error, fields ...zap.Field) {
	appkitlog.Log(a.Logger, msg, err, fields...)
}

func (a *Application[T]) Sync() {
	if a.Logger != nil {
		_ = a.Logger.Sync()
	}
}
