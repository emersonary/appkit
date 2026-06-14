package currency

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

const defaultConfigBaseCurrency = BaseCurrencyCode

type Config struct {
	Schema       string   `yaml:"schema"`
	BaseCurrency string   `yaml:"base_currency"`
	Currencies   []string `yaml:"currencies"`
	SkipSeed     bool     `yaml:"skip_seed"`
}

func LoadConfig(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, wrapErr(ErrLoadConfig, "path", fmt.Errorf("%s: %w", path, err))
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
	if c.BaseCurrency == "" {
		c.BaseCurrency = defaultConfigBaseCurrency
	}

	c.BaseCurrency = NormalizeCode(c.BaseCurrency)
	c.Currencies = normalizeCodeList(c.Currencies)
}

func (c Config) Validate() error {
	if c.Schema == "" {
		return ErrSchemaRequired
	}

	if err := validateIdent(c.Schema); err != nil {
		return ErrInvalidSchema.With("schema", err.Error())
	}

	if err := validateISO4217Code(c.BaseCurrency); err != nil {
		return err
	}

	if c.BaseCurrency != BaseCurrencyCode {
		return ErrInvalidBaseCurrency.With("base_currency", c.BaseCurrency)
	}

	if len(c.Currencies) == 0 {
		return ErrEmptyCurrencies
	}

	seen := make(map[string]struct{}, len(c.Currencies))
	hasBase := false

	for _, code := range c.Currencies {
		if err := validateISO4217Code(code); err != nil {
			return err
		}

		if _, dup := seen[code]; dup {
			return ErrDuplicateCurrency.With("code", code)
		}

		seen[code] = struct{}{}
		if code == c.BaseCurrency {
			hasBase = true
		}
	}

	if !hasBase {
		return ErrBaseCurrencyMissing
	}

	return nil
}

func (c Config) EnabledSet() map[string]struct{} {
	set := make(map[string]struct{}, len(c.Currencies))
	for _, code := range c.Currencies {
		set[code] = struct{}{}
	}

	return set
}

func (c Config) EnabledCodes() []string {
	codes := append([]string(nil), c.Currencies...)
	sort.Strings(codes)
	return codes
}

func normalizeCodeList(codes []string) []string {
	if len(codes) == 0 {
		return nil
	}

	out := make([]string, 0, len(codes))
	for _, code := range codes {
		out = append(out, NormalizeCode(code))
	}

	sort.Strings(out)
	return out
}

func pgTextArray(values []string) string {
	if len(values) == 0 {
		return "'{}'"
	}

	parts := make([]string, len(values))
	for i, value := range values {
		parts[i] = quoteLiteral(value)
	}

	return "ARRAY[" + strings.Join(parts, ", ") + "]"
}
