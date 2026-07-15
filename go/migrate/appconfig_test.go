package migrate

import "testing"

func TestResolveBlockConfigEnabledOnly(t *testing.T) {
	cfg, err := ResolveBlockConfig(AppConfig{Enabled: true})
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.ExcludePatterns) != 0 {
		t.Fatalf("expected empty exclude patterns, got %#v", cfg.ExcludePatterns)
	}
}

func TestFilterTablesByMarkers(t *testing.T) {
	tables := []Table{
		{Schema: "trip", Name: "tbl_trip", Audit: true, Repo: true},
		{Schema: "core", Name: "tbl_customer", Audit: true},
		{Schema: "trip", Name: "tbl_other", Repo: true},
	}
	audit := filterTables(tables, func(t Table) bool { return t.Audit })
	if len(audit) != 2 {
		t.Fatalf("audit: %#v", audit)
	}
	repo := filterTables(tables, func(t Table) bool { return t.Repo })
	if len(repo) != 2 {
		t.Fatalf("repo: %#v", repo)
	}
}
