package permissions

import (
	"context"
	"database/sql"
)

type Service struct {
	setup Setup
	store *Store
}

func NewService(db *sql.DB, setup Setup) (*Service, error) {
	setup.normalize()
	if err := setup.Validate(); err != nil {
		return nil, err
	}
	return &Service{
		setup: setup,
		store: NewStore(db, setup),
	}, nil
}

func (s *Service) Setup() Setup {
	return s.setup
}

func (s *Service) Store() *Store {
	return s.store
}

func (s *Service) DefaultProfile() string {
	return s.setup.DefaultProfile
}

func (s *Service) AdminProfile() string {
	return s.setup.AdminProfileID()
}

func (s *Service) HasPermission(ctx context.Context, accountID, idPermission string, actionBit int) (bool, error) {
	if !ValidAction(actionBit) {
		return false, ErrInvalidAction
	}

	idProfile, err := s.store.GetAccountProfileID(ctx, accountID)
	if err != nil {
		if IsNotFound(err) {
			idProfile = s.setup.DefaultProfile
		} else {
			return false, err
		}
	}

	return s.HasProfilePermission(ctx, idProfile, idPermission, actionBit)
}

func (s *Service) HasProfilePermission(ctx context.Context, idProfile, idPermission string, actionBit int) (bool, error) {
	if !ValidAction(actionBit) {
		return false, ErrInvalidAction
	}
	return s.store.HasProfilePermission(ctx, idProfile, idPermission, actionBit)
}

func (s *Service) RequirePermission(ctx context.Context, accountID, idPermission string, actionBit int) error {
	ok, err := s.HasPermission(ctx, accountID, idPermission, actionBit)
	if err != nil {
		return err
	}
	if !ok {
		return ErrPermissionDenied.With("permission", idPermission)
	}
	return nil
}

func (s *Service) ListFlatForAccount(ctx context.Context, accountID string) ([]FlatPermission, error) {
	idProfile, err := s.store.GetAccountProfileID(ctx, accountID)
	if err != nil {
		if IsNotFound(err) {
			idProfile = s.setup.DefaultProfile
		} else {
			return nil, err
		}
	}
	return s.store.ListFlatForProfile(ctx, idProfile)
}

func (s *Service) ListCatalog(ctx context.Context) (groups []Group, categories []Category, permissions []Permission, err error) {
	groups, err = s.store.ListGroups(ctx)
	if err != nil {
		return nil, nil, nil, err
	}
	categories, err = s.store.ListCategories(ctx)
	if err != nil {
		return nil, nil, nil, err
	}
	permissions, err = s.store.ListPermissions(ctx)
	if err != nil {
		return nil, nil, nil, err
	}
	return groups, categories, permissions, nil
}

func (s *Service) AssignAccountProfile(ctx context.Context, accountID, idProfile string) error {
	if _, err := s.store.GetProfile(ctx, idProfile); err != nil {
		if IsNotFound(err) {
			return ErrProfileNotFound.With("id_profile", idProfile)
		}
		return err
	}
	return s.store.SetAccountProfileID(ctx, accountID, idProfile)
}

func (s *Service) EnsureAccountDefaultProfile(ctx context.Context, accountID string) error {
	return s.store.EnsureAccountDefaultProfile(ctx, accountID, s.setup.DefaultProfile)
}

// AssignNewAccountProfile sets profile for a newly created account.
func (s *Service) AssignNewAccountProfile(ctx context.Context, accountID string, asAdmin bool) error {
	profileID := s.setup.DefaultProfile
	if asAdmin {
		profileID = s.setup.AdminProfileID()
	}
	return s.AssignAccountProfile(ctx, accountID, profileID)
}
