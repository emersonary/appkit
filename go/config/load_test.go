package config

import "testing"

func TestLoad_BaseExampleYAML(t *testing.T) {
	cfg := loadYAMLFixture[BaseConfig](t, baseExampleYAML, LoadOptions{
		DefaultAppName: "fallback api",
	})

	if cfg.App.Name != "example api" {
		t.Fatalf("App.Name: got %q want %q", cfg.App.Name, "example api")
	}
	if cfg.Log.Level != "info" || cfg.Log.Format != "json" {
		t.Fatalf("Log: got level=%q format=%q", cfg.Log.Level, cfg.Log.Format)
	}
	if !cfg.Server.HTTP.Active() {
		t.Fatal("expected server.http active")
	}
	if cfg.Server.HTTP.Addr != ":8080" {
		t.Fatalf("Server.HTTP.Addr: got %q", cfg.Server.HTTP.Addr)
	}
	if cfg.Server.GRPC.Addr != ":9090" {
		t.Fatalf("Server.GRPC.Addr: got %q", cfg.Server.GRPC.Addr)
	}
	if cfg.Server.WebSocket.Active() {
		t.Fatal("expected server.websocket inactive")
	}
	if cfg.Database.Host != "localhost" || cfg.Database.Port != 5432 {
		t.Fatalf("Database: got host=%q port=%d", cfg.Database.Host, cfg.Database.Port)
	}
	if cfg.Database.DSN() != "postgres://app:secret@localhost:5432/app?sslmode=disable" {
		t.Fatalf("Database.DSN: got %q", cfg.Database.DSN())
	}
	if cfg.NATS.URL != "nats://localhost:4222" {
		t.Fatalf("NATS.URL: got %q", cfg.NATS.URL)
	}
	if cfg.Migrations.Path != "db/migrations" {
		t.Fatalf("Migrations.Path: got %q", cfg.Migrations.Path)
	}
	if cfg.Accounts.Enabled {
		t.Fatal("expected accounts.enabled=false in example")
	}
	if cfg.Currency.Enabled {
		t.Fatal("expected currency.enabled=false in example")
	}
	if cfg.Email.Brand != "Example" {
		t.Fatalf("Email.Brand: got %q", cfg.Email.Brand)
	}
}

func TestLoad_ProductExampleYAML(t *testing.T) {
	type productConfig struct {
		BaseConfig `mapstructure:",squash"`
		Flag       bool `mapstructure:"flag"`
	}

	cfg := loadYAMLFixture[productConfig](t, productExampleYAML, LoadOptions{
		DefaultAppName: "test api",
	})

	if cfg.App.Name != "test api" {
		t.Fatalf("App.Name default: got %q want %q", cfg.App.Name, "test api")
	}
	if !cfg.Server.HTTP.Enabled {
		t.Fatal("expected server.http.enabled=true from yaml")
	}
	if cfg.Server.HTTP.Addr != ":8080" {
		t.Fatalf("Server.HTTP.Addr default: got %q", cfg.Server.HTTP.Addr)
	}
	if cfg.Migrations.Path != "db/migrations" {
		t.Fatalf("Migrations.Path default: got %q", cfg.Migrations.Path)
	}
	if !cfg.Flag {
		t.Fatal("expected flag=true from yaml")
	}
}

func TestListenerConfig_Active(t *testing.T) {
	if (ListenerConfig{}).Active() {
		t.Fatal("empty listener should be inactive")
	}
	if !(ListenerConfig{Enabled: true, Addr: ":8080"}).Active() {
		t.Fatal("enabled listener with addr should be active")
	}
	if (ListenerConfig{Enabled: true, Addr: "  "}).Active() {
		t.Fatal("enabled listener without addr should be inactive")
	}
}
