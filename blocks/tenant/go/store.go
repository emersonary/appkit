package tenants

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type Store struct {
	db     *sql.DB
	schema string
	logger *zap.Logger
}

func NewStore(db *sql.DB, schema string, logger *zap.Logger) *Store {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &Store{db: db, schema: schema, logger: logger}
}

func (s *Store) tenantsTable() string     { return qualifiedName(s.schema, "tenants") }
func (s *Store) accountsTable() string    { return qualifiedName(s.schema, "tenant_accounts") }
func (s *Store) invitesTable() string     { return qualifiedName(s.schema, "tenant_invites") }

func (s *Store) CreateTenant(ctx context.Context, slug, name, timezone string) (Tenant, error) {
	if timezone == "" {
		timezone = "UTC"
	}
	var tenant Tenant
	err := s.db.QueryRowContext(ctx, `
		INSERT INTO `+s.tenantsTable()+` (slug, name, timezone, status)
		VALUES ($1, $2, $3, $4)
		RETURNING id, slug, name, timezone, status
	`, slug, name, timezone, StatusTrial).Scan(
		&tenant.ID, &tenant.Slug, &tenant.Name, &tenant.Timezone, &tenant.Status,
	)
	return tenant, err
}

func (s *Store) UpsertCatalogTenant(ctx context.Context, slug, name, timezone string) (Tenant, error) {
	if timezone == "" {
		timezone = "UTC"
	}
	var tenant Tenant
	err := s.db.QueryRowContext(ctx, `
		INSERT INTO `+s.tenantsTable()+` (slug, name, timezone, status)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (slug) DO UPDATE
		SET name = EXCLUDED.name,
		    timezone = EXCLUDED.timezone,
		    status = EXCLUDED.status,
		    updated_at = NOW()
		RETURNING id, slug, name, timezone, status
	`, slug, name, timezone, StatusActive).Scan(
		&tenant.ID, &tenant.Slug, &tenant.Name, &tenant.Timezone, &tenant.Status,
	)
	return tenant, err
}

func (s *Store) AddMember(ctx context.Context, tenantID, accountID, role string) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO `+s.accountsTable()+` (tenant_id, account_id, role)
		VALUES ($1, $2, $3)
		ON CONFLICT (tenant_id, account_id) DO UPDATE SET role = EXCLUDED.role, updated_at = NOW()
	`, tenantID, accountID, role)
	return err
}

func (s *Store) GetTenant(ctx context.Context, tenantID string) (Tenant, error) {
	var tenant Tenant
	err := s.db.QueryRowContext(ctx, `
		SELECT id, slug, name, timezone, status
		FROM `+s.tenantsTable()+`
		WHERE id = $1
	`, tenantID).Scan(&tenant.ID, &tenant.Slug, &tenant.Name, &tenant.Timezone, &tenant.Status)
	if errors.Is(err, sql.ErrNoRows) {
		return Tenant{}, ErrNotFound
	}
	return tenant, err
}

func (s *Store) GetMembership(ctx context.Context, tenantID, accountID string) (string, error) {
	var role string
	err := s.db.QueryRowContext(ctx, `
		SELECT role FROM `+s.accountsTable()+`
		WHERE tenant_id = $1 AND account_id = $2
	`, tenantID, accountID).Scan(&role)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrNotFound
	}
	return role, err
}

func (s *Store) ListMemberships(ctx context.Context, accountID string) ([]Membership, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT t.id, t.slug, t.name, t.timezone, t.status, ta.role
		FROM `+s.accountsTable()+` ta
		INNER JOIN `+s.tenantsTable()+` t ON t.id = ta.tenant_id
		WHERE ta.account_id = $1
		ORDER BY t.name
	`, accountID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Membership
	for rows.Next() {
		var m Membership
		if err := rows.Scan(
			&m.Tenant.ID, &m.Tenant.Slug, &m.Tenant.Name, &m.Tenant.Timezone, &m.Tenant.Status, &m.Role,
		); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

func (s *Store) CreateInvite(ctx context.Context, tenantID, email, role, tokenHash string, expiresAt time.Time) (string, error) {
	id := uuid.NewString()
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO `+s.invitesTable()+` (id, tenant_id, email, role, token_hash, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, id, tenantID, strings.ToLower(strings.TrimSpace(email)), role, tokenHash, expiresAt)
	return id, err
}

func (s *Store) AcceptInvite(ctx context.Context, accountID, tokenHash string) (Membership, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Membership{}, err
	}
	defer tx.Rollback()

	var tenantID, role, email string
	var expiresAt time.Time
	var acceptedAt sql.NullTime
	err = tx.QueryRowContext(ctx, `
		SELECT tenant_id, email, role, expires_at, accepted_at
		FROM `+s.invitesTable()+`
		WHERE token_hash = $1
	`, tokenHash).Scan(&tenantID, &email, &role, &expiresAt, &acceptedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return Membership{}, ErrInvalidToken
	}
	if err != nil {
		return Membership{}, err
	}
	if acceptedAt.Valid {
		return Membership{}, ErrInvalidToken
	}
	if time.Now().After(expiresAt) {
		return Membership{}, ErrInvalidToken
	}

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO `+s.accountsTable()+` (tenant_id, account_id, role)
		VALUES ($1, $2, $3)
		ON CONFLICT (tenant_id, account_id) DO UPDATE SET role = EXCLUDED.role, updated_at = NOW()
	`, tenantID, accountID, role); err != nil {
		return Membership{}, err
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE `+s.invitesTable()+` SET accepted_at = NOW() WHERE token_hash = $1
	`, tokenHash); err != nil {
		return Membership{}, err
	}

	var tenant Tenant
	err = tx.QueryRowContext(ctx, `
		SELECT id, slug, name, timezone, status FROM `+s.tenantsTable()+` WHERE id = $1
	`, tenantID).Scan(&tenant.ID, &tenant.Slug, &tenant.Name, &tenant.Timezone, &tenant.Status)
	if err != nil {
		return Membership{}, err
	}

	if err := tx.Commit(); err != nil {
		return Membership{}, err
	}

	return Membership{Tenant: tenant, Role: role}, nil
}

func isUniqueViolation(err error) bool {
	return err != nil && strings.Contains(strings.ToLower(err.Error()), "duplicate")
}
