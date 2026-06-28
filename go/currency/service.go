package currency

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"go.uber.org/zap"
)

type Options struct {
	Logger *zap.Logger
	APIURL string
}

func (o Options) normalized() Options {
	out := o
	if out.APIURL == "" {
		out.APIURL = DefaultAPIURL
	}

	return out
}

type Service struct {
	cfg     Config
	enabled map[string]struct{}
	store   *Store
	client  *Client
	opts    Options
}

func NewService(db *sql.DB, cfg Config, opts Options) (*Service, error) {
	cfg.normalize()
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	opts = opts.normalized()

	return &Service{
		cfg:     cfg,
		enabled: cfg.EnabledSet(),
		store:   NewStore(db, cfg.Schema),
		client:  NewClient(opts.APIURL),
		opts:    opts,
	}, nil
}

func (s *Service) Config() Config {
	return s.cfg
}

func (s *Service) ValidateCode(code string) error {
	normalized := NormalizeCode(code)
	if normalized == "" {
		return ErrUnknownISO4217.With("code", code)
	}

	if err := validateISO4217Code(normalized); err != nil {
		return err
	}

	if _, ok := s.enabled[normalized]; !ok {
		return ErrUnknownCurrency.With("code", normalized)
	}

	return nil
}

func (s *Service) Store() *Store {
	return s.store
}

func (s *Service) SyncExchangeRates(ctx context.Context) (SyncResult, error) {
	snapshot, err := s.client.FetchUSDRates(ctx)
	if err != nil {
		return SyncResult{}, err
	}

	result := SyncResult{Source: snapshot.Source}
	for _, code := range s.cfg.Currencies {
		if code == s.cfg.BaseCurrency {
			continue
		}

		if err := s.ValidateCode(code); err != nil {
			return result, err
		}

		rate, ok := snapshot.Rates[code]
		if !ok || rate <= 0 {
			result.Skipped++
			if s.opts.Logger != nil {
				s.opts.Logger.Warn("currency rate missing from feed",
					zap.String("code", code),
					zap.String("feed", snapshot.Source),
				)
			}

			continue
		}

		if err := s.store.UpsertExchangeRate(ctx, code, rate, s.cfg.BaseCurrency, snapshot.FetchedAt); err != nil {
			return result, wrapErr(ErrSyncRates, "upsert", err)
		}
		if err := s.store.InsertExchangeRateHistory(ctx, code, rate, s.cfg.BaseCurrency, snapshot.FetchedAt); err != nil {
			return result, wrapErr(ErrSyncRates, "history", err)
		}

		result.Updated++
	}

	if s.opts.Logger != nil {
		s.opts.Logger.Info("currency rates synced",
			zap.Int("updated", result.Updated),
			zap.Int("skipped", result.Skipped),
			zap.String("feed", snapshot.Source),
		)
	}

	return result, nil
}

func (s *Service) RunExchangeRateUpdater(ctx context.Context, interval time.Duration) {
	if interval <= 0 {
		interval = time.Hour
	}

	run := func() {
		syncCtx, cancel := context.WithTimeout(ctx, 2*time.Minute)
		defer cancel()

		if _, err := s.SyncExchangeRates(syncCtx); err != nil && s.opts.Logger != nil {
			s.opts.Logger.Error("currency rate sync failed", zap.Error(err))
		}
	}

	run()

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			run()
		}
	}
}

func (s *Service) Convert(ctx context.Context, amount float64, fromCode, toCode string) (float64, error) {
	if amount <= 0 {
		return 0, ErrInvalidAmount
	}

	fromCode = NormalizeCode(fromCode)
	toCode = NormalizeCode(toCode)

	if fromCode == toCode {
		return 0, ErrSameCurrency
	}

	if err := s.ValidateCode(fromCode); err != nil {
		return 0, err
	}
	if err := s.ValidateCode(toCode); err != nil {
		return 0, err
	}

	usdAmount, err := s.toBaseAmount(ctx, amount, fromCode)
	if err != nil {
		return 0, err
	}

	return s.fromBaseAmount(ctx, usdAmount, toCode)
}

func (s *Service) toBaseAmount(ctx context.Context, amount float64, fromCode string) (float64, error) {
	if fromCode == s.cfg.BaseCurrency {
		return amount, nil
	}

	rate, err := s.store.GetExchangeRate(ctx, fromCode)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, ErrRateNotFound.With("code", fromCode)
		}

		return 0, wrapErr(ErrSyncRates, "get_rate", err)
	}

	return amount / rate.Rate, nil
}

func (s *Service) fromBaseAmount(ctx context.Context, usdAmount float64, toCode string) (float64, error) {
	if toCode == s.cfg.BaseCurrency {
		return usdAmount, nil
	}

	rate, err := s.store.GetExchangeRate(ctx, toCode)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, ErrRateNotFound.With("code", toCode)
		}

		return 0, wrapErr(ErrSyncRates, "get_rate", err)
	}

	return usdAmount * rate.Rate, nil
}

func (s *Service) ListCurrencies(ctx context.Context) ([]Currency, error) {
	items, err := s.store.ListCurrencies(ctx)
	if err != nil {
		return nil, err
	}

	filtered := make([]Currency, 0, len(items))
	for _, item := range items {
		if err := s.ValidateCode(item.IDCurrency); err != nil {
			continue
		}

		filtered = append(filtered, item)
	}

	return filtered, nil
}

func (s *Service) ListExchangeRates(ctx context.Context) ([]ExchangeRate, error) {
	items, err := s.store.ListExchangeRates(ctx)
	if err != nil {
		return nil, err
	}

	filtered := make([]ExchangeRate, 0, len(items))
	for _, item := range items {
		if err := s.ValidateCode(item.IDCurrency); err != nil {
			continue
		}

		filtered = append(filtered, item)
	}

	return filtered, nil
}

func (s *Service) GetExchangeRate(ctx context.Context, code string) (ExchangeRate, error) {
	if err := s.ValidateCode(code); err != nil {
		return ExchangeRate{}, err
	}

	if code == s.cfg.BaseCurrency {
		return ExchangeRate{}, ErrRateNotFound.With("code", code)
	}

	item, err := s.store.GetExchangeRate(ctx, code)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ExchangeRate{}, ErrRateNotFound.With("code", code)
		}

		return ExchangeRate{}, wrapErr(ErrSyncRates, "get_rate", err)
	}

	return item, nil
}
