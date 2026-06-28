package currency

import (
	"os"
	"strings"
	"time"
)

// AppConfig is the application-level currency block config (main YAML currency node).
type AppConfig struct {
	Enabled        bool          `mapstructure:"enabled" json:"enabled"`
	ConfigPath     string        `mapstructure:"config_path" json:"config_path,omitempty"`
	Schema         string        `mapstructure:"schema" json:"schema"`
	BaseCurrency   string        `mapstructure:"base_currency" json:"base_currency"`
	Currencies     []string      `mapstructure:"currencies" json:"currencies"`
	SkipSeed       bool          `mapstructure:"skip_seed" json:"skip_seed"`
	APIURL         string        `mapstructure:"api_url" json:"api_url"`
	UpdateInterval time.Duration `mapstructure:"update_interval" json:"update_interval"`
}

// ApplyDefaults fills zero values for optional app-level fields.
func (c *AppConfig) ApplyDefaults() {
	if c.APIURL == "" {
		c.APIURL = DefaultAPIURL
	}
	if c.UpdateInterval == 0 {
		c.UpdateInterval = time.Hour
	}
	if strings.TrimSpace(c.BaseCurrency) == "" {
		c.BaseCurrency = BaseCurrencyCode
	}
}

func (c AppConfig) BlockConfig() (Config, error) {
	c.ApplyDefaults()

	cfg := Config{
		Schema:       c.Schema,
		BaseCurrency: c.BaseCurrency,
		Currencies:   append([]string(nil), c.Currencies...),
		SkipSeed:     c.SkipSeed,
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
	if app.BaseCurrency != "" {
		block.BaseCurrency = NormalizeCode(app.BaseCurrency)
	}
	if len(app.Currencies) > 0 {
		block.Currencies = normalizeCodeList(app.Currencies)
	}
	block.SkipSeed = app.SkipSeed || block.SkipSeed
}
