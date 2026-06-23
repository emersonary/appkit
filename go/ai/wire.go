package ai

import (
	"context"
	"fmt"
)

// WireOptions configures optional AI wiring behavior.
type WireOptions struct{}

// Wire builds the AI service when cfg.Enabled is true. Returns nil, nil when disabled.
func Wire(ctx context.Context, cfg AIConfig, _ WireOptions) (*Service, error) {
	if !cfg.Enabled {
		return nil, nil
	}

	resolved, err := ResolveBlockConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("resolve ai config: %w", err)
	}

	svc, err := NewService(resolved)
	if err != nil {
		return nil, err
	}

	return svc, nil
}
