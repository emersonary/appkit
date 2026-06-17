package tenants

const defaultAppConfigPath = "config/tenants.yaml"

// AppConfig is the application-level tenants block config (main YAML tenants node).
type AppConfig struct {
	Enabled    bool   `mapstructure:"enabled" json:"enabled"`
	ConfigPath string `mapstructure:"config_path" json:"config_path"`
	JWTSecret  string `mapstructure:"jwt_secret" json:"-"`
}

func (c *AppConfig) ApplyDefaults() {
	if c.ConfigPath == "" {
		c.ConfigPath = defaultAppConfigPath
	}
}

func (c AppConfig) jwtSecret() string {
	return c.JWTSecret
}
