package config

import (
	"fmt"
	"strings"

	appkitlog "github.com/emersonary/appkit/log"
)

type App struct {
	Name string `mapstructure:"name" json:"name"`
}

// ListenerConfig is a bind address toggled by enabled.
type ListenerConfig struct {
	Enabled bool   `mapstructure:"enabled" json:"enabled"`
	Addr    string `mapstructure:"addr" json:"addr"`
}

func (l ListenerConfig) Active() bool {
	return l.Enabled && strings.TrimSpace(l.Addr) != ""
}

type Server struct {
	HTTP      ListenerConfig `mapstructure:"http" json:"http"`
	GRPC      ListenerConfig `mapstructure:"grpc" json:"grpc"`
	WebSocket ListenerConfig `mapstructure:"websocket" json:"websocket"`
}

type Database struct {
	Host     string `mapstructure:"host" json:"host"`
	Port     int    `mapstructure:"port" json:"port"`
	User     string `mapstructure:"user" json:"user"`
	Password string `mapstructure:"password" json:"-"`
	Name     string `mapstructure:"name" json:"name"`
	SSLMode  string `mapstructure:"sslmode" json:"sslmode"`
}

func (d Database) DSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		d.User,
		d.Password,
		d.Host,
		d.Port,
		d.Name,
		d.SSLMode,
	)
}

type NATS struct {
	URL string `mapstructure:"url" json:"url"`
}

type Redis struct {
	Enabled  bool   `mapstructure:"enabled" json:"enabled"`
	Addr     string `mapstructure:"addr" json:"addr"`
	Password string `mapstructure:"password" json:"-"`
	DB       int    `mapstructure:"db" json:"db"`
}

type Migrations struct {
	Path string `mapstructure:"path" json:"path"`
}

// BaseConfig is the shared application config loaded from YAML (one node per concern).
// Reference shape: testdata/base.example.yaml
type BaseConfig struct {
	App        App              `mapstructure:"app" json:"app"`
	Log        appkitlog.Config `mapstructure:"log" json:"log"`
	Server     Server           `mapstructure:"server" json:"server"`
	Database   Database         `mapstructure:"database" json:"database"`
	NATS       NATS             `mapstructure:"nats" json:"nats"`
	Redis      Redis            `mapstructure:"redis" json:"redis"`
	Migrations Migrations       `mapstructure:"migrations" json:"migrations"`
	Blocks     `mapstructure:",squash"`
}

// AppConfig is implemented by product configs that embed BaseConfig.
type AppConfig interface {
	Infra() *BaseConfig
}

// AppName returns the configured application name.
func (b *BaseConfig) AppName() string {
	return b.App.Name
}

// ApplyDefaults fills zero values with standard defaults.
func (b *BaseConfig) ApplyDefaults(defaultAppName string) {
	if b.App.Name == "" {
		b.App.Name = defaultAppName
	}
	if b.Server.HTTP.Enabled && b.Server.HTTP.Addr == "" {
		b.Server.HTTP.Addr = ":8080"
	}
	if b.Server.GRPC.Enabled && b.Server.GRPC.Addr == "" {
		b.Server.GRPC.Addr = ":9090"
	}
	if b.Migrations.Path == "" {
		b.Migrations.Path = "db/migrations"
	}
	if b.NATS.URL == "" {
		b.NATS.URL = "nats://localhost:4222"
	}
	if b.Redis.Enabled && b.Redis.Addr == "" {
		b.Redis.Addr = "localhost:6379"
	}
	if b.Log.Level == "" {
		b.Log.Level = "info"
	}
	if b.Log.Format == "" {
		b.Log.Format = "json"
	}
}
