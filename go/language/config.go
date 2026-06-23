package language

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

const defaultLanguageCode = "en"

// LanguageConfig is the language block configuration (main YAML node or standalone language YAML).
type LanguageConfig struct {
	Enabled         bool     `mapstructure:"enabled" yaml:"enabled" json:"enabled"`
	ConfigPath      string   `mapstructure:"config_path" yaml:"config_path,omitempty" json:"config_path,omitempty"`
	Schema          string   `mapstructure:"schema" yaml:"schema" json:"schema"`
	DefaultLanguage string   `mapstructure:"default_language" yaml:"default_language" json:"default_language"`
	Languages       []string `mapstructure:"languages" yaml:"languages" json:"languages"`
	SkipSeed        bool     `mapstructure:"skip_seed" yaml:"skip_seed" json:"skip_seed"`
}

// ApplyDefaults fills zero values for optional fields.
func (c *LanguageConfig) ApplyDefaults() {
	if strings.TrimSpace(c.DefaultLanguage) == "" {
		c.DefaultLanguage = defaultLanguageCode
	}
}

// Resolved returns a normalized, validated copy suitable for schema and service wiring.
func (c LanguageConfig) Resolved() (LanguageConfig, error) {
	out := c
	out.ApplyDefaults()
	out.normalize()
	if err := out.Validate(); err != nil {
		return LanguageConfig{}, err
	}
	return out, nil
}

// ResolveBlockConfig returns the effective language config.
// When config_path points at a legacy block YAML file, that file is loaded and input fields override it.
func ResolveBlockConfig(input LanguageConfig) (LanguageConfig, error) {
	input.ApplyDefaults()

	path := strings.TrimSpace(input.ConfigPath)
	if path != "" {
		if _, err := os.Stat(path); err == nil {
			fileCfg, err := LoadConfig(path)
			if err != nil {
				return LanguageConfig{}, err
			}
			mergeLanguageConfig(&fileCfg, input)
			fileCfg.normalize()
			if err := fileCfg.Validate(); err != nil {
				return LanguageConfig{}, err
			}
			return fileCfg, nil
		}
	}

	return input.Resolved()
}

func mergeLanguageConfig(base *LanguageConfig, overlay LanguageConfig) {
	if overlay.Schema != "" {
		base.Schema = overlay.Schema
	}
	if overlay.DefaultLanguage != "" {
		base.DefaultLanguage = NormalizeCode(overlay.DefaultLanguage)
	}
	if len(overlay.Languages) > 0 {
		base.Languages = normalizeCodeList(overlay.Languages)
	}
	base.SkipSeed = overlay.SkipSeed || base.SkipSeed
}

func LoadConfig(path string) (LanguageConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return LanguageConfig{}, wrapErr(ErrLoadConfig, "path", fmt.Errorf("%s: %w", path, err))
	}

	var cfg LanguageConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return LanguageConfig{}, wrapErr(ErrLoadConfig, "parse", err)
	}

	cfg.normalize()
	if err := cfg.Validate(); err != nil {
		return LanguageConfig{}, err
	}

	return cfg, nil
}

func (c *LanguageConfig) normalize() {
	c.Schema = strings.TrimSpace(c.Schema)
	if c.DefaultLanguage == "" {
		c.DefaultLanguage = defaultLanguageCode
	}

	c.DefaultLanguage = NormalizeCode(c.DefaultLanguage)
	c.Languages = normalizeCodeList(c.Languages)
}

func (c LanguageConfig) Validate() error {
	if c.Schema == "" {
		return ErrSchemaRequired
	}

	if err := validateIdent(c.Schema); err != nil {
		return ErrInvalidSchema.With("schema", err.Error())
	}

	if strings.TrimSpace(c.DefaultLanguage) == "" {
		return ErrDefaultLanguageRequired
	}

	if err := validateCatalogCode(c.DefaultLanguage); err != nil {
		return err
	}

	if len(c.Languages) == 0 {
		return ErrEmptyLanguages
	}

	seen := make(map[string]struct{}, len(c.Languages))
	hasDefault := false

	for _, code := range c.Languages {
		if err := validateCatalogCode(code); err != nil {
			return err
		}

		if _, dup := seen[code]; dup {
			return ErrDuplicateLanguage.With("code", code)
		}

		seen[code] = struct{}{}
		if code == c.DefaultLanguage {
			hasDefault = true
		}
	}

	if !hasDefault {
		return ErrDefaultLanguageMissing
	}

	return nil
}

func (c LanguageConfig) ensureDefaultInLanguages() error {
	if _, ok := c.EnabledSet()[c.DefaultLanguage]; !ok {
		return ErrDefaultLanguageMissing
	}
	return nil
}

func (c LanguageConfig) EnabledSet() map[string]struct{} {
	set := make(map[string]struct{}, len(c.Languages))
	for _, code := range c.Languages {
		set[code] = struct{}{}
	}

	return set
}

func (c LanguageConfig) EnabledCodes() []string {
	codes := append([]string(nil), c.Languages...)
	sort.Strings(codes)
	return codes
}
