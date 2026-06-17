package config

import (
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/spf13/viper"
)

type LoadOptions struct {
	// EnvPrefix is the Viper env prefix (e.g. "VIA_JERI" → VIA_JERI_DATABASE_HOST).
	EnvPrefix string
	// ConfigEnvVar, when set, names an env var that points to a specific config file.
	ConfigEnvVar string
	// ConfigName is the YAML file name without extension (default "config").
	ConfigName string
	// Paths are directories searched for ConfigName.yaml (default "./config", ".").
	Paths []string
	// DefaultAppName fills appkit BaseConfig defaults when T embeds config.BaseConfig.
	DefaultAppName string
}

// Load reads YAML config via Viper, unmarshals into T, and applies BaseConfig defaults
// when DefaultAppName is set and T embeds config.BaseConfig.
func Load[T any](opts LoadOptions) (*T, error) {
	if opts.ConfigName == "" {
		opts.ConfigName = "config"
	}
	if len(opts.Paths) == 0 {
		opts.Paths = []string{"./config", "."}
	}

	if opts.ConfigEnvVar != "" {
		if path := os.Getenv(opts.ConfigEnvVar); path != "" {
			viper.SetConfigFile(path)
		}
	}

	if viper.ConfigFileUsed() == "" {
		viper.SetConfigName(opts.ConfigName)
		viper.SetConfigType("yaml")
		for _, path := range opts.Paths {
			viper.AddConfigPath(path)
		}
	}

	if opts.EnvPrefix != "" {
		viper.SetEnvPrefix(opts.EnvPrefix)
		viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
		viper.AutomaticEnv()
	}

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	var cfg T
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	applyEmbeddedBaseDefaults(&cfg, opts.DefaultAppName)

	return &cfg, nil
}

func applyEmbeddedBaseDefaults(cfg any, defaultAppName string) {
	if defaultAppName == "" {
		return
	}

	v := reflect.ValueOf(cfg)
	if v.Kind() != reflect.Pointer || v.IsNil() {
		return
	}
	v = v.Elem()
	if v.Kind() != reflect.Struct {
		return
	}

	f := v.FieldByName("BaseConfig")
	if !f.IsValid() || !f.CanAddr() {
		return
	}

	base, ok := f.Addr().Interface().(*BaseConfig)
	if !ok {
		return
	}

	base.ApplyDefaults(defaultAppName)
}
