package ai

import (
	"fmt"
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

const (
	defaultOpenAIBaseURL   = "https://api.openai.com/v1"
	defaultOpenAIModel     = "gpt-4o-mini"
	defaultOpenAIAPIKeyEnv = "OPENAI_API_KEY"
	defaultProviderTimeout = 30 * time.Second
	defaultOpenAIDriver    = "openai"
	defaultChatGPTDriver   = "chatgpt" // alias for openai
	defaultLocalDriver     = "local"
)

func normalizeDriver(driver string) string {
	switch strings.ToLower(strings.TrimSpace(driver)) {
	case defaultOpenAIDriver, defaultChatGPTDriver, "gpt":
		return defaultOpenAIDriver
	default:
		return strings.TrimSpace(driver)
	}
}

// ProviderConfig configures a single AI provider instance.
type ProviderConfig struct {
	Driver       string        `mapstructure:"driver" yaml:"driver" json:"driver"`
	Enabled      *bool         `mapstructure:"enabled" yaml:"enabled" json:"enabled,omitempty"`
	APIKey       string        `mapstructure:"api_key" yaml:"api_key" json:"api_key,omitempty"`
	APIKeyEnv    string        `mapstructure:"api_key_env" yaml:"api_key_env" json:"api_key_env,omitempty"`
	BaseURL      string        `mapstructure:"base_url" yaml:"base_url" json:"base_url,omitempty"`
	DefaultModel string        `mapstructure:"default_model" yaml:"default_model" json:"default_model,omitempty"`
	Timeout      time.Duration `mapstructure:"timeout" yaml:"timeout" json:"timeout,omitempty"`
}

func (p ProviderConfig) isEnabled() bool {
	if p.Enabled == nil {
		return true
	}
	return *p.Enabled
}

func (p ProviderConfig) driver(providerID string) string {
	if d := strings.TrimSpace(p.Driver); d != "" {
		return d
	}
	return strings.TrimSpace(providerID)
}

func (p *ProviderConfig) normalize(providerID string) {
	p.Driver = normalizeDriver(p.driver(providerID))
	if strings.TrimSpace(p.BaseURL) == "" && p.Driver == defaultOpenAIDriver {
		p.BaseURL = defaultOpenAIBaseURL
	}
	if strings.TrimSpace(p.DefaultModel) == "" && p.Driver == defaultOpenAIDriver {
		p.DefaultModel = defaultOpenAIModel
	}
	if strings.TrimSpace(p.APIKeyEnv) == "" && p.Driver == defaultOpenAIDriver {
		p.APIKeyEnv = defaultOpenAIAPIKeyEnv
	}
	if p.Timeout <= 0 {
		p.Timeout = defaultProviderTimeout
	}
}

func (p ProviderConfig) resolvedAPIKey() string {
	if key := strings.TrimSpace(p.APIKey); key != "" {
		return key
	}
	if env := strings.TrimSpace(p.APIKeyEnv); env != "" {
		return strings.TrimSpace(os.Getenv(env))
	}
	return ""
}

// AIConfig is the AI block configuration (main YAML node or standalone ai YAML).
type AIConfig struct {
	Enabled    bool                      `mapstructure:"enabled" yaml:"enabled" json:"enabled"`
	ConfigPath string                    `mapstructure:"config_path" yaml:"config_path,omitempty" json:"config_path,omitempty"`
	Providers  map[string]ProviderConfig `mapstructure:"providers" yaml:"providers" json:"providers"`
	Routes     map[string]any            `mapstructure:"routes" yaml:"routes" json:"routes"`
	routeTable RouteTable                `yaml:"-" mapstructure:"-" json:"-"`
}

// Resolved returns a normalized, validated copy.
func (c AIConfig) Resolved() (AIConfig, error) {
	out := c
	out.normalize()
	if err := out.Validate(); err != nil {
		return AIConfig{}, err
	}
	return out, nil
}

// ResolveBlockConfig loads an external file when config_path is set, otherwise resolves inline input.
func ResolveBlockConfig(input AIConfig) (AIConfig, error) {
	path := strings.TrimSpace(input.ConfigPath)
	if path != "" {
		if _, err := os.Stat(path); err == nil {
			fileCfg, err := LoadConfig(path)
			if err != nil {
				return AIConfig{}, err
			}
			mergeAIConfig(&fileCfg, input)
			fileCfg.normalize()
			if err := fileCfg.Validate(); err != nil {
				return AIConfig{}, err
			}
			return fileCfg, nil
		}
	}
	return input.Resolved()
}

func mergeAIConfig(base *AIConfig, overlay AIConfig) {
	if overlay.Enabled {
		base.Enabled = overlay.Enabled
	}
	if len(overlay.Providers) > 0 {
		base.Providers = copyProviders(overlay.Providers)
	}
	if len(overlay.Routes) > 0 {
		base.Routes = copyRouteMap(overlay.Routes)
	}
}

func LoadConfig(path string) (AIConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return AIConfig{}, wrapErr(ErrLoadConfig, "path", fmt.Errorf("%s: %w", path, err))
	}

	var cfg AIConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return AIConfig{}, wrapErr(ErrLoadConfig, "parse", err)
	}

	cfg.normalize()
	if err := cfg.Validate(); err != nil {
		return AIConfig{}, err
	}
	return cfg, nil
}

func (c *AIConfig) normalize() {
	if c.Providers == nil {
		c.Providers = map[string]ProviderConfig{}
	}
	if c.Routes == nil {
		c.Routes = map[string]any{}
	}
	for id := range c.Providers {
		p := c.Providers[id]
		p.normalize(id)
		c.Providers[id] = p
	}
	if len(c.Routes) > 0 {
		table, err := ParseRouteMap(c.Routes)
		if err == nil {
			c.routeTable = table
		}
	}
}

func (c AIConfig) Validate() error {
	if !c.Enabled {
		return nil
	}

	if len(c.Providers) == 0 {
		return invalidConfig("providers", "at least one provider is required when ai is enabled")
	}

	for id, provider := range c.Providers {
		if strings.TrimSpace(id) == "" {
			return invalidConfig("providers", "provider id is required")
		}
		if !provider.isEnabled() {
			continue
		}
		if err := validateProviderConfig(id, provider); err != nil {
			return err
		}
	}

	if len(c.Routes) == 0 {
		return invalidConfig("routes", "at least one route is required when ai is enabled")
	}

	table, err := ParseRouteMap(c.Routes)
	if err != nil {
		return err
	}
	if table.IsEmpty() {
		return invalidConfig("routes", "at least one route is required when ai is enabled")
	}

	if err := validateRouteTable(table, c.Providers); err != nil {
		return err
	}

	if _, ok := table.resolve(CapabilityTranslation, OpTranslate); !ok {
		return invalidConfig("routes.translation", "translation.translate route is required when ai is enabled")
	}

	return nil
}

func validateProviderConfig(id string, provider ProviderConfig) error {
	switch normalizeDriver(provider.driver(id)) {
	case defaultOpenAIDriver:
		if provider.resolvedAPIKey() == "" {
			return invalidConfigf("providers.%s.api_key", id, "api_key or %s env var is required", provider.APIKeyEnv)
		}
		if strings.TrimSpace(provider.BaseURL) == "" {
			return invalidConfigf("providers.%s.base_url", id, "base_url is required")
		}
		if strings.TrimSpace(provider.DefaultModel) == "" {
			return invalidConfigf("providers.%s.default_model", id, "default_model is required")
		}
	case defaultLocalDriver:
		return nil
	default:
		return invalidConfigf("providers.%s.driver", id, "unsupported driver %q", provider.driver(id))
	}
	return nil
}

func (p ProviderConfig) normalizedDriver(providerID string) string {
	return normalizeDriver(p.driver(providerID))
}

func validateRouteTable(table RouteTable, providers map[string]ProviderConfig) error {
	for capability, route := range table.Capabilities {
		if err := validateCapabilityRoute(string(capability), route, providers); err != nil {
			return err
		}
	}
	return nil
}

func validateCapabilityRoute(capabilityName string, route CapabilityRoute, providers map[string]ProviderConfig) error {
	capability, err := parseCapability(capabilityName)
	if err != nil {
		return err
	}

	if defaultID := strings.TrimSpace(route.Default); defaultID != "" {
		if err := validateRouteTarget(capability, "", defaultID, providers); err != nil {
			return invalidConfigf("routes.%s.default", capabilityName, "%s", err.Error())
		}
	}

	for op, providerID := range route.Operations {
		operation, err := parseOperation(capability, op)
		if err != nil {
			return invalidConfigf("routes.%s.operations.%s", capabilityName, op, "%s", err.Error())
		}
		if err := validateRouteTarget(capability, operation, providerID, providers); err != nil {
			return invalidConfigf("routes.%s.operations.%s", capabilityName, op, "%s", err.Error())
		}
	}

	if strings.TrimSpace(route.Default) == "" && len(route.Operations) == 0 {
		return invalidConfigf("routes.%s", capabilityName, "default or operations is required")
	}
	return nil
}

func validateRouteTarget(capability Capability, operation Operation, providerID string, providers map[string]ProviderConfig) error {
	providerID = strings.TrimSpace(providerID)
	if providerID == "" {
		return fmt.Errorf("provider id is required")
	}
	provider, ok := providers[providerID]
	if !ok {
		return fmt.Errorf("unknown provider %q", providerID)
	}
	if !provider.isEnabled() {
		return fmt.Errorf("provider %q is disabled", providerID)
	}
	driver := provider.normalizedDriver(providerID)
	if operation == "" {
		return nil
	}
	if !providerSupportsOperation(driver, capability, operation) {
		return fmt.Errorf("provider %q (driver %q) does not support %s.%s", providerID, driver, capability, operation)
	}
	return nil
}

func copyProviders(in map[string]ProviderConfig) map[string]ProviderConfig {
	out := make(map[string]ProviderConfig, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

func copyRouteMap(in map[string]any) map[string]any {
	out := make(map[string]any, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}
