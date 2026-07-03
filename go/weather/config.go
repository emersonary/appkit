package weather

import (
	"fmt"
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

const (
	defaultOpenMeteoBaseURL = "https://api.open-meteo.com/v1/forecast"
	defaultLatitude         = -2.795
	defaultLongitude        = -40.514
	defaultForecastDays     = 7
	defaultWindSpeedUnit    = "kn"
	defaultTimezone         = "auto"
	defaultLowWindMaxKnots  = 12.0
	defaultRefreshInterval  = time.Hour
	defaultCacheTTL         = 10 * 24 * time.Hour
	defaultKeyPrefix        = "weather:openmeteo:jeri"
)

type OpenMeteoConfig struct {
	BaseURL       string  `mapstructure:"base_url" yaml:"base_url" json:"base_url"`
	Latitude      float64 `mapstructure:"latitude" yaml:"latitude" json:"latitude"`
	Longitude     float64 `mapstructure:"longitude" yaml:"longitude" json:"longitude"`
	ForecastDays  int     `mapstructure:"forecast_days" yaml:"forecast_days" json:"forecast_days"`
	WindSpeedUnit string  `mapstructure:"wind_speed_unit" yaml:"wind_speed_unit" json:"wind_speed_unit"`
	Timezone      string  `mapstructure:"timezone" yaml:"timezone" json:"timezone"`
}

// AppConfig is the application-level weather block config (main YAML weather node).
type AppConfig struct {
	Enabled         bool            `mapstructure:"enabled" yaml:"enabled" json:"enabled"`
	ConfigPath      string          `mapstructure:"config_path" yaml:"config_path,omitempty" json:"config_path,omitempty"`
	KeyPrefix       string          `mapstructure:"key_prefix" yaml:"key_prefix" json:"key_prefix"`
	RefreshInterval time.Duration   `mapstructure:"refresh_interval" yaml:"refresh_interval" json:"refresh_interval"`
	CacheTTL        time.Duration   `mapstructure:"cache_ttl" yaml:"cache_ttl" json:"cache_ttl"`
	LowWindMaxKnots float64         `mapstructure:"low_wind_max_knots" yaml:"low_wind_max_knots" json:"low_wind_max_knots"`
	LowGustMaxKnots float64         `mapstructure:"low_gust_max_knots" yaml:"low_gust_max_knots,omitempty" json:"low_gust_max_knots,omitempty"`
	OpenMeteo       OpenMeteoConfig `mapstructure:"open_meteo" yaml:"open_meteo" json:"open_meteo"`
}

func (c AppConfig) Resolved() (AppConfig, error) {
	out := c
	out.normalize()
	if err := out.Validate(); err != nil {
		return AppConfig{}, err
	}
	return out, nil
}

func ResolveBlockConfig(input AppConfig) (AppConfig, error) {
	path := strings.TrimSpace(input.ConfigPath)
	if path != "" {
		if _, err := os.Stat(path); err == nil {
			fileCfg, err := LoadConfig(path)
			if err != nil {
				return AppConfig{}, err
			}
			mergeAppConfig(&fileCfg, input)
			fileCfg.normalize()
			if err := fileCfg.Validate(); err != nil {
				return AppConfig{}, err
			}
			return fileCfg, nil
		}
	}
	return input.Resolved()
}

func LoadConfig(path string) (AppConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return AppConfig{}, wrapErr(ErrLoadConfig, "path", fmt.Errorf("%s: %w", path, err))
	}

	var cfg AppConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return AppConfig{}, wrapErr(ErrLoadConfig, "parse", err)
	}

	cfg.normalize()
	if err := cfg.Validate(); err != nil {
		return AppConfig{}, err
	}
	return cfg, nil
}

func mergeAppConfig(base *AppConfig, overlay AppConfig) {
	if overlay.Enabled {
		base.Enabled = true
	}
	if strings.TrimSpace(overlay.KeyPrefix) != "" {
		base.KeyPrefix = overlay.KeyPrefix
	}
	if overlay.RefreshInterval > 0 {
		base.RefreshInterval = overlay.RefreshInterval
	}
	if overlay.CacheTTL > 0 {
		base.CacheTTL = overlay.CacheTTL
	}
	if overlay.LowWindMaxKnots > 0 {
		base.LowWindMaxKnots = overlay.LowWindMaxKnots
	}
	if overlay.LowGustMaxKnots > 0 {
		base.LowGustMaxKnots = overlay.LowGustMaxKnots
	}
	mergeOpenMeteoConfig(&base.OpenMeteo, overlay.OpenMeteo)
}

func mergeOpenMeteoConfig(base *OpenMeteoConfig, overlay OpenMeteoConfig) {
	if strings.TrimSpace(overlay.BaseURL) != "" {
		base.BaseURL = overlay.BaseURL
	}
	if overlay.Latitude != 0 {
		base.Latitude = overlay.Latitude
	}
	if overlay.Longitude != 0 {
		base.Longitude = overlay.Longitude
	}
	if overlay.ForecastDays > 0 {
		base.ForecastDays = overlay.ForecastDays
	}
	if strings.TrimSpace(overlay.WindSpeedUnit) != "" {
		base.WindSpeedUnit = overlay.WindSpeedUnit
	}
	if strings.TrimSpace(overlay.Timezone) != "" {
		base.Timezone = overlay.Timezone
	}
}

func (c *AppConfig) normalize() {
	if strings.TrimSpace(c.KeyPrefix) == "" {
		c.KeyPrefix = defaultKeyPrefix
	}
	if c.RefreshInterval <= 0 {
		c.RefreshInterval = defaultRefreshInterval
	}
	if c.CacheTTL <= 0 {
		c.CacheTTL = defaultCacheTTL
	}
	if c.LowWindMaxKnots <= 0 {
		c.LowWindMaxKnots = defaultLowWindMaxKnots
	}
	c.OpenMeteo.normalize()
}

func (c *OpenMeteoConfig) normalize() {
	if strings.TrimSpace(c.BaseURL) == "" {
		c.BaseURL = defaultOpenMeteoBaseURL
	}
	if c.Latitude == 0 {
		c.Latitude = defaultLatitude
	}
	if c.Longitude == 0 {
		c.Longitude = defaultLongitude
	}
	if c.ForecastDays <= 0 {
		c.ForecastDays = defaultForecastDays
	}
	if strings.TrimSpace(c.WindSpeedUnit) == "" {
		c.WindSpeedUnit = defaultWindSpeedUnit
	}
	if strings.TrimSpace(c.Timezone) == "" {
		c.Timezone = defaultTimezone
	}
}

func (c AppConfig) Validate() error {
	if !c.Enabled {
		return nil
	}
	if strings.TrimSpace(c.KeyPrefix) == "" {
		return invalidConfig("key_prefix", "required")
	}
	if c.RefreshInterval <= 0 {
		return invalidConfig("refresh_interval", "must be positive")
	}
	if c.CacheTTL <= 0 {
		return invalidConfig("cache_ttl", "must be positive")
	}
	if c.LowWindMaxKnots <= 0 {
		return invalidConfig("low_wind_max_knots", "must be positive")
	}
	return c.OpenMeteo.Validate()
}

func (c OpenMeteoConfig) Validate() error {
	if strings.TrimSpace(c.BaseURL) == "" {
		return invalidConfig("open_meteo.base_url", "required")
	}
	if c.Latitude < -90 || c.Latitude > 90 {
		return invalidConfigf("open_meteo.latitude", "must be between -90 and 90, got %f", c.Latitude)
	}
	if c.Longitude < -180 || c.Longitude > 180 {
		return invalidConfigf("open_meteo.longitude", "must be between -180 and 180, got %f", c.Longitude)
	}
	if c.ForecastDays <= 0 {
		return invalidConfig("open_meteo.forecast_days", "must be positive")
	}
	return nil
}
