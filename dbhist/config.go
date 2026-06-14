package dbhist

import (
	"fmt"
	"os"
	"regexp"

	"gopkg.in/yaml.v3"
)

const (
	defaultTablePattern = "tbl_%"
)

var (
	implicitExcludePatterns = []string{"%_hist", "%_hist_detail"}
	identPattern            = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)
	tablePatternPattern     = regexp.MustCompile(`^[a-zA-Z0-9_%]+$`)
)

type Modules struct {
	Audit         bool `yaml:"audit"`
	History       bool `yaml:"history"`
	RepoFunctions bool `yaml:"repo_functions"`
}

type Config struct {
	Schemas         []string `yaml:"schemas"`
	ExcludePatterns []string `yaml:"exclude_patterns"`
	TablePattern    string   `yaml:"table_pattern"`
	Modules         Modules  `yaml:"modules"`
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

	cfg.applyDefaults()
	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func (c *Config) applyDefaults() {
	if c.TablePattern == "" {
		c.TablePattern = defaultTablePattern
	}

	if !c.Modules.Audit && !c.Modules.History && !c.Modules.RepoFunctions {
		c.Modules.Audit = true
		c.Modules.History = true
		c.Modules.RepoFunctions = true
	}
}

func (c Config) Validate() error {
	if len(c.Schemas) == 0 {
		return ErrSchemasEmpty
	}

	for _, schema := range c.Schemas {
		if err := validateIdent(schema); err != nil {
			return ErrInvalidSchema.With(schema, err.Error())
		}
	}

	if !tablePatternPattern.MatchString(c.TablePattern) {
		return ErrInvalidTablePattern.With("table_pattern", c.TablePattern)
	}

	for _, pattern := range c.ExcludePatterns {
		if !tablePatternPattern.MatchString(pattern) {
			return ErrInvalidExcludePattern.With("exclude_pattern", pattern)
		}
	}

	return nil
}

func (c Config) allExcludePatterns() []string {
	patterns := make([]string, 0, len(implicitExcludePatterns)+len(c.ExcludePatterns))
	patterns = append(patterns, implicitExcludePatterns...)
	patterns = append(patterns, c.ExcludePatterns...)
	return patterns
}

func validateIdent(name string) error {
	if !identPattern.MatchString(name) {
		return fmt.Errorf("invalid identifier %q", name)
	}

	return nil
}