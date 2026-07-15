package migrate

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfigFromFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "dbhist.yaml")
	content := `
exclude_patterns:
  - tbl_%_staging
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadConfig(path)
	if err != nil {
		t.Fatal(err)
	}

	if len(cfg.ExcludePatterns) != 1 {
		t.Fatalf("exclude_patterns: %#v", cfg.ExcludePatterns)
	}
}

func TestCommentHasMarker(t *testing.T) {
	comment := "audit=true repo=true"
	if !commentHasMarker(comment, auditCommentMarker) {
		t.Fatal("expected AUDIT marker")
	}
	if !commentHasMarker(comment, repoCommentMarker) {
		t.Fatal("expected REPO marker")
	}
	if commentHasMarker("AUDIT=true", repoCommentMarker) {
		t.Fatal("did not expect REPO marker")
	}
}
