package ai

import (
	"context"
	"fmt"

	"go.uber.org/zap"
)

// WireOptions configures optional AI wiring behavior.
type WireOptions struct {
	Logger *zap.Logger
}

// Wire builds the AI service when cfg.Enabled is true. Returns nil, nil when disabled.
func Wire(ctx context.Context, cfg AIConfig, opts WireOptions) (*Service, error) {
	if !cfg.Enabled {
		return nil, nil
	}

	resolved, err := ResolveBlockConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("resolve ai config: %w", err)
	}

	svc, err := NewService(resolved, opts.Logger)
	if err != nil {
		return nil, err
	}

	svc.logger.Info("ai block enabled",
		zap.Any("routes", svc.RouteSummary()),
	)
	return svc, nil
}
