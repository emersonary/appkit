package tenants

import (
	"os"
	"strings"
)

const defaultAppConfigPath = "config/tenants.yaml"

// AppConfig is the application-level tenants block config (main YAML tenants node).
type AppConfig struct {
	Enabled        bool        `mapstructure:"enabled" json:"enabled"`
	ConfigPath     string      `mapstructure:"config_path" json:"config_path,omitempty"`
	JWTSecret      string      `mapstructure:"jwt_secret" json:"-"`
	Schema         string      `mapstructure:"schema" json:"schema,omitempty"`
	Mode           string      `mapstructure:"mode" json:"mode,omitempty"`
	InviteTokenTTL string      `mapstructure:"invite_token_ttl" json:"invite_token_ttl,omitempty"`
	Feed           []FeedEntry `mapstructure:"feed" json:"feed,omitempty"`
}

func (c *AppConfig) ApplyDefaults() {
	if c.ConfigPath == "" && len(c.Feed) == 0 {
		c.ConfigPath = defaultAppConfigPath
	}
}

func (c AppConfig) jwtSecret() string {
	return c.JWTSecret
}

// ResolveBlockConfig returns inline tenants config or loads a legacy external file.
func ResolveBlockConfig(app AppConfig) (Config, error) {
	app.ApplyDefaults()

	path := strings.TrimSpace(app.ConfigPath)
	if path != "" {
		if _, err := os.Stat(path); err == nil {
			return LoadConfig(path)
		}
	}

	cfg := Config{
		Schema:         app.Schema,
		Mode:           app.Mode,
		InviteTokenTTL: app.InviteTokenTTL,
		Feed:           app.Feed,
	}
	cfg.normalize()
	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}
