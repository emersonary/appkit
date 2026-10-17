package docupload

import (
	"context"
	"time"

	"go.uber.org/zap"
)

// StartSweeper deletes expired sessions until ctx is cancelled.
func (s *Service) StartSweeper(ctx context.Context, every time.Duration) {
	if s == nil || ctx == nil {
		return
	}
	if every <= 0 {
		every = time.Minute
	}
	go func() {
		ticker := time.NewTicker(every)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if _, err := s.Sweep(context.Background(), s.now()); err != nil {
					s.logger.Warn("docupload sweep failed", zap.Error(err))
				}
			}
		}
	}()
}
