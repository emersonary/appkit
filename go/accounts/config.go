package accounts

import (
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Schema   string         `yaml:"schema"`
	Tenancy  TenancyConfig  `yaml:"tenancy"`
	Session  SessionConfig  `yaml:"session"`
	URLs     URLsConfig     `yaml:"urls"`
	Tokens   TokensConfig   `yaml:"tokens"`
	OAuth    OAuthConfig    `yaml:"oauth"`
	Features FeaturesConfig `yaml:"features"`
}

type TenancyConfig struct {
	Enabled         bool   `yaml:"enabled"`
	DefaultTenantID string `yaml:"default_tenant_id"`
}

type SessionConfig struct {
	AccessTokenTTL string `yaml:"access_token_ttl"`
}

type URLsConfig struct {
	FrontendURL  string `yaml:"frontend_url"`
	APIPublicURL string `yaml:"api_public_url"`
}

type TokensConfig struct {
	VerificationTTL   string `yaml:"verification_token_ttl"`
	PasswordResetTTL  string `yaml:"password_reset_token_ttl"`
}

type OAuthConfig struct {
	StateCookieName string       `yaml:"state_cookie_name"`
	Google          GoogleConfig `yaml:"google"`
}

type GoogleConfig struct {
	Enabled     bool   `yaml:"enabled"`
	ClientID    string `yaml:"client_id"`
	ClientSecret string `yaml:"client_secret"`
	RedirectURL string `yaml:"redirect_url"`
}

type FeaturesConfig struct {
	AdminFlag bool `yaml:"admin_flag"`
}

// Secrets holds runtime credentials not stored in YAML files.
type Secrets struct {
	JWTSecret          string
	GoogleClientID     string
	GoogleClientSecret string
}

func LoadConfig(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, wrapErr(ErrLoadConfig, "read", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Config{}, wrapErr(ErrLoadConfig, "parse", err)
	}

	cfg.normalize()
	if err := cfg.Validate(); err != nil {
		return errConfig(cfg, err)
	}
	return cfg, nil
}

func errConfig(cfg Config, err error) (Config, error) {
	return cfg, err
}

func (c *Config) normalize() {
	c.Schema = strings.TrimSpace(c.Schema)
	c.Tenancy.DefaultTenantID = strings.TrimSpace(c.Tenancy.DefaultTenantID)
	c.Session.AccessTokenTTL = strings.TrimSpace(c.Session.AccessTokenTTL)
	if c.Session.AccessTokenTTL == "" {
		c.Session.AccessTokenTTL = "24h"
	}
	c.Tokens.VerificationTTL = strings.TrimSpace(c.Tokens.VerificationTTL)
	if c.Tokens.VerificationTTL == "" {
		c.Tokens.VerificationTTL = "24h"
	}
	c.Tokens.PasswordResetTTL = strings.TrimSpace(c.Tokens.PasswordResetTTL)
	if c.Tokens.PasswordResetTTL == "" {
		c.Tokens.PasswordResetTTL = "1h"
	}
	c.OAuth.StateCookieName = strings.TrimSpace(c.OAuth.StateCookieName)
	if c.OAuth.StateCookieName == "" {
		c.OAuth.StateCookieName = "appkit_oauth_state"
	}
	c.URLs.FrontendURL = stringsTrimRightSlash(c.URLs.FrontendURL)
	c.URLs.APIPublicURL = stringsTrimRightSlash(c.URLs.APIPublicURL)
}

func (c Config) Validate() error {
	if c.Schema == "" {
		return ErrSchemaRequired
	}
	if err := validateIdent(c.Schema); err != nil {
		return ErrInvalidSchema.With("schema", err.Error())
	}
	if !c.Tenancy.Enabled && c.Tenancy.DefaultTenantID == "" {
		return ErrDefaultTenantRequired
	}
	if c.Tenancy.Enabled && c.Tenancy.DefaultTenantID == "" {
		return ErrDefaultTenantRequired
	}
	if _, err := time.ParseDuration(c.Session.AccessTokenTTL); err != nil {
		return ErrInvalidArgument.With("session.access_token_ttl", err.Error())
	}
	if _, err := time.ParseDuration(c.Tokens.VerificationTTL); err != nil {
		return ErrInvalidArgument.With("tokens.verification_token_ttl", err.Error())
	}
	if _, err := time.ParseDuration(c.Tokens.PasswordResetTTL); err != nil {
		return ErrInvalidArgument.With("tokens.password_reset_token_ttl", err.Error())
	}
	return nil
}

func (c Config) AccessTokenTTL() time.Duration {
	d, _ := time.ParseDuration(c.Session.AccessTokenTTL)
	return d
}

func (c Config) VerificationTokenTTL() time.Duration {
	d, _ := time.ParseDuration(c.Tokens.VerificationTTL)
	return d
}

func (c Config) PasswordResetTokenTTL() time.Duration {
	d, _ := time.ParseDuration(c.Tokens.PasswordResetTTL)
	return d
}

func (g GoogleConfig) EnabledWithSecrets(secrets Secrets) bool {
	if !g.Enabled {
		return false
	}
	clientID := strings.TrimSpace(secrets.GoogleClientID)
	if clientID == "" {
		clientID = strings.TrimSpace(g.ClientID)
	}
	clientSecret := strings.TrimSpace(secrets.GoogleClientSecret)
	if clientSecret == "" {
		clientSecret = strings.TrimSpace(g.ClientSecret)
	}
	return clientID != "" && clientSecret != "" && strings.TrimSpace(g.RedirectURL) != ""
}

func (c Config) EffectiveTenantID() string {
	return c.Tenancy.DefaultTenantID
}

func stringsTrimRightSlash(value string) string {
	for len(value) > 0 && value[len(value)-1] == '/' {
		value = value[:len(value)-1]
	}
	return value
}
