package accounts

import (
	"os"
	"strings"
	"time"
)

// AppConfig is the application-level accounts block config (main YAML accounts node).
type AppConfig struct {
	Enabled               bool           `mapstructure:"enabled" json:"enabled"`
	ConfigPath            string         `mapstructure:"config_path" json:"config_path,omitempty"`
	Schema                string         `mapstructure:"schema" json:"schema"`
	Tenancy               TenancyConfig  `mapstructure:"tenancy" json:"tenancy"`
	JWTSecret             string         `mapstructure:"jwt_secret" json:"-"`
	AccessTokenTTL        time.Duration  `mapstructure:"access_token_ttl" json:"access_token_ttl"`
	FrontendURL           string         `mapstructure:"frontend_url" json:"frontend_url"`
	APIPublicURL          string         `mapstructure:"api_public_url" json:"api_public_url"`
	VerificationTokenTTL  time.Duration  `mapstructure:"verification_token_ttl" json:"verification_token_ttl"`
	PasswordResetTokenTTL time.Duration  `mapstructure:"password_reset_token_ttl" json:"password_reset_token_ttl"`
	RegistrationEnabled   *bool          `mapstructure:"registration_enabled" json:"registration_enabled,omitempty"`
	RegisterAsAdmin       bool           `mapstructure:"register_as_admin" json:"register_as_admin"`
	SkipEmailVerification bool           `mapstructure:"skip_email_verification" json:"skip_email_verification"`
	OAuth                 AppOAuthConfig `mapstructure:"oauth" json:"oauth"`
}

// AppOAuthConfig is the application-level OAuth settings (main YAML accounts.oauth node).
type AppOAuthConfig struct {
	Enabled         *bool                `mapstructure:"enabled" json:"enabled,omitempty"`
	StateCookieName string               `mapstructure:"state_cookie_name" json:"state_cookie_name"`
	Google          AppGoogleOAuthConfig `mapstructure:"google" json:"google"`
	Facebook        ProviderToggle       `mapstructure:"facebook" json:"facebook"`
	Apple           ProviderToggle       `mapstructure:"apple" json:"apple"`
}

// AppGoogleOAuthConfig holds Google OAuth settings and secrets from application config.
type AppGoogleOAuthConfig struct {
	Enabled      bool   `mapstructure:"enabled" json:"enabled"`
	ClientID     string `mapstructure:"client_id" json:"client_id"`
	ClientSecret string `mapstructure:"client_secret" json:"-"`
	RedirectURL  string `mapstructure:"redirect_url" json:"redirect_url"`
}

// ApplyDefaults fills zero values for optional app-level fields.
func (c *AppConfig) ApplyDefaults() {
	if c.RegistrationEnabled == nil {
		enabled := true
		c.RegistrationEnabled = &enabled
	}
	if strings.TrimSpace(c.Schema) == "" {
		c.Schema = "account"
	}
}

func (c AppConfig) secrets() Secrets {
	return Secrets{
		JWTSecret:          c.JWTSecret,
		GoogleClientID:     c.OAuth.Google.ClientID,
		GoogleClientSecret: c.OAuth.Google.ClientSecret,
	}
}

// BlockConfig builds the runtime block config from application YAML.
func (c AppConfig) BlockConfig() (Config, error) {
	c.ApplyDefaults()

	enabled := c.Enabled
	cfg := Config{
		Enabled: &enabled,
		Schema:  c.Schema,
		Tenancy: c.Tenancy,
		URLs: URLsConfig{
			FrontendURL:  c.FrontendURL,
			APIPublicURL: c.APIPublicURL,
		},
		OAuth: OAccountConfig{
			Enabled:         c.OAuth.Enabled,
			StateCookieName: c.OAuth.StateCookieName,
			Google: GoogleConfig{
				Enabled:      c.OAuth.Google.Enabled,
				ClientID:     c.OAuth.Google.ClientID,
				ClientSecret: c.OAuth.Google.ClientSecret,
				RedirectURL:  c.OAuth.Google.RedirectURL,
			},
			Facebook: c.OAuth.Facebook,
			Apple:    c.OAuth.Apple,
		},
		Features: FeaturesConfig{
			RegistrationEnabled:   c.RegistrationEnabled,
			RegisterAsAdmin:       c.RegisterAsAdmin,
			SkipEmailVerification: c.SkipEmailVerification,
		},
	}

	if c.AccessTokenTTL > 0 {
		cfg.Session.AccessTokenTTL = c.AccessTokenTTL.String()
	}
	if c.VerificationTokenTTL > 0 {
		cfg.Tokens.VerificationTTL = c.VerificationTokenTTL.String()
	}
	if c.PasswordResetTokenTTL > 0 {
		cfg.Tokens.PasswordResetTTL = c.PasswordResetTokenTTL.String()
	}

	cfg.normalize()
	if err := cfg.Validate(); err != nil {
		return cfg, err
	}
	return cfg, nil
}

// ResolveBlockConfig returns the effective block config.
// When config_path points at a legacy block YAML file, that file is loaded and app fields override it.
func ResolveBlockConfig(app AppConfig) (Config, error) {
	app.ApplyDefaults()

	path := strings.TrimSpace(app.ConfigPath)
	if path != "" {
		if _, err := os.Stat(path); err == nil {
			blockCfg, err := LoadConfig(path)
			if err != nil {
				return Config{}, err
			}
			mergeAppConfig(&blockCfg, app)
			return blockCfg, nil
		}
	}

	return app.BlockConfig()
}

func mergeAppConfig(block *Config, app AppConfig) {
	if app.Schema != "" {
		block.Schema = app.Schema
	}
	if app.Tenancy.DefaultTenantID != "" || app.Tenancy.Enabled {
		block.Tenancy = app.Tenancy
	}
	if app.FrontendURL != "" {
		block.URLs.FrontendURL = app.FrontendURL
	}
	if app.APIPublicURL != "" {
		block.URLs.APIPublicURL = app.APIPublicURL
	}
	if app.AccessTokenTTL > 0 {
		block.Session.AccessTokenTTL = app.AccessTokenTTL.String()
	}
	if app.VerificationTokenTTL > 0 {
		block.Tokens.VerificationTTL = app.VerificationTokenTTL.String()
	}
	if app.PasswordResetTokenTTL > 0 {
		block.Tokens.PasswordResetTTL = app.PasswordResetTokenTTL.String()
	}
	if app.OAuth.Enabled != nil {
		block.OAuth.Enabled = app.OAuth.Enabled
	}
	if name := strings.TrimSpace(app.OAuth.StateCookieName); name != "" {
		block.OAuth.StateCookieName = name
	}
	if app.OAuth.Google.Enabled {
		block.OAuth.Google.Enabled = true
	}
	if id := strings.TrimSpace(app.OAuth.Google.ClientID); id != "" {
		block.OAuth.Google.ClientID = id
	}
	if secret := strings.TrimSpace(app.OAuth.Google.ClientSecret); secret != "" {
		block.OAuth.Google.ClientSecret = secret
	}
	if redirect := strings.TrimSpace(app.OAuth.Google.RedirectURL); redirect != "" {
		block.OAuth.Google.RedirectURL = redirect
	}
	block.OAuth.Facebook.Enabled = app.OAuth.Facebook.Enabled || block.OAuth.Facebook.Enabled
	block.OAuth.Apple.Enabled = app.OAuth.Apple.Enabled || block.OAuth.Apple.Enabled
	if app.RegistrationEnabled != nil {
		block.Features.RegistrationEnabled = app.RegistrationEnabled
	}
	block.Features.RegisterAsAdmin = app.RegisterAsAdmin || block.Features.RegisterAsAdmin
	block.Features.SkipEmailVerification = app.SkipEmailVerification || block.Features.SkipEmailVerification
}
