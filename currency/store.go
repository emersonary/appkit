package currency

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
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

func (s *Store) ListCurrencies(ctx context.Context) ([]Currency, error) {
	query := fmt.Sprintf(`
		SELECT code, name, symbol, is_base, website_languages, created_at, updated_at
		FROM %s
		ORDER BY is_base DESC, code
	`, s.table("currencies"))

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Currency
	for rows.Next() {
		var item Currency
		var languages pgtype.FlatArray[string]

		if err := rows.Scan(
			&item.Code,
			&item.Name,
			&item.Symbol,
			&item.IsBase,
			&languages,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, err
		}

		item.WebsiteLanguages = []string(languages)
		out = append(out, item)
	}

	return out, rows.Err()
}

func (s *Store) ListTrackedCurrencyCodes(ctx context.Context) ([]string, error) {
	query := fmt.Sprintf(`
		SELECT code FROM %s WHERE NOT is_base ORDER BY code
	`, s.table("currencies"))

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var codes []string
	for rows.Next() {
		var code string
		if err := rows.Scan(&code); err != nil {
			return nil, err
		}

		codes = append(codes, code)
	}

	return codes, rows.Err()
}

func (s *Store) CurrencyExists(ctx context.Context, code string) (bool, error) {
	query := fmt.Sprintf(`SELECT EXISTS (SELECT 1 FROM %s WHERE code = $1)`, s.table("currencies"))

	var exists bool
	if err := s.db.QueryRowContext(ctx, query, strings.ToUpper(strings.TrimSpace(code))).Scan(&exists); err != nil {
		return false, err
	}

	return exists, nil
}

func (s *Store) UpsertExchangeRate(ctx context.Context, currencyCode string, rate float64, source string, fetchedAt time.Time) error {
	query := fmt.Sprintf(`
		INSERT INTO %s (currency_code, rate, source, fetched_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (currency_code) DO UPDATE SET
			rate = EXCLUDED.rate,
			source = EXCLUDED.source,
			fetched_at = EXCLUDED.fetched_at
	`, s.table("currency_exchange_rates"))

	_, err := s.db.ExecContext(ctx, query, currencyCode, rate, source, fetchedAt)
	if err != nil {
		return fmt.Errorf("upsert exchange rate %s: %w", currencyCode, err)
	}

	return nil
}

func (s *Store) InsertExchangeRateHistory(ctx context.Context, currencyCode string, rate float64, source string, recordedAt time.Time) error {
	query := fmt.Sprintf(`
		INSERT INTO %s (currency_code, rate, source, recorded_at)
		VALUES ($1, $2, $3, $4)
	`, s.table("currency_exchange_rate_history"))

	_, err := s.db.ExecContext(ctx, query, currencyCode, rate, source, recordedAt)
	if err != nil {
		return fmt.Errorf("insert exchange rate history %s: %w", currencyCode, err)
	}

	return nil
}

func (s *Store) GetExchangeRate(ctx context.Context, currencyCode string) (ExchangeRate, error) {
	query := fmt.Sprintf(`
		SELECT currency_code, rate, source, fetched_at
		FROM %s
		WHERE currency_code = $1
	`, s.table("currency_exchange_rates"))

	var item ExchangeRate
	err := s.db.QueryRowContext(ctx, query, currencyCode).Scan(
		&item.CurrencyCode,
		&item.Rate,
		&item.Source,
		&item.FetchedAt,
	)
	if err != nil {
		return ExchangeRate{}, err
	}

	return item, nil
}

func (s *Store) ListExchangeRates(ctx context.Context) ([]ExchangeRate, error) {
	query := fmt.Sprintf(`
		SELECT currency_code, rate, source, fetched_at
		FROM %s
		ORDER BY currency_code
	`, s.table("currency_exchange_rates"))

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []ExchangeRate
	for rows.Next() {
		var item ExchangeRate
		if err := rows.Scan(&item.CurrencyCode, &item.Rate, &item.Source, &item.FetchedAt); err != nil {
			return nil, err
		}

		out = append(out, item)
	}

	return out, rows.Err()
}
