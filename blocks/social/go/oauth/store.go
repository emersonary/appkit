package oauth

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ConnectionStore persists tenant publishing OAuth tokens.
type ConnectionStore interface {
	Get(ctx context.Context, tenantID, platformID, language string) (Connection, error)
	Upsert(ctx context.Context, conn Connection) error
	Delete(ctx context.Context, tenantID, platformID, language string) error
	ValidAt(ctx context.Context, tenantID, platformID, language string, now time.Time) (Connection, bool, error)
}

// StoreOptions configures the qualified connection table for host apps.
type StoreOptions struct {
	// Table is the fully qualified table name, e.g. "posts.social_platform_connections".
	Table string
	// TenantColumn is the tenant id column name (project_id or tenant_id).
	TenantColumn string
	// LanguageColumn is the language id column name (id_language).
	LanguageColumn string
}

func (o StoreOptions) normalized() StoreOptions {
	out := o
	if out.Table == "" {
		out.Table = "social.platform_connections"
	}
	if out.TenantColumn == "" {
		out.TenantColumn = "tenant_id"
	}
	if out.LanguageColumn == "" {
		out.LanguageColumn = "id_language"
	}
	return out
}

// PgxConnectionStore implements ConnectionStore on PostgreSQL.
type PgxConnectionStore struct {
	pool     *pgxpool.Pool
	table    string
	tenant   string
	language string
}

func NewPgxConnectionStore(pool *pgxpool.Pool, opts StoreOptions) *PgxConnectionStore {
	o := opts.normalized()
	return &PgxConnectionStore{
		pool:     pool,
		table:    o.Table,
		tenant:   o.TenantColumn,
		language: o.LanguageColumn,
	}
}

func (s *PgxConnectionStore) Get(ctx context.Context, tenantID, platformID, language string) (Connection, error) {
	q := fmt.Sprintf(`
		SELECT id, %s, platform_id, %s, account_id, access_token, expires_at, scopes, created_at, updated_at
		FROM %s
		WHERE %s = $1 AND platform_id = $2 AND %s = $3
	`, s.tenant, s.language, s.table, s.tenant, s.language)

	row := s.pool.QueryRow(ctx, q, tenantID, platformID, language)

	var conn Connection
	if err := row.Scan(
		&conn.ID, &conn.TenantID, &conn.PlatformID, &conn.Language, &conn.AccountID, &conn.AccessToken,
		&conn.ExpiresAt, &conn.Scopes, &conn.CreatedAt, &conn.UpdatedAt,
	); err != nil {
		if err == pgx.ErrNoRows {
			return Connection{}, fmt.Errorf("social connection not found")
		}
		return Connection{}, err
	}
	return conn, nil
}

func (s *PgxConnectionStore) Upsert(ctx context.Context, conn Connection) error {
	language := strings.TrimSpace(conn.Language)
	if language == "" {
		return fmt.Errorf("social connection language required")
	}
	q := fmt.Sprintf(`
		INSERT INTO %s (%s, platform_id, %s, account_id, access_token, expires_at, scopes)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (%s, platform_id, %s) DO UPDATE SET
			account_id = EXCLUDED.account_id,
			access_token = EXCLUDED.access_token,
			expires_at = EXCLUDED.expires_at,
			scopes = EXCLUDED.scopes,
			updated_at = NOW()
	`, s.table, s.tenant, s.language, s.tenant, s.language)

	_, err := s.pool.Exec(ctx, q,
		conn.TenantID, conn.PlatformID, language, conn.AccountID, conn.AccessToken, conn.ExpiresAt, conn.Scopes,
	)
	return err
}

func (s *PgxConnectionStore) Delete(ctx context.Context, tenantID, platformID, language string) error {
	q := fmt.Sprintf(`DELETE FROM %s WHERE %s = $1 AND platform_id = $2 AND %s = $3`, s.table, s.tenant, s.language)
	_, err := s.pool.Exec(ctx, q, tenantID, platformID, language)
	return err
}

func (s *PgxConnectionStore) ValidAt(ctx context.Context, tenantID, platformID, language string, now time.Time) (Connection, bool, error) {
	conn, err := s.Get(ctx, tenantID, platformID, language)
	if err != nil {
		return Connection{}, false, nil
	}
	return conn, conn.ValidAt(now), nil
}
