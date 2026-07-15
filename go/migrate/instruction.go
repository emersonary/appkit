package migrate

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
)

// Instruction is one Up-only migration unit executed in registration order.
type Instruction struct {
	ID     string
	SQL    []string
	Source string
}

// Record is the persisted execution state for an instruction.
type Record struct {
	ID          string
	Source      string
	SQLChecksum string
	Status      string
	AppliedAt   *string
	Error       string
	DurationMS  int64
}

// Validate checks instruction fields before registration.
func (i Instruction) Validate() error {
	if strings.TrimSpace(i.ID) == "" {
		return fmt.Errorf("migrate: instruction ID is required")
	}
	if len(i.SQL) == 0 {
		return fmt.Errorf("migrate: instruction %q has no SQL", i.ID)
	}
	for idx, stmt := range i.SQL {
		if strings.TrimSpace(stmt) == "" {
			return fmt.Errorf("migrate: instruction %q SQL[%d] is empty", i.ID, idx)
		}
	}
	return nil
}

func checksumSQL(sql []string) string {
	h := sha256.Sum256([]byte(strings.Join(sql, "\n;\n")))
	return hex.EncodeToString(h[:])
}
