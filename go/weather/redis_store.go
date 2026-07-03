package weather

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisStore struct {
	client    *redis.Client
	keyPrefix string
	ttl       time.Duration
}

func NewRedisStore(client *redis.Client, keyPrefix string, ttl time.Duration) *RedisStore {
	return &RedisStore{
		client:    client,
		keyPrefix: strings.TrimRight(strings.TrimSpace(keyPrefix), ":"),
		ttl:       ttl,
	}
}

func (s *RedisStore) SetDay(ctx context.Context, forecast DayForecast) error {
	body, err := json.Marshal(forecast)
	if err != nil {
		return err
	}
	return s.client.Set(ctx, s.Key(forecast.Date), body, s.ttl).Err()
}

func (s *RedisStore) GetDay(ctx context.Context, date string) (DayForecast, error) {
	raw, err := s.client.Get(ctx, s.Key(date)).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return DayForecast{}, ErrForecastNotFound.With("date", date)
		}
		return DayForecast{}, err
	}

	var forecast DayForecast
	if err := json.Unmarshal(raw, &forecast); err != nil {
		return DayForecast{}, err
	}
	return forecast, nil
}

func (s *RedisStore) Key(date string) string {
	return s.keyPrefix + ":" + strings.TrimSpace(date)
}
