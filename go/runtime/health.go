package runtime

import (
	"github.com/emersonary/appkit/health"
)

func (a *Application[T]) createHealthHandler() {
	healthLogger := a.Logger.Named("health")
	healthService := health.NewService(a.Pool, a.NATS, healthLogger)
	a.HealthHandler = health.NewHandler(healthService, healthLogger, a.Base().AppName())
}
