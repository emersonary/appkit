package messenger

// AppConfig is the application-level messenger block config (main YAML messenger node).
type AppConfig struct {
	Enabled       bool   `mapstructure:"enabled" json:"enabled"`
	Schema        string `mapstructure:"schema" json:"schema,omitempty"`
	AccountSchema string `mapstructure:"account_schema" json:"account_schema,omitempty"`
	TenantSchema  string `mapstructure:"tenant_schema" json:"tenant_schema,omitempty"`
	AgentAPIKey   string `mapstructure:"agent_api_key" json:"-"`
}

func (c AppConfig) blockConfig() Config {
	cfg := Config{
		Schema:        c.Schema,
		AccountSchema: c.AccountSchema,
		TenantSchema:  c.TenantSchema,
		AgentAPIKey:   c.AgentAPIKey,
	}
	cfg.normalize()
	return cfg
}

func ResolveBlockConfig(app AppConfig) (Config, error) {
	cfg := app.blockConfig()
	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}
