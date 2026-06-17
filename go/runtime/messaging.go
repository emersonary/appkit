package runtime

import (
	"fmt"

	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

func (a *Application[T]) createNATSConnection(url string) error {
	nc, err := nats.Connect(url)
	if err != nil {
		return fmt.Errorf("connect nats: %w", err)
	}

	a.NATS = nc
	a.Logger.Info("nats connected", zap.String("url", url))
	return nil
}
