package runtime

import (
	"github.com/emersonary/appkit/health"
)

func (a *Application[T]) createHealthService() {
	a.Health = health.NewService(a.Pool, a.NATS, a.Logger.Named("health"))
}
