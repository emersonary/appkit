package dbhist

import "testing"

func TestVersionedRepoFunctionName(t *testing.T) {
	got := versionedRepoFunctionName("tbl_trip", "insert", 1)
	if got != "func_insert_tbl_trip_v000001" {
		t.Fatalf("got %q", got)
	}

	got = versionedRepoFunctionName("tbl_trip", "get", 42)
	if got != "func_get_tbl_trip_v000042" {
		t.Fatalf("got %q", got)
	}
}

func TestParentInsertHashIncludesChildInsertHash(t *testing.T) {
	luggage := Table{
		Schema:     "trip",
		Name:       "tbl_trip_luggage",
		PrimaryKey: "id_tripluggage",
		Columns: []Column{
			{Name: "id_tripluggage", Type: "uuid", UdtName: "uuid"},
			{Name: "id_tripparticipant", Type: "uuid", UdtName: "uuid"},
			{Name: "n_checkedluggagecount", Type: "integer"},
		},
	}
	participant := Table{
		Schema:     "trip",
		Name:       "tbl_trip_participant",
		PrimaryKey: "id_tripparticipant",
		Columns: []Column{
			{Name: "id_tripparticipant", Type: "uuid", UdtName: "uuid"},
			{Name: "id_trip", Type: "uuid", UdtName: "uuid"},
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
	}
	tableMap := indexTables([]Table{luggage, participant})

	resolved := map[funcKey]resolvedFunc{}
	childSnap, err := buildRepoSnapshot(luggage, "insert", tableMap, resolved)
	if err != nil {
		t.Fatal(err)
	}
	childHash, _, err := hashRepoSnapshot(childSnap)
	if err != nil {
		t.Fatal(err)
	}
	resolved[makeFuncKey("trip", "tbl_trip_luggage", "insert")] = resolvedFunc{
		Version: 1,
		Name:    versionedRepoFunctionName("tbl_trip_luggage", "insert", 1),
		Hash:    childHash,
	}

	parentSnap, err := buildRepoSnapshot(participant, "insert", tableMap, resolved)
	if err != nil {
		t.Fatal(err)
	}
	if len(parentSnap.Calls) != 1 || parentSnap.Calls[0].Hash != childHash {
		t.Fatalf("expected parent insert to call child hash, got %#v", parentSnap.Calls)
	}

	parentHash, _, err := hashRepoSnapshot(parentSnap)
	if err != nil {
		t.Fatal(err)
	}

	resolved[makeFuncKey("trip", "tbl_trip_luggage", "insert")] = resolvedFunc{
		Version: 2,
		Name:    versionedRepoFunctionName("tbl_trip_luggage", "insert", 2),
		Hash:    "changed-child-hash",
	}
	parentSnap2, err := buildRepoSnapshot(participant, "insert", tableMap, resolved)
	if err != nil {
		t.Fatal(err)
	}
	parentHash2, _, err := hashRepoSnapshot(parentSnap2)
	if err != nil {
		t.Fatal(err)
	}
	if parentHash == parentHash2 {
		t.Fatal("expected parent hash to change when child hash changes")
	}
}
