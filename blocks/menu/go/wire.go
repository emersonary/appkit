package menu

import (
	"context"
	"fmt"

	"github.com/emersonary/appkit/permissions"
	"go.uber.org/zap"
)

// WireOptions configures menu wiring.
type WireOptions struct {
	Permissions *permissions.Service
	Logger      *zap.Logger
}

// Wire builds the menu service when app.Enabled is true. Returns nil, nil when disabled.
func Wire(ctx context.Context, app AppConfig, opts WireOptions) (*Service, error) {
	if !app.Enabled {
		return nil, nil
	}

	if opts.Permissions == nil {
		return nil, ErrPermissionsRequired
	}

	setup, err := ResolveSetup(app)
	if err != nil {
		return nil, fmt.Errorf("resolve menu setup: %w", err)
	}

	svc, err := NewService(setup, opts.Permissions, opts.Logger)
	if err != nil {
		return nil, err
	}

	svc.logger.Info("menu block enabled",
		zap.Int("menus", len(svc.Setup().Menus)),
	)
	return svc, nil
}
