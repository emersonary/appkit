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
		SELECT id_currency, name, symbol, is_base, website_languages, created_at, updated_at
		FROM %s
		ORDER BY is_base DESC, id_currency
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
			&item.IDCurrency,
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
		SELECT id_currency FROM %s WHERE NOT is_base ORDER BY id_currency
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
	query := fmt.Sprintf(`SELECT EXISTS (SELECT 1 FROM %s WHERE id_currency = $1)`, s.table("currencies"))

	var exists bool
	if err := s.db.QueryRowContext(ctx, query, strings.ToUpper(strings.TrimSpace(code))).Scan(&exists); err != nil {
		return false, err
	}

	return exists, nil
}

func (s *Store) UpsertExchangeRate(ctx context.Context, currencyCode string, rate float64, source string, fetchedAt time.Time) error {
	query := fmt.Sprintf(`
		INSERT INTO %s (id_currency, rate, source, fetched_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (id_currency) DO UPDATE SET
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
		INSERT INTO %s (id_currency, rate, source, recorded_at)
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
		SELECT id_currency, rate, source, fetched_at
		FROM %s
		WHERE id_currency = $1
	`, s.table("currency_exchange_rates"))

	var item ExchangeRate
	err := s.db.QueryRowContext(ctx, query, currencyCode).Scan(
		&item.IDCurrency,
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
		SELECT id_currency, rate, source, fetched_at
		FROM %s
		ORDER BY id_currency
	`, s.table("currency_exchange_rates"))

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []ExchangeRate
	for rows.Next() {
		var item ExchangeRate
		if err := rows.Scan(&item.IDCurrency, &item.Rate, &item.Source, &item.FetchedAt); err != nil {
			return nil, err
		}

		out = append(out, item)
	}

	return out, rows.Err()
}

func (s *Store) QualifiedCurrenciesTable() string {
	return s.table("currencies")
}
