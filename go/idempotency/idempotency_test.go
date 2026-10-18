package idempotency

import (
	"errors"
	"testing"
	"time"
)

func TestFingerprintStable(t *testing.T) {
	a, err := Fingerprint(map[string]int{"b": 2, "a": 1})
	if err != nil {
		t.Fatal(err)
	}
	b, err := Fingerprint(map[string]int{"a": 1, "b": 2})
	if err != nil {
		t.Fatal(err)
	}
	if a != b {
		t.Fatal("fingerprint not deterministic")
	}
}

func TestDecode(t *testing.T) {
	req := Request{Scope: "x", Key: "y", Fingerprint: "abc"}
	now := time.Now().UTC()
	r := record{fingerprint: "abc", status: "COMPLETED", result: []byte(`{"id":1}`), started: now}
	r.completed.Time = now
	r.completed.Valid = true
	got, err := decode[struct {
		ID int `json:"id"`
	}](req, r)
	if err != nil || got.Value.ID != 1 || !got.Cached {
		t.Fatalf("got=%+v err=%v", got, err)
	}
	r.fingerprint = "other"
	_, err = decode[string](req, r)
	if !errors.Is(err, ErrConflict) {
		t.Fatal(err)
	}
}
