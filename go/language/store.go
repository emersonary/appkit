package language

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

type Store struct {
	db     *sql.DB
	schema string
}

func NewStore(db *sql.DB, schema string) *Store {
	return &Store{db: db, schema: schema}
}

func (s *Store) table(name string) string {
	return qualifiedName(s.schema, name)
}

func (s *Store) ListLanguages(ctx context.Context) ([]Language, error) {
	query := fmt.Sprintf(`
		SELECT id_language, name, native_name, direction, is_default, created_at, updated_at
		FROM %s
		ORDER BY is_default DESC, id_language
	`, s.table("languages"))

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Language
	for rows.Next() {
		var item Language
		if err := rows.Scan(
			&item.IDLanguage,
			&item.Name,
			&item.NativeName,
			&item.Direction,
			&item.IsDefault,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, err
		}

		out = append(out, item)
	}

	return out, rows.Err()
}

func (s *Store) GetLanguage(ctx context.Context, code string) (Language, error) {
	query := fmt.Sprintf(`
		SELECT id_language, name, native_name, direction, is_default, created_at, updated_at
		FROM %s
		WHERE id_language = $1
	`, s.table("languages"))

	var item Language
	err := s.db.QueryRowContext(ctx, query, NormalizeCode(code)).Scan(
		&item.IDLanguage,
		&item.Name,
		&item.NativeName,
		&item.Direction,
		&item.IsDefault,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if err != nil {
		return Language{}, err
	}

	return item, nil
}

func (s *Store) LanguageExists(ctx context.Context, code string) (bool, error) {
	query := fmt.Sprintf(`SELECT EXISTS (SELECT 1 FROM %s WHERE id_language = $1)`, s.table("languages"))

	var exists bool
	if err := s.db.QueryRowContext(ctx, query, NormalizeCode(code)).Scan(&exists); err != nil {
		return false, err
	}

	return exists, nil
}

func IsNotFound(err error) bool {
	return err == sql.ErrNoRows || strings.Contains(err.Error(), "no rows")
}

func (s *Store) QualifiedLanguagesTable() string {
	return s.table("languages")
}
