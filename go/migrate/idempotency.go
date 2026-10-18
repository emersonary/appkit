package migrate

import "github.com/emersonary/appkit/idempotency"

const (
	InstructionIdempotencySchema = "appkit.idempotency.schema.v1"
	InstructionIdempotencyTables = "appkit.idempotency.tables.v1"
	InstructionIdempotencyPost   = "appkit.idempotency.post.v1"
)

// IdempotencyInstructions returns DDL for idempotency.tbl_idempotency_records.
func IdempotencyInstructions() []Instruction {
	return []Instruction{
		{
			ID:     InstructionIdempotencySchema,
			SQL:    []string{idempotency.MigrationExtrasSQL},
			Source: "idempotency/migrate_extras.sql",
		},
		{
			ID:     InstructionIdempotencyTables,
			SQL:    []string{idempotency.MigrationSQL},
			Source: "idempotency/migration.gen.sql",
		},
		{
			ID:     InstructionIdempotencyPost,
			SQL:    []string{idempotency.MigrationPostSQL},
			Source: "idempotency/migrate_post.sql",
		},
	}
}
