package social

import (
	"fmt"
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

const (
	defaultGraphAPIBaseURL      = "https://graph.facebook.com/v21.0"
	defaultPinterestAPIBaseURL  = "https://api.pinterest.com/v5"
	defaultTikTokAPIBaseURL     = "https://open.tiktokapis.com"
	defaultLinkedInAPIBaseURL   = "https://api.linkedin.com"
	defaultYouTubeAPIBaseURL    = "https://www.googleapis.com/youtube/v3"
	defaultPlatformTimeout      = 30 * time.Second
	defaultYouTubePostType      = "community"
)

// PlatformConfig configures one social platform client.
type PlatformConfig struct {
	Enabled        *bool         `mapstructure:"enabled" yaml:"enabled" json:"enabled,omitempty"`
	Dispatch       string        `mapstructure:"dispatch" yaml:"dispatch" json:"dispatch,omitempty"`
	Driver         string        `mapstructure:"driver" yaml:"driver" json:"driver,omitempty"`
	AccountID      string        `mapstructure:"account_id" yaml:"account_id" json:"account_id,omitempty"`
	PageID         string        `mapstructure:"page_id" yaml:"page_id" json:"page_id,omitempty"`
	BoardID        string        `mapstructure:"board_id" yaml:"board_id" json:"board_id,omitempty"`
	BoardName      string        `mapstructure:"board_name" yaml:"board_name" json:"board_name,omitempty"`
	AccessToken    string        `mapstructure:"access_token" yaml:"access_token" json:"access_token,omitempty"`
	AccessTokenEnv string        `mapstructure:"access_token_env" yaml:"access_token_env" json:"access_token_env,omitempty"`
	OAuth          bool          `mapstructure:"oauth" yaml:"oauth" json:"oauth,omitempty"`
	APIBaseURL     string        `mapstructure:"api_base_url" yaml:"api_base_url" json:"api_base_url,omitempty"`
	Timeout        time.Duration `mapstructure:"timeout" yaml:"timeout" json:"timeout,omitempty"`
	YouTubePostType string       `mapstructure:"youtube_post_type" yaml:"youtube_post_type" json:"youtube_post_type,omitempty"`
}

func (p PlatformConfig) isEnabled() bool {
	if p.Enabled == nil {
		return true
	}
	return *p.Enabled
}

func (p PlatformConfig) resolvedDispatch(defaultMode DispatchMode) (DispatchMode, error) {
	if strings.TrimSpace(p.Dispatch) == "" {
		return defaultMode, nil
	}
	return ParseDispatchMode(p.Dispatch)
}

func (p PlatformConfig) driver(platformID PlatformID) string {
	if d := strings.TrimSpace(p.Driver); d != "" {
		return d
	}
	meta, ok := PlatformMetaFor(platformID)
	if !ok {
		return ""
	}
	return meta.DefaultDriver
}

func (p *PlatformConfig) normalize(platformID PlatformID) {
	p.Driver = p.driver(platformID)
	if p.Timeout <= 0 {
		p.Timeout = defaultPlatformTimeout
	}
	if strings.TrimSpace(p.APIBaseURL) == "" {
		switch p.Driver {
		case "instagram", "facebook", "threads":
			p.APIBaseURL = defaultGraphAPIBaseURL
		case "pinterest":
			p.APIBaseURL = defaultPinterestAPIBaseURL
		case "tiktok":
			p.APIBaseURL = defaultTikTokAPIBaseURL
		case "linkedin":
			p.APIBaseURL = defaultLinkedInAPIBaseURL
		case "youtube":
			p.APIBaseURL = defaultYouTubeAPIBaseURL
		}
	}
	if strings.TrimSpace(p.YouTubePostType) == "" {
		p.YouTubePostType = defaultYouTubePostType
	}
}

func (p PlatformConfig) resolvedAccessToken() string {
	if token := strings.TrimSpace(p.AccessToken); token != "" {
		return token
	}
	if env := strings.TrimSpace(p.AccessTokenEnv); env != "" {
		return strings.TrimSpace(os.Getenv(env))
	}
	return ""
}

// SocialConfig is the social block configuration.
type SocialConfig struct {
	Enabled         bool                      `mapstructure:"enabled" yaml:"enabled" json:"enabled"`
	ConfigPath      string                    `mapstructure:"config_path" yaml:"config_path,omitempty" json:"config_path,omitempty"`
	DefaultDispatch string                    `mapstructure:"default_dispatch" yaml:"default_dispatch" json:"default_dispatch,omitempty"`
	Platforms       map[string]PlatformConfig `mapstructure:"platforms" yaml:"platforms" json:"platforms"`
}

// Resolved returns a normalized, validated copy.
func (c SocialConfig) Resolved() (SocialConfig, error) {
	out := c
	out.normalize()
	if err := out.Validate(); err != nil {
		return SocialConfig{}, err
	}
	return out, nil
}

// ResolveBlockConfig loads an external file when config_path is set.
func ResolveBlockConfig(input SocialConfig) (SocialConfig, error) {
	path := strings.TrimSpace(input.ConfigPath)
	if path != "" {
		if _, err := os.Stat(path); err == nil {
			fileCfg, err := LoadConfig(path)
			if err != nil {
				return SocialConfig{}, err
			}
			mergeSocialConfig(&fileCfg, input)
			fileCfg.normalize()
			if err := fileCfg.Validate(); err != nil {
				return SocialConfig{}, err
			}
			return fileCfg, nil
		}
	}
	return input.Resolved()
}

func mergeSocialConfig(base *SocialConfig, overlay SocialConfig) {
	if overlay.Enabled {
		base.Enabled = overlay.Enabled
	}
	if strings.TrimSpace(overlay.DefaultDispatch) != "" {
		base.DefaultDispatch = overlay.DefaultDispatch
	}
	if len(overlay.Platforms) > 0 {
		base.Platforms = copyPlatforms(overlay.Platforms)
	}
}

func LoadConfig(path string) (SocialConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return SocialConfig{}, wrapErr(ErrLoadConfig, "path", fmt.Errorf("%s: %w", path, err))
	}

	var cfg SocialConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return SocialConfig{}, wrapErr(ErrLoadConfig, "parse", err)
	}

	cfg.normalize()
	if err := cfg.Validate(); err != nil {
		return SocialConfig{}, err
	}
	return cfg, nil
}

func (c *SocialConfig) normalize() {
	if c.Platforms == nil {
		c.Platforms = map[string]PlatformConfig{}
	}
	if strings.TrimSpace(c.DefaultDispatch) == "" {
		c.DefaultDispatch = string(DispatchServer)
	}
	for id := range c.Platforms {
		platformID := PlatformID(id)
		p := c.Platforms[id]
		p.normalize(platformID)
		c.Platforms[id] = p
	}
}

func (c SocialConfig) defaultDispatch() (DispatchMode, error) {
	return ParseDispatchMode(c.DefaultDispatch)
}

func (c SocialConfig) Validate() error {
	if !c.Enabled {
		return nil
	}

	if len(c.Platforms) == 0 {
		return invalidConfig("platforms", "at least one platform is required when social is enabled")
	}

	if _, err := c.defaultDispatch(); err != nil {
		return err
	}

	defaultDispatch, err := c.defaultDispatch()
	if err != nil {
		return err
	}

	enabledCount := 0
	for rawID, platform := range c.Platforms {
		platformID, err := ParsePlatformID(rawID)
		if err != nil {
			return err
		}
		if !platform.isEnabled() {
			continue
		}
		enabledCount++

		if _, err := platform.resolvedDispatch(defaultDispatch); err != nil {
			return invalidConfigf("platforms.%s.dispatch", rawID, "%s", err.Error())
		}

		driver := platform.driver(platformID)
		switch driver {
		case "instagram", "facebook", "threads":
			if strings.TrimSpace(platform.AccountID) == "" {
				return invalidConfigf("platforms.%s.account_id", rawID, "account_id is required for %s", driver)
			}
		case "pinterest":
			if strings.TrimSpace(platform.BoardID) == "" {
				return invalidConfigf("platforms.%s.board_id", rawID, "board_id is required for pinterest")
			}
		case "tiktok", "linkedin", "youtube":
			if !platform.OAuth && strings.TrimSpace(platform.AccountID) == "" {
				return invalidConfigf("platforms.%s.account_id", rawID, "account_id is required for %s", driver)
			}
		default:
			return invalidConfigf("platforms.%s.driver", rawID, "unsupported driver %q", driver)
		}

		if !platform.OAuth && platform.resolvedAccessToken() == "" {
			envHint := strings.TrimSpace(platform.AccessTokenEnv)
			if envHint == "" {
				envHint = "access_token"
			}
			return invalidConfigf("platforms.%s.access_token", rawID, "access_token or %s env var is required", envHint)
		}
	}

	if enabledCount == 0 {
		return invalidConfig("platforms", "at least one enabled platform is required when social is enabled")
	}

	return nil
}

// ValidateDefaults checks global platform templates (drivers, dispatch) without requiring credentials.
// Use for application-level social config; per-project secrets are validated with Validate.
func (c SocialConfig) ValidateDefaults() error {
	if !c.Enabled {
		return nil
	}

	if _, err := c.defaultDispatch(); err != nil {
		return err
	}

	defaultDispatch, err := c.defaultDispatch()
	if err != nil {
		return err
	}

	for rawID, platform := range c.Platforms {
		platformID, err := ParsePlatformID(rawID)
		if err != nil {
			return err
		}
		if !platform.isEnabled() {
			continue
		}
		if _, err := platform.resolvedDispatch(defaultDispatch); err != nil {
			return invalidConfigf("platforms.%s.dispatch", rawID, "%s", err.Error())
		}
		driver := platform.driver(platformID)
		switch driver {
		case "instagram", "facebook", "threads", "pinterest", "tiktok", "linkedin", "youtube":
		default:
			return invalidConfigf("platforms.%s.driver", rawID, "unsupported driver %q", driver)
		}
	}

	return nil
}

func copyPlatforms(in map[string]PlatformConfig) map[string]PlatformConfig {
	out := make(map[string]PlatformConfig, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}
