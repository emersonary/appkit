package accounts

import (
	"context"
	"testing"
)

func TestWire_Disabled(t *testing.T) {
	svc, err := Wire(context.Background(), nil, AppConfig{Enabled: false}, Options{})
	if err != nil {
		t.Fatalf("Wire: %v", err)
	}
	if svc != nil {
		t.Fatal("expected nil service when disabled")
	}
}
