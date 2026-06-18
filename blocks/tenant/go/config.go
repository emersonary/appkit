package tenants

import (
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

const (
	ModeDynamic = "dynamic"
	ModeFixed   = "fixed"
)

// FeedEntry is a fixed-catalog tenant (feed) declared in tenants YAML.
type FeedEntry struct {
	ID       string         `yaml:"id"`
	Name     string         `yaml:"name"`
	Timezone string         `yaml:"timezone"`
	Metadata map[string]any `yaml:"metadata"`
}

type Config struct {
	Schema         string      `yaml:"schema"`
	Mode           string      `yaml:"mode"`
	InviteTokenTTL string      `yaml:"invite_token_ttl"`
	Feed           []FeedEntry `yaml:"feed"`
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
	c.Mode = strings.TrimSpace(strings.ToLower(c.Mode))
	if c.Mode == "" {
		c.Mode = ModeDynamic
	}
	c.InviteTokenTTL = strings.TrimSpace(c.InviteTokenTTL)
	if c.InviteTokenTTL == "" {
		c.InviteTokenTTL = "168h"
	}

	for i := range c.Feed {
		c.Feed[i].ID = strings.TrimSpace(strings.ToLower(c.Feed[i].ID))
		c.Feed[i].Name = strings.TrimSpace(c.Feed[i].Name)
		c.Feed[i].Timezone = strings.TrimSpace(c.Feed[i].Timezone)
		if c.Feed[i].Timezone == "" {
			c.Feed[i].Timezone = "UTC"
		}
		if c.Feed[i].Metadata == nil {
			c.Feed[i].Metadata = map[string]any{}
		}
	}
}

func (c Config) Validate() error {
	if c.Schema == "" {
		return ErrSchemaRequired
	}
	if err := validateIdent(c.Schema); err != nil {
		return ErrInvalidSchema.With("schema", err.Error())
	}
	if c.Mode != ModeDynamic && c.Mode != ModeFixed {
		return ErrInvalidArgument.With("mode", c.Mode)
	}
	if _, err := time.ParseDuration(c.InviteTokenTTL); err != nil {
		return ErrInvalidArgument.With("invite_token_ttl", err.Error())
	}

	if c.Mode == ModeFixed {
		if len(c.Feed) == 0 {
			return ErrFeedCatalogRequired
		}
		seen := make(map[string]struct{}, len(c.Feed))
		for _, entry := range c.Feed {
			if entry.ID == "" {
				return ErrInvalidArgument.With("feed.id", "required")
			}
			if err := validateFeedID(entry.ID); err != nil {
				return ErrInvalidArgument.With("feed.id", err.Error())
			}
			if entry.Name == "" {
				return ErrInvalidArgument.With("feed.name", "required")
			}
			if _, dup := seen[entry.ID]; dup {
				return ErrInvalidArgument.With("feed.id", "duplicate "+entry.ID)
			}
			seen[entry.ID] = struct{}{}
		}
	}

	return nil
}

func (c Config) IsFixedMode() bool {
	return c.Mode == ModeFixed
}

func (c Config) CatalogIDs() []string {
	ids := make([]string, 0, len(c.Feed))
	for _, entry := range c.Feed {
		ids = append(ids, entry.ID)
	}
	return ids
}

func (c Config) HasCatalogID(id string) bool {
	id = strings.TrimSpace(strings.ToLower(id))
	for _, entry := range c.Feed {
		if entry.ID == id {
			return true
		}
	}
	return false
}

func (c Config) inviteTokenTTL() time.Duration {
	d, _ := time.ParseDuration(c.InviteTokenTTL)
	return d
}
