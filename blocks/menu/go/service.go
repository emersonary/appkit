package menu

import (
	"context"
	"strings"

	"github.com/emersonary/appkit/permissions"
)

// Service resolves backend-defined menus for authenticated accounts.
type Service struct {
	setup       Setup
	permissions *permissions.Service
}

func NewService(setup Setup, perms *permissions.Service) (*Service, error) {
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
	return &Service{setup: setup, permissions: perms}, nil
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
