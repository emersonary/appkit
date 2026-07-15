package migrate

import (
	"testing"
	"time"
)

func TestTableFunctionsMapEntryHoldsCoreNames(t *testing.T) {
	functionsMu.Lock()
	functions = map[string]TableFunctions{
		"tenant.tenant_business_profiles": {
			Schema: "tenant",
			Table:  "tenant_business_profiles",
			Insert: "func_insert_tenant_business_profiles_v000001",
			Update: "func_update_tenant_business_profiles_v000001",
			Delete: "func_delete_tenant_business_profiles_v000001",
			Get:      "func_get_tenant_business_profiles_v000001",
			Upsert:   "func_upsert_tenant_business_profiles_v000001",
			List:     "func_list_tenant_business_profiles_v000001",
			ListRepo: false,
			Audit:    true,
			Repo:     true,
		},
	}
	functionsMu.Unlock()

	fn, err := MustFunctions("tenant", "tenant_business_profiles")
	if err != nil {
		t.Fatal(err)
	}
	if fn.Insert == "" || fn.Update == "" || fn.Delete == "" || fn.Get == "" {
		t.Fatalf("expected insert/update/delete/get names, got %+v", fn)
	}
	if fn.List == "" {
		t.Fatalf("expected list name, got %+v", fn)
	}
	if !fn.Audit || !fn.Repo {
		t.Fatalf("expected audit/repo true, got audit=%v repo=%v", fn.Audit, fn.Repo)
	}
	if fn.ListRepo {
		t.Fatal("expected ListRepo false when table has no children")
	}
}

func TestRepoEnabledReadsTableFunctions(t *testing.T) {
	functionsMu.Lock()
	functions = map[string]TableFunctions{
		"tenant.tenant_locations": {Schema: "tenant", Table: "tenant_locations", Audit: true, Repo: true},
		"tenant.other":            {Schema: "tenant", Table: "other", Audit: true, Repo: false},
	}
	functionsMu.Unlock()

	if !RepoEnabled("tenant", "tenant_locations") {
		t.Fatal("expected locations repo enabled")
	}
	if !AuditEnabled("tenant", "tenant_locations") {
		t.Fatal("expected locations audit enabled")
	}
	if RepoEnabled("tenant", "other") {
		t.Fatal("expected other repo disabled")
	}
	if !AuditEnabled("tenant", "other") {
		t.Fatal("expected other audit enabled")
	}
	if RepoEnabled("tenant", "missing") {
		t.Fatal("expected missing table repo false")
	}
}

func TestRegistryFlagsListWithoutChildren(t *testing.T) {
	table := Table{Audit: true, Repo: true, Children: nil}
	audit, repo := registryFlags(table, "list")
	if !audit || repo {
		t.Fatalf("list without children: audit=%v repo=%v", audit, repo)
	}

	table.Children = []ChildRelation{{Table: "child"}}
	audit, repo = registryFlags(table, "list")
	if !audit || !repo {
		t.Fatalf("list with children: audit=%v repo=%v", audit, repo)
	}

	_, repo = registryFlags(table, "get")
	if !repo {
		t.Fatal("get should keep table.Repo")
	}
}

func TestMarkRepoExecutedSetsDirtyAndLastExecuted(t *testing.T) {
	executionsMu.Lock()
	executions = map[string]functionExec{}
	executionsMu.Unlock()

	before := time.Now().UTC().Add(-time.Second)
	fnName := "func_get_tenant_locations_v000001"
	MarkRepoExecuted("tenant", "tenant_locations", fnName)

	last, ok := LastExecuted("tenant", "tenant_locations", fnName)
	if !ok {
		t.Fatal("expected LastExecuted present")
	}
	if last.Before(before) || last.IsZero() {
		t.Fatalf("expected LastExecuted >= %v, got %v", before, last)
	}

	executionsMu.RLock()
	ex := executions[executionKey("tenant", "tenant_locations", fnName)]
	executionsMu.RUnlock()
	if !ex.Dirty {
		t.Fatal("expected Dirty true")
	}

	MarkRepoExecuted("tenant", "tenant_locations", "")
	if _, ok := LastExecuted("tenant", "tenant_locations", ""); ok {
		t.Fatal("empty function name should be ignored")
	}
}

func TestClearDirtyIfUnchangedKeepsNewerMark(t *testing.T) {
	synced := time.Now().UTC().Add(-time.Minute)
	newer := synced.Add(time.Second)
	fnName := "func_get_tenant_locations_v000001"
	executionsMu.Lock()
	executions = map[string]functionExec{
		executionKey("tenant", "tenant_locations", fnName): {
			Schema:       "tenant",
			Table:        "tenant_locations",
			FunctionName: fnName,
			Dirty:        true,
			LastExecuted: newer,
		},
	}
	executionsMu.Unlock()

	clearDirtyIfUnchanged("tenant", "tenant_locations", fnName, synced)
	executionsMu.RLock()
	ex := executions[executionKey("tenant", "tenant_locations", fnName)]
	executionsMu.RUnlock()
	if !ex.Dirty {
		t.Fatal("expected Dirty kept when LastExecuted is newer than sync snapshot")
	}

	clearDirtyIfUnchanged("tenant", "tenant_locations", fnName, newer)
	executionsMu.RLock()
	ex = executions[executionKey("tenant", "tenant_locations", fnName)]
	executionsMu.RUnlock()
	if ex.Dirty {
		t.Fatal("expected Dirty cleared when LastExecuted matches sync snapshot")
	}
}
