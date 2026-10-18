package uuidv7_test

import (
	"testing"
	"time"

	"github.com/emersonary/appkit/uuidv7"
)

func TestNewUUIDv7Ordered(t *testing.T) {
	a := uuidv7.NewUUIDv7()
	time.Sleep(2 * time.Millisecond)
	b := uuidv7.NewUUIDv7()
	if b.String() <= a.String() {
		t.Fatalf("expected lexicographic order: %s then %s", a, b)
	}
	if uuidv7.Version(a) != 7 || uuidv7.Version(b) != 7 {
		t.Fatalf("expected version 7, got %d and %d", uuidv7.Version(a), uuidv7.Version(b))
	}
}

func TestUUIDv7FromTimeStampRoundTrip(t *testing.T) {
	ts := time.Date(2024, 11, 15, 12, 0, 0, 123456789, time.UTC)
	id := uuidv7.UUIDv7FromTimeStamp(ts)
	got := uuidv7.TimeStampFromUUIDv7(id)
	if got.UnixMilli() != ts.UnixMilli() {
		t.Fatalf("millis mismatch: got %v want %v", got, ts)
	}
}
