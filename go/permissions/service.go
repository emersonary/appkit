package permissions

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
	"go.uber.org/zap"
)

type Service struct {
	setup  Setup
	store  *Store
	logger *zap.Logger
}

func NewService(db *sql.DB, setup Setup, logger *zap.Logger) (*Service, error) {
	setup.normalize()
	if err := setup.Validate(); err != nil {
		return nil, err
	}
	if logger == nil {
		logger = zap.NewNop()
	}
	return &Service{
		setup:  setup,
		store:  NewStore(db, setup, logger.Named("store")),
		logger: logger,
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
		if IsNotFound(err) || isAccountProfileLookupSkippable(err) {
			idProfile = s.setup.DefaultProfile
		} else {
			return nil, err
		}
	}
	return s.store.ListFlatForProfile(ctx, idProfile)
}

// isAccountProfileLookupSkippable reports lookup errors that should fall back to default_profile
// (e.g. launch JWT subjects that are not persisted account UUIDs).
func isAccountProfileLookupSkippable(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "22P02", "42P01": // invalid uuid / missing accounts table
			return true
		}
	}
	return false
}

// GroupNameForCategory returns the display name of the permission group for a category id.
func (s *Service) GroupNameForCategory(categoryID string) string {
	categoryID = strings.TrimSpace(categoryID)
	for _, category := range s.setup.Categories {
		if strings.TrimSpace(category.ID) != categoryID {
			continue
		}
		groupID := strings.TrimSpace(category.Group)
		for _, group := range s.setup.Groups {
			if strings.TrimSpace(group.ID) == groupID {
				return strings.TrimSpace(group.Name)
			}
		}
	}
	return ""
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
