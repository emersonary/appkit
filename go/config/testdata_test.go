package config

import (
	_ "embed"
	"os"
	"path/filepath"
	"testing"
)

const testConfigEnvVar = "APPKIT_TEST_CONFIG"

//go:embed testdata/base.example.yaml
var baseExampleYAML []byte

// Expected GET /config response for base.example.yaml (secrets omitted via json:"-").
//
//go:embed testdata/base.example.json
var baseExampleJSON []byte

//go:embed testdata/product.example.yaml
var productExampleYAML []byte

func loadYAMLFixture[T any](t *testing.T, content []byte, opts LoadOptions) *T {
	t.Helper()

	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	t.Setenv(testConfigEnvVar, path)

	opts.ConfigEnvVar = testConfigEnvVar

	cfg, err := Load[T](opts)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	return cfg
}
