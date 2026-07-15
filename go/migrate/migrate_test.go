package migrate_test

import (
	"testing"

	"github.com/emersonary/appkit/migrate"
)

func TestInstructionValidate(t *testing.T) {
	t.Parallel()
	if err := (migrate.Instruction{}).Validate(); err == nil {
		t.Fatal("expected error for empty instruction")
	}
	if err := (migrate.Instruction{ID: "x", SQL: []string{"select 1"}}).Validate(); err != nil {
		t.Fatalf("valid instruction: %v", err)
	}
}

func TestRegisterDuplicateID(t *testing.T) {
	t.Parallel()
	runner := migrate.NewRunner(nil)
	if err := runner.Register(migrate.Instruction{ID: "a", SQL: []string{"select 1"}}); err != nil {
		t.Fatal(err)
	}
	if err := runner.Register(migrate.Instruction{ID: "a", SQL: []string{"select 2"}}); err == nil {
		t.Fatal("expected duplicate ID error")
	}
}
