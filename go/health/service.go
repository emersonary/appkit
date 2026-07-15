package health

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

type Service struct {
	db     *pgxpool.Pool
	nc     *nats.Conn
	logger *zap.Logger
}

func NewService(db *pgxpool.Pool, nc *nats.Conn, logger *zap.Logger) *Service {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &Service{
		db:     db,
		nc:     nc,
		logger: logger.Named("service"),
	}
}

func (s *Service) IsHealthy() bool {
	return true
}

func (s *Service) IsInfraHealthy() error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	s.logger.Debug("checking infra health")

	if err := s.db.Ping(ctx); err != nil {
		s.logger.Error("postgres health check failed", zap.Error(err))
		return err
	}

	if s.nc != nil && !s.nc.IsConnected() {
		s.logger.Error("nats health check failed", zap.Error(ErrNatsDisconnected))
		return ErrNatsDisconnected
	}

	s.logger.Debug("infra health check passed")
	return nil
}
