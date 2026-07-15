package migrate

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	auditCommentMarker   = "AUDIT=true"
	repoCommentMarker    = "REPO=true"
	repoVersionWidth     = 6
	repoGeneratorVersion = "dbhist-repo/1"
)

var (
	implicitExcludePatterns = []string{"%_hist", "%_hist_detail"}
	tablePatternPattern     = regexp.MustCompile(`^[a-zA-Z0-9_%]+$`)
)

// Config is the resolved dbhist settings used by UpdateHist.
// Table selection is driven by COMMENT markers (AUDIT=true / REPO=true), not YAML modules.
type Config struct {
	ExcludePatterns []string `yaml:"exclude_patterns" mapstructure:"exclude_patterns" json:"exclude_patterns,omitempty"`
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

	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func (c Config) Validate() error {
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

func commentHasMarker(comment, marker string) bool {
	return strings.Contains(
		strings.ToLower(comment),
		strings.ToLower(strings.TrimSpace(marker)),
	)
}
