package permissions

import (
	"context"
	"database/sql"
	"fmt"
)

type WireOptions struct{}

func ApplySchemaFromSetupInput(ctx context.Context, db *sql.DB, input SetupInput) error {
	if !input.Enabled {
		return nil
	}

	setup, err := ResolveSetup(input)
	if err != nil {
		return fmt.Errorf("resolve permissions setup: %w", err)
	}

	if err := ApplySchema(ctx, db, setup); err != nil {
		return fmt.Errorf("apply permissions schema: %w", err)
	}

	return nil
}

func Wire(ctx context.Context, db *sql.DB, input SetupInput, _ WireOptions) (*Service, error) {
	if !input.Enabled {
		return nil, nil
	}

	setup, err := ResolveSetup(input)
	if err != nil {
		return nil, fmt.Errorf("resolve permissions setup: %w", err)
	}

	if err := ApplySchema(ctx, db, setup); err != nil {
		return nil, fmt.Errorf("apply permissions schema: %w", err)
	}

	svc, err := NewService(db, setup)
	if err != nil {
		return nil, err
	}

	return svc, nil
}
