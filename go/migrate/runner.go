package migrate

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// Runner applies registered Up-only SQL instructions and tracks execution state.
type Runner struct {
	db           *sql.DB
	store        *store
	instructions []Instruction
	byID         map[string]Instruction
}

// NewRunner creates a migration runner for the given database connection.
func NewRunner(db *sql.DB) *Runner {
	return &Runner{
		db:    db,
		store: newStore(db),
		byID:  make(map[string]Instruction),
	}
}

// Register adds instructions executed in registration order.
func (r *Runner) Register(instructions ...Instruction) error {
	for _, inst := range instructions {
		if err := inst.Validate(); err != nil {
			return err
		}
		if _, exists := r.byID[inst.ID]; exists {
			return fmt.Errorf("migrate: duplicate instruction ID %q", inst.ID)
		}
		r.instructions = append(r.instructions, inst)
		r.byID[inst.ID] = inst
	}
	return nil
}

// Apply runs pending instructions (Up-only).
func (r *Runner) Apply(ctx context.Context) error {
	if err := r.store.ensureMeta(ctx); err != nil {
		return fmt.Errorf("migrate bootstrap: %w", err)
	}

	for _, inst := range r.instructions {
		rec, err := r.store.getRecord(ctx, inst.ID)
		if err != nil {
			return err
		}
		sum := checksumSQL(inst.SQL)
		if rec != nil && rec.Status == "applied" {
			if rec.SQLChecksum != sum {
				return fmt.Errorf("migrate: instruction %q SQL changed after apply (checksum mismatch)", inst.ID)
			}
			continue
		}

		start := time.Now()
		if err := r.applyInstruction(ctx, inst); err != nil {
			if markErr := r.store.markFailed(ctx, inst, err); markErr != nil {
				return fmt.Errorf("migrate instruction %q: %w (record failure: %v)", inst.ID, err, markErr)
			}
			return fmt.Errorf("migrate instruction %q: %w", inst.ID, err)
		}
		if err := r.store.markApplied(ctx, inst, time.Since(start)); err != nil {
			return fmt.Errorf("migrate record %q: %w", inst.ID, err)
		}
	}
	return nil
}

func (r *Runner) applyInstruction(ctx context.Context, inst Instruction) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, stmt := range inst.SQL {
		if _, err := tx.ExecContext(ctx, stmt); err != nil {
			return err
		}
	}
	return tx.Commit()
}

// Status returns persisted instruction records.
func (r *Runner) Status(ctx context.Context) ([]Record, error) {
	if err := r.store.ensureMeta(ctx); err != nil {
		return nil, fmt.Errorf("migrate bootstrap: %w", err)
	}
	return r.store.listRecords(ctx)
}

// Instructions returns registered instructions in order.
func (r *Runner) Instructions() []Instruction {
	out := make([]Instruction, len(r.instructions))
	copy(out, r.instructions)
	return out
}
