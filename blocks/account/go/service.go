package accounts

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/emersonary/appkit/accounts/oauth"
)

type Service struct {
	cfg         Config
	secrets     Secrets
	store       *Store
	tokens      *TokenService
	mailer      Mailer
	oauth       map[string]oauth.Provider
	afterCreate func(ctx context.Context, account Account, registerAsAdmin bool) error
}

type Options struct {
	Mailer Mailer
	// AfterCreate runs after a new account row is inserted (e.g. assign permissions profile).
	AfterCreate func(ctx context.Context, account Account, registerAsAdmin bool) error
}

func (o Options) normalized() Options {
	if o.Mailer == nil {
		o.Mailer = NoopMailer{}
	}
	return o
}

func New(db *sql.DB, cfg Config, secrets Secrets, opts Options) (*Service, error) {
	cfg.normalize()
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	if strings.TrimSpace(secrets.JWTSecret) == "" {
		return nil, ErrInvalidArgument.With("jwt_secret", "required")
	}

	opts = opts.normalized()
	svc := &Service{
		cfg:         cfg,
		secrets:     secrets,
		store:       NewStore(db, cfg),
		tokens:      NewTokenService(secrets.JWTSecret, cfg.AccessTokenTTL(), cfg.EffectiveTenantID()),
		mailer:      opts.Mailer,
		oauth:       map[string]oauth.Provider{},
		afterCreate: opts.AfterCreate,
	}
	return svc, nil
}

func (s *Service) RegisterProvider(provider oauth.Provider) {
	if provider == nil {
		return
	}
	s.oauth[provider.Name()] = provider
}

func (s *Service) Config() Config {
	return s.cfg
}

func (s *Service) Secrets() Secrets {
	return s.secrets
}

func (s *Service) OAccountProvider(name string) (oauth.Provider, bool) {
	p, ok := s.oauth[name]
	return p, ok
}

func (s *Service) Register(ctx context.Context, emailAddr, password string, firstName, lastName *string) (RegisterResult, error) {
	emailAddr = strings.TrimSpace(emailAddr)
	firstName = trimStringPtr(firstName)
	lastName = trimStringPtr(lastName)
	if !isValidEmail(emailAddr) || len(password) < 6 || firstName == nil {
		return RegisterResult{}, ErrInvalidArgument
	}

	if !s.cfg.RegistrationEnabled() {
		return RegisterResult{}, ErrRegistrationDisabled
	}

	hash, err := HashPassword(password)
	if err != nil {
		return RegisterResult{}, err
	}

	skipVerify := s.cfg.SkipEmailVerification()
	registerAsAdmin := s.cfg.RegisterAsAdmin()
	account, err := s.store.Create(ctx, emailAddr, hash, firstName, lastName, skipVerify, registerAsAdmin)
	if err != nil {
		if isUniqueViolation(err) {
			return RegisterResult{}, ErrAlreadyExists
		}
		return RegisterResult{}, err
	}

	if s.afterCreate != nil {
		if err := s.afterCreate(ctx, account, registerAsAdmin); err != nil {
			return RegisterResult{}, err
		}
	}

	if s.cfg.Tenancy.Enabled {
		if err := s.store.JoinDefaultTenant(ctx, account.ID, s.cfg.Tenancy.DefaultTenantID); err != nil {
			return RegisterResult{}, err
		}
	}

	if skipVerify {
		session, err := s.tokens.Issue(account, "")
		if err != nil {
			return RegisterResult{}, err
		}
		return RegisterResult{
			VerificationRequired: false,
			Account:              account,
			Session:              &session,
		}, nil
	}

	if err := s.sendVerificationEmail(ctx, account); err != nil {
		// Account exists; caller can resend.
		_ = err
	}

	return RegisterResult{
		VerificationRequired: true,
		Account:              account,
	}, nil
}

func (s *Service) Login(ctx context.Context, emailAddr, password string) (Session, error) {
	emailAddr = strings.TrimSpace(emailAddr)
	if !isValidEmail(emailAddr) || password == "" {
		return Session{}, ErrInvalidArgument
	}

	account, err := s.store.GetByEmail(ctx, emailAddr)
	if err != nil {
		if err == ErrNotFound {
			return Session{}, ErrUnauthenticated
		}
		return Session{}, err
	}
	if account.PasswordHash == nil || !CheckPassword(*account.PasswordHash, password) {
		return Session{}, ErrUnauthenticated
	}
	if !account.EmailVerified() {
		return Session{}, ErrEmailNotVerified
	}

	return s.tokens.Issue(account, "")
}

// IssueSession mints a JWT for an account, optionally overriding the configured default tenant id.
func (s *Service) IssueSession(ctx context.Context, accountID, tenantID string) (Session, error) {
	account, err := s.store.GetByID(ctx, accountID)
	if err != nil {
		if err == ErrNotFound {
			return Session{}, ErrUnauthenticated
		}
		return Session{}, err
	}
	return s.tokens.Issue(account, tenantID)
}

func (s *Service) SessionFromToken(ctx context.Context, accessToken string) (Session, error) {
	claims, err := s.tokens.Parse(accessToken)
	if err != nil {
		return Session{}, ErrUnauthenticated
	}

	account, err := s.store.GetByID(ctx, claims.AccountID)
	if err != nil {
		if err == ErrNotFound {
			return Session{}, ErrUnauthenticated
		}
		return Session{}, err
	}

	return Session{
		AccessToken: accessToken,
		Account:     account,
		TenantID:    claims.TenantID,
	}, nil
}

func (s *Service) LoginOAuth(
	ctx context.Context,
	provider, providerUserID, emailAddr string,
	firstName, lastName *string,
	avatarURL *string,
) (Session, error) {
	if !s.cfg.RegistrationEnabled() {
		return Session{}, ErrRegistrationDisabled
	}

	if provider == "" || providerUserID == "" || !isValidEmail(emailAddr) {
		return Session{}, ErrInvalidArgument
	}

	account, err := s.store.FindOrCreateOAuthAccount(ctx, provider, providerUserID, emailAddr, trimStringPtr(firstName), trimStringPtr(lastName), avatarURL)
	if err != nil {
		return Session{}, fmt.Errorf("oauth account: %w", err)
	}

	if s.cfg.Tenancy.Enabled {
		if err := s.store.JoinDefaultTenant(ctx, account.ID, s.cfg.Tenancy.DefaultTenantID); err != nil {
			return Session{}, err
		}
	}

	return s.tokens.Issue(account, "")
}

func (s *Service) VerifyEmail(ctx context.Context, plainToken string) (VerifyEmailResult, error) {
	plainToken = strings.TrimSpace(plainToken)
	if plainToken == "" {
		return VerifyEmailResult{}, ErrInvalidArgument
	}

	hash := HashVerificationToken(plainToken)
	accountID, err := s.store.ConsumeVerificationToken(ctx, hash, time.Now())
	if err == nil {
		account, err := s.store.GetByID(ctx, accountID)
		if err != nil {
			return VerifyEmailResult{}, err
		}
		return VerifyEmailResult{Email: account.Email}, nil
	}
	if err != ErrNotFound {
		return VerifyEmailResult{}, err
	}

	if email, ok := s.store.EmailByVerifiedToken(ctx, hash); ok {
		return VerifyEmailResult{Email: email, AlreadyVerified: true}, nil
	}
	return VerifyEmailResult{}, ErrInvalidArgument
}

func (s *Service) IsEmailVerified(ctx context.Context, emailAddr string) (bool, error) {
	emailAddr = strings.TrimSpace(strings.ToLower(emailAddr))
	if !isValidEmail(emailAddr) {
		return false, ErrInvalidArgument
	}
	return s.store.IsEmailVerified(ctx, emailAddr)
}

func (s *Service) ResendVerificationEmail(ctx context.Context, emailAddr string) error {
	emailAddr = strings.TrimSpace(emailAddr)
	if !isValidEmail(emailAddr) {
		return ErrInvalidArgument
	}

	account, err := s.store.GetByEmail(ctx, emailAddr)
	if err != nil {
		if err == ErrNotFound {
			return nil
		}
		return err
	}
	if account.PasswordHash == nil || account.EmailVerified() {
		return nil
	}

	return s.sendVerificationEmail(ctx, account)
}

func (s *Service) RequestPasswordReset(ctx context.Context, emailAddr string) error {
	emailAddr = strings.TrimSpace(emailAddr)
	if !isValidEmail(emailAddr) {
		return ErrInvalidArgument
	}

	account, err := s.store.GetByEmail(ctx, emailAddr)
	if err != nil {
		if err == ErrNotFound {
			return nil
		}
		return err
	}
	if account.PasswordHash == nil {
		return nil
	}

	plain, hash, err := NewVerificationToken()
	if err != nil {
		return err
	}

	expiresAt := time.Now().Add(s.cfg.PasswordResetTokenTTL())
	if err := s.store.CreatePasswordResetToken(ctx, account.ID, hash, expiresAt); err != nil {
		return err
	}

	resetURL := fmt.Sprintf(
		"%s/reset-password?token=%s",
		s.cfg.URLs.FrontendURL,
		plain,
	)
	return s.mailer.SendPasswordResetEmail(ctx, account.Email, resetURL)
}

func (s *Service) ResetPassword(ctx context.Context, plainToken, newPassword string) error {
	plainToken = strings.TrimSpace(plainToken)
	if plainToken == "" || len(newPassword) < 6 {
		return ErrInvalidArgument
	}

	hash := HashVerificationToken(plainToken)
	accountID, err := s.store.ConsumePasswordResetToken(ctx, hash, time.Now())
	if err != nil {
		if err == ErrNotFound {
			return ErrInvalidArgument
		}
		return err
	}

	passwordHash, err := HashPassword(newPassword)
	if err != nil {
		return err
	}

	return s.store.UpdatePasswordHash(ctx, accountID, passwordHash)
}

func (s *Service) sendVerificationEmail(ctx context.Context, account Account) error {
	plain, hash, err := NewVerificationToken()
	if err != nil {
		return err
	}

	expiresAt := time.Now().Add(s.cfg.VerificationTokenTTL())
	if err := s.store.CreateVerificationToken(ctx, account.ID, hash, expiresAt); err != nil {
		return err
	}

	verifyURL := fmt.Sprintf(
		"%s/account/verify-email?token=%s",
		s.cfg.URLs.APIPublicURL,
		plain,
	)
	return s.mailer.SendVerificationEmail(ctx, account.Email, verifyURL)
}

func isValidEmail(email string) bool {
	return strings.Contains(email, "@") && strings.Contains(email, ".")
}
