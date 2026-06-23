package social

import (
	"context"
	"fmt"
)

// WireOptions configures optional social wiring behavior.
type WireOptions struct{}

// Wire builds the social service when cfg.Enabled is true. Returns nil, nil when disabled.
func Wire(ctx context.Context, cfg SocialConfig, _ WireOptions) (*Service, error) {
	if !cfg.Enabled {
		return nil, nil
	}

	resolved, err := ResolveBlockConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("resolve social config: %w", err)
	}

	templates, err := NewTemplateRenderer()
	if err != nil {
		return nil, err
	}

	svc, err := NewService(resolved, templates, nil)
	if err != nil {
		return nil, err
	}

	return svc, nil
}
