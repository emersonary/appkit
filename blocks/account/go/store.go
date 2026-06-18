package accounts

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

type Store struct {
	db     *sql.DB
	schema string
}

func NewStore(db *sql.DB, cfg Config) *Store {
	return &Store{
		db:     db,
		schema: cfg.Schema,
	}
}

func (s *Store) accountsTable() string {
	return qualifiedName(s.schema, "accounts")
}

func (s *Store) oauthTable() string {
	return qualifiedName(s.schema, "oauth_identities")
}

func (s *Store) verifyTable() string {
	return qualifiedName(s.schema, "email_verification_tokens")
}

func (s *Store) resetTable() string {
	return qualifiedName(s.schema, "password_reset_tokens")
}

func (s *Store) accountTenantsTable() string {
	return qualifiedName(s.schema, "account_tenants")
}

func (s *Store) selectColumns() string {
	return `id, email, password_hash, first_name, last_name, avatar_url, email_verified_at, is_admin, created_at, updated_at`
}

func (s *Store) Create(ctx context.Context, email, passwordHash string, firstName, lastName *string, emailVerified, isAdmin bool) (Account, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	row := s.db.QueryRowContext(ctx, `
		INSERT INTO `+s.accountsTable()+` (email, password_hash, first_name, last_name, email_verified_at, is_admin)
		VALUES ($1, NULLIF($2, ''), $3, $4, CASE WHEN $5 THEN NOW() ELSE NULL END, $6)
		RETURNING `+s.selectColumns()+`
	`, email, passwordHash, firstName, lastName, emailVerified, isAdmin)
	return scanAccount(row)
}

func (s *Store) GetByEmail(ctx context.Context, email string) (Account, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	row := s.db.QueryRowContext(ctx, `
		SELECT `+s.selectColumns()+`
		FROM `+s.accountsTable()+` WHERE email = $1
	`, email)
	return scanAccount(row)
}

func (s *Store) GetByID(ctx context.Context, id string) (Account, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT `+s.selectColumns()+`
		FROM `+s.accountsTable()+` WHERE id = $1
	`, id)
	return scanAccount(row)
}

func (s *Store) UpdatePasswordHash(ctx context.Context, accountID, passwordHash string) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE `+s.accountsTable()+` SET password_hash = $2, updated_at = NOW() WHERE id = $1
	`, accountID, passwordHash)
	return err
}

func (s *Store) FindOrCreateOAuthAccount(
	ctx context.Context,
	provider, providerUserID, email string,
	firstName, lastName *string,
	avatarURL *string,
) (Account, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Account{}, err
	}
	defer tx.Rollback()

	var account Account
	err = tx.QueryRowContext(ctx, `
		SELECT a.id, a.email, a.password_hash, a.first_name, a.last_name, a.avatar_url, a.email_verified_at,
		       a.is_admin, a.created_at, a.updated_at
		FROM `+s.oauthTable()+` o
		JOIN `+s.accountsTable()+` a ON a.id = o.account_id
		WHERE o.provider = $1 AND o.provider_user_id = $2
	`, provider, providerUserID).Scan(
		&account.ID, &account.Email, &account.PasswordHash, &account.FirstName, &account.LastName, &account.AvatarURL,
		&account.EmailVerifiedAt, &account.IsAdmin, &account.CreatedAt, &account.UpdatedAt,
	)
	if err == nil {
		account, err = refreshOAuthProfile(ctx, tx, s, account, firstName, lastName, avatarURL)
		if err != nil {
			return Account{}, err
		}
		if err := tx.Commit(); err != nil {
			return Account{}, err
		}
		return account, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return Account{}, err
	}

	email = strings.ToLower(strings.TrimSpace(email))
	err = tx.QueryRowContext(ctx, `
		SELECT `+s.selectColumns()+`
		FROM `+s.accountsTable()+` WHERE email = $1
	`, email).Scan(
		&account.ID, &account.Email, &account.PasswordHash, &account.FirstName, &account.LastName, &account.AvatarURL,
		&account.EmailVerifiedAt, &account.IsAdmin, &account.CreatedAt, &account.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		err = tx.QueryRowContext(ctx, `
			INSERT INTO `+s.accountsTable()+` (email, first_name, last_name, avatar_url, email_verified_at)
			VALUES ($1, $2, $3, $4, NOW())
			RETURNING `+s.selectColumns()+`
		`, email, firstName, lastName, avatarURL).Scan(
			&account.ID, &account.Email, &account.PasswordHash, &account.FirstName, &account.LastName, &account.AvatarURL,
			&account.EmailVerifiedAt, &account.IsAdmin, &account.CreatedAt, &account.UpdatedAt,
		)
	} else if err == nil {
		account, err = refreshOAuthProfile(ctx, tx, s, account, firstName, lastName, avatarURL)
		if err == nil && account.EmailVerifiedAt == nil {
			_, err = tx.ExecContext(ctx, `
				UPDATE `+s.accountsTable()+` SET email_verified_at = NOW() WHERE id = $1
			`, account.ID)
			if err == nil {
				now := time.Now()
				account.EmailVerifiedAt = &now
			}
		}
	}
	if err != nil {
		return Account{}, err
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO `+s.oauthTable()+` (account_id, provider, provider_user_id)
		VALUES ($1, $2, $3)
		ON CONFLICT (provider, provider_user_id) DO NOTHING
	`, account.ID, provider, providerUserID)
	if err != nil {
		return Account{}, err
	}

	if err := tx.Commit(); err != nil {
		return Account{}, err
	}
	return account, nil
}

func refreshOAuthProfile(
	ctx context.Context,
	tx *sql.Tx,
	s *Store,
	account Account,
	firstName, lastName *string,
	avatarURL *string,
) (Account, error) {
	if (firstName == nil || strings.TrimSpace(*firstName) == "") &&
		(lastName == nil || strings.TrimSpace(*lastName) == "") &&
		(avatarURL == nil || strings.TrimSpace(*avatarURL) == "") {
		return account, nil
	}

	err := tx.QueryRowContext(ctx, `
		UPDATE `+s.accountsTable()+`
		SET
			first_name = CASE
				WHEN $2::text IS NOT NULL AND NULLIF(TRIM($2::text), '') IS NOT NULL THEN $2::text
				ELSE first_name
			END,
			last_name = CASE
				WHEN $3::text IS NOT NULL AND NULLIF(TRIM($3::text), '') IS NOT NULL THEN $3::text
				ELSE last_name
			END,
			avatar_url = CASE
				WHEN $4::text IS NOT NULL AND NULLIF(TRIM($4::text), '') IS NOT NULL THEN $4::text
				ELSE avatar_url
			END
		WHERE id = $1
		RETURNING `+s.selectColumns()+`
	`, account.ID, firstName, lastName, avatarURL).Scan(
		&account.ID, &account.Email, &account.PasswordHash, &account.FirstName, &account.LastName, &account.AvatarURL,
		&account.EmailVerifiedAt, &account.IsAdmin, &account.CreatedAt, &account.UpdatedAt,
	)
	if err != nil {
		return Account{}, fmt.Errorf("refresh oauth profile: %w", err)
	}
	return account, nil
}

func (s *Store) CreateVerificationToken(ctx context.Context, accountID, tokenHash string, expiresAt time.Time) error {
	_, err := s.db.ExecContext(ctx, `
		DELETE FROM `+s.verifyTable()+`
		WHERE account_id = $1 AND used_at IS NULL
	`, accountID)
	if err != nil {
		return err
	}
	_, err = s.db.ExecContext(ctx, `
		INSERT INTO `+s.verifyTable()+` (account_id, token_hash, expires_at)
		VALUES ($1, $2, $3)
	`, accountID, tokenHash, expiresAt)
	return err
}

func (s *Store) ConsumeVerificationToken(ctx context.Context, tokenHash string, now time.Time) (string, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return "", err
	}
	defer tx.Rollback()

	var accountID string
	err = tx.QueryRowContext(ctx, `
		UPDATE `+s.verifyTable()+`
		SET used_at = $2
		WHERE token_hash = $1
		  AND used_at IS NULL
		  AND expires_at > $2
		RETURNING account_id
	`, tokenHash, now).Scan(&accountID)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrNotFound
	}
	if err != nil {
		return "", err
	}

	_, err = tx.ExecContext(ctx, `
		UPDATE `+s.accountsTable()+` SET email_verified_at = $2 WHERE id = $1
	`, accountID, now)
	if err != nil {
		return "", err
	}
	if err := tx.Commit(); err != nil {
		return "", err
	}
	return accountID, nil
}

func (s *Store) EmailByVerifiedToken(ctx context.Context, tokenHash string) (string, bool) {
	var email string
	err := s.db.QueryRowContext(ctx, `
		SELECT a.email
		FROM `+s.verifyTable()+` t
		JOIN `+s.accountsTable()+` a ON a.id = t.account_id
		WHERE t.token_hash = $1
		  AND t.used_at IS NOT NULL
		  AND a.email_verified_at IS NOT NULL
	`, tokenHash).Scan(&email)
	if errors.Is(err, sql.ErrNoRows) {
		return "", false
	}
	if err != nil {
		return "", false
	}
	return email, true
}

func (s *Store) IsEmailVerified(ctx context.Context, email string) (bool, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	var verified bool
	err := s.db.QueryRowContext(ctx, `
		SELECT email_verified_at IS NOT NULL
		FROM `+s.accountsTable()+`
		WHERE email = $1
	`, email).Scan(&verified)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	return verified, err
}

func (s *Store) CreatePasswordResetToken(ctx context.Context, accountID, tokenHash string, expiresAt time.Time) error {
	_, err := s.db.ExecContext(ctx, `
		DELETE FROM `+s.resetTable()+`
		WHERE account_id = $1 AND used_at IS NULL
	`, accountID)
	if err != nil {
		return err
	}
	_, err = s.db.ExecContext(ctx, `
		INSERT INTO `+s.resetTable()+` (account_id, token_hash, expires_at)
		VALUES ($1, $2, $3)
	`, accountID, tokenHash, expiresAt)
	return err
}

func (s *Store) ConsumePasswordResetToken(ctx context.Context, tokenHash string, now time.Time) (string, error) {
	var accountID string
	err := s.db.QueryRowContext(ctx, `
		UPDATE `+s.resetTable()+`
		SET used_at = $2
		WHERE token_hash = $1
		  AND used_at IS NULL
		  AND expires_at > $2
		RETURNING account_id
	`, tokenHash, now).Scan(&accountID)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrNotFound
	}
	return accountID, err
}

func (s *Store) JoinDefaultTenant(ctx context.Context, accountID, tenantID string) error {
	if tenantID == "" {
		return nil
	}
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO `+s.accountTenantsTable()+` (account_id, tenant_id, role)
		VALUES ($1, $2, 'member')
		ON CONFLICT (account_id, tenant_id) DO NOTHING
	`, accountID, tenantID)
	return err
}

func scanAccount(row *sql.Row) (Account, error) {
	var account Account
	err := row.Scan(
		&account.ID,
		&account.Email,
		&account.PasswordHash,
		&account.FirstName,
		&account.LastName,
		&account.AvatarURL,
		&account.EmailVerifiedAt,
		&account.IsAdmin,
		&account.CreatedAt,
		&account.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return Account{}, ErrNotFound
	}
	if err != nil {
		return Account{}, fmt.Errorf("scan account: %w", err)
	}
	return account, nil
}

func isUniqueViolation(err error) bool {
	return err != nil && strings.Contains(err.Error(), "23505")
}
