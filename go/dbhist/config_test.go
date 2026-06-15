package dbhist

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfigFromFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "dbhist.yaml")
	content := `
schemas:
  - core
  - trip
exclude_patterns:
  - tbl_%_staging
modules:
  audit: true
  history: false
  repo_functions: true
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadConfig(path)
	if err != nil {
		t.Fatal(err)
	}

	if len(cfg.Schemas) != 2 || cfg.Schemas[0] != "core" {
		t.Fatalf("unexpected schemas: %#v", cfg.Schemas)
	}

	if !cfg.Modules.Audit || cfg.Modules.History || !cfg.Modules.RepoFunctions {
		t.Fatalf("unexpected modules: %#v", cfg.Modules)
	}
}
