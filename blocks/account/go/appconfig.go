package accounts

import (
	"strings"
	"time"
)

const defaultAppConfigPath = "config/accounts.yaml"

// AppConfig is the application-level accounts block config (main YAML accounts node).
type AppConfig struct {
	Enabled               bool          `mapstructure:"enabled" json:"enabled"`
	ConfigPath            string        `mapstructure:"config_path" json:"config_path"`
	JWTSecret             string        `mapstructure:"jwt_secret" json:"-"`
	AccessTokenTTL        time.Duration `mapstructure:"access_token_ttl" json:"access_token_ttl"`
	FrontendURL           string        `mapstructure:"frontend_url" json:"frontend_url"`
	APIPublicURL          string        `mapstructure:"api_public_url" json:"api_public_url"`
	VerificationTokenTTL  time.Duration `mapstructure:"verification_token_ttl" json:"verification_token_ttl"`
	PasswordResetTokenTTL time.Duration `mapstructure:"password_reset_token_ttl" json:"password_reset_token_ttl"`
	Google                GoogleSecrets `mapstructure:"google" json:"google"`
}

// GoogleSecrets holds OAuth client credentials from the application config.
type GoogleSecrets struct {
	ClientID     string `mapstructure:"client_id" json:"client_id"`
	ClientSecret string `mapstructure:"client_secret" json:"-"`
}

// ApplyDefaults fills zero values for optional app-level fields.
func (c *AppConfig) ApplyDefaults() {
	if c.ConfigPath == "" {
		c.ConfigPath = defaultAppConfigPath
	}
}

func (c AppConfig) secrets() Secrets {
	return Secrets{
		JWTSecret:          c.JWTSecret,
		GoogleClientID:     c.Google.ClientID,
		GoogleClientSecret: c.Google.ClientSecret,
	}
}

func mergeAppConfig(block *Config, app AppConfig) {
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
	if id := strings.TrimSpace(app.Google.ClientID); id != "" {
		block.OAuth.Google.ClientID = id
	}
	if secret := strings.TrimSpace(app.Google.ClientSecret); secret != "" {
		block.OAuth.Google.ClientSecret = secret
	}
}
