package menu

import (
	"context"
	"strings"

	"github.com/emersonary/appkit/permissions"
	"go.uber.org/zap"
)

// Service resolves backend-defined menus for authenticated accounts.
type Service struct {
	setup       Setup
	permissions *permissions.Service
	logger      *zap.Logger
}

func NewService(setup Setup, perms *permissions.Service, logger *zap.Logger) (*Service, error) {
	if perms == nil {
		return nil, ErrPermissionsRequired
	}
	setup.normalize()
	if err := setup.Validate(); err != nil {
		return nil, err
	}
	if err := setup.ValidateAgainstPermissions(perms.Setup()); err != nil {
		return nil, err
	}
	if logger == nil {
		logger = zap.NewNop()
	}
	return &Service{setup: setup, permissions: perms, logger: logger}, nil
}

func (s *Service) Setup() Setup {
	return s.setup
}

func (s *Service) GetMenu(ctx context.Context, accountID string) (Layout, error) {
	accountID = strings.TrimSpace(accountID)
	if accountID == "" {
		return Layout{}, ErrUnauthenticated.With("account_id", "required")
	}

	grants, err := s.permissions.ListFlatForAccount(ctx, accountID)
	if err != nil {
		return Layout{}, err
	}

	tree, err := permissions.NewPermissionTreeFromConfigs(s.permissions.Setup().Permissions)
	if err != nil {
		return Layout{}, err
	}

	return ResolveLayout(s.setup, tree, grants, s.permissions.GroupNameForCategory)
}
