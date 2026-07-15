package dbhist

// AppConfig is the application-level dbhist block config (main YAML dbhist node).
// Only Enabled is configurable; AUDIT/REPO selection comes from table comments.
type AppConfig struct {
	Enabled bool `mapstructure:"enabled" json:"enabled"`
}

// BlockConfig builds a Config from inline AppConfig fields.
func (c AppConfig) BlockConfig() (Config, error) {
	cfg := Config{}
	if err := cfg.Validate(); err != nil {
		return cfg, err
	}
	return cfg, nil
}

// ResolveBlockConfig returns the effective dbhist config from the main YAML block.
func ResolveBlockConfig(app AppConfig) (Config, error) {
	return app.BlockConfig()
}
