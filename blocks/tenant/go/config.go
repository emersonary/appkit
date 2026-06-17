package tenants

import (
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Schema         string `yaml:"schema"`
	InviteTokenTTL string `yaml:"invite_token_ttl"`
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
		return Config{}, err
	}
	return cfg, nil
}

func (c *Config) normalize() {
	c.Schema = strings.TrimSpace(c.Schema)
	c.InviteTokenTTL = strings.TrimSpace(c.InviteTokenTTL)
	if c.InviteTokenTTL == "" {
		c.InviteTokenTTL = "168h"
	}
}

func (c Config) Validate() error {
	if c.Schema == "" {
		return ErrSchemaRequired
	}
	if err := validateIdent(c.Schema); err != nil {
		return ErrInvalidSchema.With("schema", err.Error())
	}
	if _, err := time.ParseDuration(c.InviteTokenTTL); err != nil {
		return ErrInvalidArgument.With("invite_token_ttl", err.Error())
	}
	return nil
}

func (c Config) inviteTokenTTL() time.Duration {
	d, _ := time.ParseDuration(c.InviteTokenTTL)
	return d
}
