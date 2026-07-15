package tenants

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"
)

type Service struct {
	store  *Store
	cfg    Config
	logger *zap.Logger
}

func New(db *sql.DB, cfg Config, logger *zap.Logger) (*Service, error) {
	cfg.normalize()
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	if logger == nil {
		logger = zap.NewNop()
	}
	return &Service{
		store:  NewStore(db, cfg.Schema, logger.Named("store")),
		cfg:    cfg,
		logger: logger,
	}, nil
}

func (s *Service) CreateTenant(ctx context.Context, accountID, name, slug, timezone string) (Membership, error) {
	if s.cfg.IsFixedMode() {
		return Membership{}, ErrFixedMode
	}

	accountID = strings.TrimSpace(accountID)
	name = strings.TrimSpace(name)
	slug = normalizeSlug(slug)
	if accountID == "" || name == "" {
		return Membership{}, ErrInvalidArgument
	}
	if err := validateSlug(slug); err != nil {
		return Membership{}, ErrInvalidArgument.With("slug", err.Error())
	}

	tenant, err := s.store.CreateTenant(ctx, slug, name, timezone)
	if err != nil {
		if isUniqueViolation(err) {
			return Membership{}, ErrAlreadyExists
		}
		return Membership{}, err
	}

	if err := s.store.AddMember(ctx, tenant.ID, accountID, RoleOwner); err != nil {
		return Membership{}, err
	}

	return Membership{Tenant: tenant, Role: RoleOwner}, nil
}

func (s *Service) ListMyTenants(ctx context.Context, accountID string) ([]Membership, error) {
	accountID = strings.TrimSpace(accountID)
	if accountID == "" {
		return nil, ErrUnauthenticated
	}
	return s.store.ListMemberships(ctx, accountID)
}

func (s *Service) GetTenant(ctx context.Context, accountID, tenantID string) (Tenant, error) {
	if _, err := s.store.GetMembership(ctx, tenantID, accountID); err != nil {
		return Tenant{}, ErrForbidden
	}
	return s.store.GetTenant(ctx, tenantID)
}

func (s *Service) InviteMember(ctx context.Context, inviterAccountID, tenantID, email, role string) (string, error) {
	inviterRole, err := s.store.GetMembership(ctx, tenantID, inviterAccountID)
	if err != nil {
		return "", ErrForbidden
	}
	if inviterRole != RoleOwner && inviterRole != RoleAdmin {
		return "", ErrForbidden
	}

	role = strings.TrimSpace(strings.ToLower(role))
	if role == "" {
		role = RoleStaff
	}
	if !isValidRole(role) || role == RoleOwner {
		return "", ErrInvalidArgument.With("role", role)
	}

	email = strings.TrimSpace(strings.ToLower(email))
	if email == "" {
		return "", ErrInvalidArgument.With("email", "required")
	}

	token, tokenHash, err := newInviteToken()
	if err != nil {
		return "", err
	}

	_, err = s.store.CreateInvite(ctx, tenantID, email, role, tokenHash, time.Now().Add(s.cfg.inviteTokenTTL()))
	if err != nil {
		return "", err
	}

	return token, nil
}

func (s *Service) AcceptInvite(ctx context.Context, accountID, token string) (Membership, error) {
	accountID = strings.TrimSpace(accountID)
	token = strings.TrimSpace(token)
	if accountID == "" || token == "" {
		return Membership{}, ErrInvalidArgument
	}
	return s.store.AcceptInvite(ctx, accountID, hashToken(token))
}

// Config returns the loaded tenants block configuration.
func (s *Service) Config() Config {
	return s.cfg
}

// SyncFixedCatalog upserts fixed-mode feed entries into tenant.tenants.
func (s *Service) SyncFixedCatalog(ctx context.Context) (int, error) {
	if !s.cfg.IsFixedMode() {
		return 0, nil
	}

	var synced int
	for _, entry := range s.cfg.Feed {
		if _, err := s.store.UpsertCatalogTenant(ctx, entry.ID, entry.Name, entry.Timezone); err != nil {
			return synced, err
		}
		synced++
	}
	return synced, nil
}

func isValidRole(role string) bool {
	switch role {
	case RoleOwner, RoleAdmin, RoleStaff, RoleViewer:
		return true
	default:
		return false
	}
}

func newInviteToken() (plain, hashed string, err error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", "", fmt.Errorf("invite token: %w", err)
	}
	plain = base64.RawURLEncoding.EncodeToString(buf)
	sum := sha256.Sum256([]byte(plain))
	hashed = hex.EncodeToString(sum[:])
	return plain, hashed, nil
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
