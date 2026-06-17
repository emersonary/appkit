package currency

import (
	"os"
	"time"
)

const defaultAppConfigPath = "config/currency.yaml"

// AppConfig is the application-level currency block config (main YAML currency node).
type AppConfig struct {
	Enabled        bool          `mapstructure:"enabled" json:"enabled"`
	ConfigPath     string        `mapstructure:"config_path" json:"config_path"`
	APIURL         string        `mapstructure:"api_url" json:"api_url"`
	UpdateInterval time.Duration `mapstructure:"update_interval" json:"update_interval"`
}

// ApplyDefaults fills zero values for optional app-level fields.
func (c *AppConfig) ApplyDefaults() {
	if c.ConfigPath == "" {
		c.ConfigPath = defaultAppConfigPath
	}
	if c.APIURL == "" {
		c.APIURL = DefaultAPIURL
	}
	if c.UpdateInterval == 0 {
		c.UpdateInterval = time.Hour
	}
}

func (c AppConfig) blockConfigPath() string {
	path := c.ConfigPath
	if path == "" {
		path = defaultAppConfigPath
	}
	if override := os.Getenv("CURRENCY_CONFIG"); override != "" {
		path = override
	}
	return path
}
