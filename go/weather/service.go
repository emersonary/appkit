package weather

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/emersonary/appkit/apperror"
	"go.uber.org/zap"
)

type Service struct {
	cfg    AppConfig
	client *OpenMeteoClient
	store  *RedisStore
	logger *zap.Logger
}

type Options struct {
	Logger *zap.Logger
}

func NewService(cfg AppConfig, client *OpenMeteoClient, store *RedisStore, opts Options) (*Service, error) {
	cfg.normalize()
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	if client == nil {
		client = NewOpenMeteoClient(cfg.OpenMeteo, nil)
	}
	if store == nil {
		return nil, ErrRedisRequired
	}
	return &Service{
		cfg:    cfg,
		client: client,
		store:  store,
		logger: opts.Logger,
	}, nil
}

func (s *Service) Config() AppConfig {
	return s.cfg
}

func (s *Service) RefreshForecast(ctx context.Context) ([]DayForecast, error) {
	resp, err := s.client.FetchForecast(ctx)
	if err != nil {
		return nil, err
	}

	days, err := BuildDailyForecasts(resp, s.cfg, time.Now())
	if err != nil {
		return nil, err
	}

	for _, day := range days {
		if err := s.store.SetDay(ctx, day); err != nil {
			return days, err
		}
	}

	if s.logger != nil {
		s.logger.Info("open-meteo forecast refreshed",
			zap.Int("days", len(days)),
			zap.String("key_prefix", s.cfg.KeyPrefix),
		)
	}
	return days, nil
}

func (s *Service) GetDay(ctx context.Context, date time.Time) (DayForecast, error) {
	return s.store.GetDay(ctx, dayKey(date))
}

func (s *Service) GetDayOrRefresh(ctx context.Context, date time.Time) (DayForecast, error) {
	forecast, err := s.GetDay(ctx, date)
	if err == nil {
		return forecast, nil
	}
	if !isForecastNotFound(err) {
		return DayForecast{}, err
	}

	if _, err := s.RefreshForecast(ctx); err != nil {
		return DayForecast{}, err
	}
	return s.GetDay(ctx, date)
}

func (s *Service) GetWindSlotsOrRefresh(ctx context.Context, date time.Time) (WindSlotsForecast, error) {
	day, err := s.GetDayOrRefresh(ctx, date)
	if err != nil {
		return WindSlotsForecast{}, err
	}
	return BuildWindSlots(day), nil
}

func (s *Service) RunCollector(ctx context.Context, interval time.Duration) {
	if interval <= 0 {
		interval = time.Hour
	}

	run := func() {
		refreshCtx, cancel := context.WithTimeout(ctx, 45*time.Second)
		defer cancel()
		if _, err := s.RefreshForecast(refreshCtx); err != nil && s.logger != nil {
			s.logger.Error("open-meteo forecast refresh failed", zap.Error(err))
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

func dayKey(date time.Time) string {
	return date.Format("2006-01-02")
}

func isForecastNotFound(err error) bool {
	var appErr apperror.Error
	return errors.As(err, &appErr) && strings.EqualFold(appErr.Code, ErrForecastNotFound.Code)
}
