package migrate

import (
	"strings"
	"testing"
)

func testRepoNames(tables ...Table) repoNames {
	names := repoNames{}
	for _, table := range tables {
		for _, op := range repoOperations() {
			names[makeFuncKey(table.Schema, table.Name, op)] = versionedRepoFunctionName(table.Name, op, 1)
		}
	}
	return names
}

func TestJsonKeyForChild(t *testing.T) {
	tests := []struct {
		table    string
		oneToOne bool
		want     string
	}{
		{"tbl_trip_stop", false, "trip_stops"},
		{"tbl_trip_shareddata", true, "trip_shareddata"},
		{"tbl_trip_participant", false, "trip_participants"},
		{"tbl_trip_luggage", true, "trip_luggage"},
	}

	for _, tc := range tests {
		if got := jsonKeyForChild(tc.table, tc.oneToOne); got != tc.want {
			t.Fatalf("jsonKeyForChild(%q, %v) = %q, want %q", tc.table, tc.oneToOne, got, tc.want)
		}
	}
}

func TestBuildInsertFunctionSQLContainsChildCalls(t *testing.T) {
	tables := []Table{
		{
			Schema:     "trip",
			Name:       "tbl_trip_luggage",
			PrimaryKey: "id_tripluggage",
			Columns: []Column{
				{Name: "id_tripluggage", Type: "uuid", UdtName: "uuid"},
				{Name: "id_tripparticipant", Type: "uuid", UdtName: "uuid"},
				{Name: "n_checkedluggagecount", Type: "integer", DefaultValue: "0"},
				{Name: "str_rowstatus", Type: "character", DefaultValue: "'A'::bpchar"},
			},
		},
		{
			Schema:     "trip",
			Name:       "tbl_trip_participant",
			PrimaryKey: "id_tripparticipant",
			Columns: []Column{
				{Name: "id_tripparticipant", Type: "uuid", UdtName: "uuid"},
				{Name: "id_trip", Type: "uuid", UdtName: "uuid"},
				{Name: "id_customer", Type: "uuid", UdtName: "uuid"},
				{Name: "str_passengername", Type: "text"},
				{Name: "n_passengercount", Type: "integer"},
				{Name: "str_rowstatus", Type: "character", DefaultValue: "'A'::bpchar"},
			},
			Children: []ChildRelation{
				{
					Schema:     "trip",
					Table:      "tbl_trip_luggage",
					PrimaryKey: "id_tripluggage",
					FKColumn:   "id_tripparticipant",
					JSONKey:    "trip_luggage",
					IsOneToOne: true,
				},
			},
		},
	}

	names := testRepoNames(tables...)
	sqlText := buildInsertFunctionSQL(tables[1], names)
	if !strings.Contains(sqlText, "func_insert_tbl_trip_luggage_v000001") {
		t.Fatal("expected participant insert to call luggage insert")
	}

	if !strings.Contains(sqlText, "trip_luggage") {
		t.Fatal("expected participant insert to read trip_luggage object")
	}
}

func TestBuildDeleteFunctionSQLSoftDeletesChildrenFirst(t *testing.T) {
	tables := indexTables([]Table{
		{
			Schema:     "trip",
			Name:       "tbl_trip_stop",
			PrimaryKey: "id_tripstop",
			Columns: []Column{
				{Name: "id_tripstop", Type: "uuid", UdtName: "uuid"},
				{Name: "id_trip", Type: "uuid", UdtName: "uuid"},
				{Name: "str_rowstatus", Type: "character"},
			},
		},
		{
			Schema:     "trip",
			Name:       "tbl_trip",
			PrimaryKey: "id_trip",
			Columns: []Column{
				{Name: "id_trip", Type: "uuid", UdtName: "uuid"},
				{Name: "str_rowstatus", Type: "character"},
			},
			Children: []ChildRelation{
				{
					Schema:     "trip",
					Table:      "tbl_trip_stop",
					PrimaryKey: "id_tripstop",
					FKColumn:   "id_trip",
					JSONKey:    "trip_stops",
				},
			},
		},
	})

	flat := []Table{tables[tableKey("trip", "tbl_trip")], tables[tableKey("trip", "tbl_trip_stop")]}
	names := testRepoNames(flat...)
	sqlText := buildDeleteFunctionSQL(tables[tableKey("trip", "tbl_trip")], tables, names)
	if !strings.Contains(sqlText, "func_delete_tbl_trip_stop_v000001") {
		t.Fatal("expected trip delete to call stop delete")
	}

	if !strings.Contains(sqlText, `'D'`) {
		t.Fatal("expected soft delete on row status")
	}
}

func TestBuildGetFunctionSQLIncludesChildren(t *testing.T) {
	tables := indexTables([]Table{
		{
			Schema:     "trip",
			Name:       "tbl_trip_stop",
			PrimaryKey: "id_tripstop",
			Columns: []Column{
				{Name: "id_tripstop", Type: "uuid", UdtName: "uuid"},
				{Name: "id_trip", Type: "uuid", UdtName: "uuid"},
				{Name: "str_stoptype", Type: "text"},
				{Name: "str_rowstatus", Type: "character"},
			},
		},
		{
			Schema:     "trip",
			Name:       "tbl_trip_participant",
			PrimaryKey: "id_tripparticipant",
			Columns: []Column{
				{Name: "id_tripparticipant", Type: "uuid", UdtName: "uuid"},
				{Name: "id_trip", Type: "uuid", UdtName: "uuid"},
				{Name: "str_passengername", Type: "text"},
				{Name: "str_rowstatus", Type: "character"},
			},
			Children: []ChildRelation{
				{
					Schema:     "trip",
					Table:      "tbl_trip_luggage",
					PrimaryKey: "id_tripluggage",
					FKColumn:   "id_tripparticipant",
					JSONKey:    "trip_luggage",
					IsOneToOne: true,
				},
			},
		},
		{
			Schema:     "trip",
			Name:       "tbl_trip_luggage",
			PrimaryKey: "id_tripluggage",
			Columns: []Column{
				{Name: "id_tripluggage", Type: "uuid", UdtName: "uuid"},
				{Name: "id_tripparticipant", Type: "uuid", UdtName: "uuid"},
				{Name: "n_checkedluggagecount", Type: "integer"},
				{Name: "str_rowstatus", Type: "character"},
			},
		},
		{
			Schema:     "trip",
			Name:       "tbl_trip",
			PrimaryKey: "id_trip",
			Columns: []Column{
				{Name: "id_trip", Type: "uuid", UdtName: "uuid"},
				{Name: "str_triptype", Type: "text"},
				{Name: "str_rowstatus", Type: "character"},
			},
			Children: []ChildRelation{
				{
					Schema:     "trip",
					Table:      "tbl_trip_stop",
					PrimaryKey: "id_tripstop",
					FKColumn:   "id_trip",
					JSONKey:    "trip_stops",
				},
				{
					Schema:     "trip",
					Table:      "tbl_trip_participant",
					PrimaryKey: "id_tripparticipant",
					FKColumn:   "id_trip",
					JSONKey:    "trip_participants",
				},
			},
		},
	})

	flat := []Table{
		tables[tableKey("trip", "tbl_trip")],
		tables[tableKey("trip", "tbl_trip_stop")],
		tables[tableKey("trip", "tbl_trip_participant")],
		tables[tableKey("trip", "tbl_trip_luggage")],
	}
	names := testRepoNames(flat...)

	sqlText := buildGetFunctionSQL(tables[tableKey("trip", "tbl_trip")], tables, names)
	if !strings.Contains(sqlText, "func_get_tbl_trip_stop_v000001") {
		t.Fatal("expected trip get to load stops")
	}

	if !strings.Contains(sqlText, "func_get_tbl_trip_participant_v000001") {
		t.Fatal("expected trip get to load participants")
	}

	if !strings.Contains(sqlText, "trip_stops") {
		t.Fatal("expected trip_stops array in result")
	}

	participantSQL := buildGetFunctionSQL(tables[tableKey("trip", "tbl_trip_participant")], tables, names)
	if !strings.Contains(participantSQL, "func_get_tbl_trip_luggage_v000001") {
		t.Fatal("expected participant get to load luggage")
	}

	listSQL := buildListFunctionSQL(tables[tableKey("trip", "tbl_trip")], names)
	if !strings.Contains(listSQL, "func_get_tbl_trip_v000001") {
		t.Fatal("expected trip list to call trip get")
	}
	if !strings.Contains(listSQL, "jsonb_agg") {
		t.Fatal("expected trip list to aggregate rows")
	}
}

func TestPrefersSingleObjectJSON(t *testing.T) {
	if !prefersSingleObjectJSON("tbl_trip_luggage") {
		t.Fatal("expected luggage to be inserted as a single JSON object")
	}
}

func TestBuildRowToJSONObject(t *testing.T) {
	table := Table{
		PrimaryKey: "id_trip",
		Columns: []Column{
			{Name: "id_trip", Type: "uuid"},
			{Name: "str_triptype", Type: "text"},
		},
	}

	got := buildRowToJSONObject(table, "v_row")
	if !strings.Contains(got, "'id_trip'") || !strings.Contains(got, "v_row.\"id_trip\"") {
		t.Fatalf("unexpected row json builder: %s", got)
	}
}

func TestLoadConfigMergesImplicitExcludePatterns(t *testing.T) {
	cfg := Config{
		ExcludePatterns: []string{"tbl_%_staging"},
	}

	if err := cfg.Validate(); err != nil {
		t.Fatal(err)
	}

	patterns := cfg.allExcludePatterns()
	if len(patterns) != 3 {
		t.Fatalf("expected 3 exclude patterns, got %d", len(patterns))
	}

	if patterns[0] != "%_hist" || patterns[1] != "%_hist_detail" {
		t.Fatalf("unexpected implicit patterns: %#v", patterns[:2])
	}
}

func TestSortTablesChildrenFirstDetectsCycle(t *testing.T) {
	tables := []Table{
		{
			Schema: "trip",
			Name:   "tbl_a",
			Children: []ChildRelation{
				{Schema: "trip", Table: "tbl_b"},
			},
		},
		{
			Schema: "trip",
			Name:   "tbl_b",
			Children: []ChildRelation{
				{Schema: "trip", Table: "tbl_a"},
			},
		},
	}

	if _, err := sortTablesChildrenFirst(tables); err == nil {
		t.Fatal("expected cycle error")
	}
}
