# Weather

Hourly Open-Meteo collector that stores low-wind forecast slices in Redis.

## What It Stores

The collector fetches the configured Open-Meteo forecast, filters hours where
`wind_speed_10m <= low_wind_max_knots`, splits the result by local day, and
stores one Redis key per day:

```text
weather:openmeteo:jeri:2026-07-03
weather:openmeteo:jeri:2026-07-04
```

Each value is JSON with the low-wind hours and wind speed/direction at 10m,
80m, 120m, and 180m, plus 10m gusts and 2m temperature.

## Config

```yaml
redis:
  enabled: true
  addr: localhost:6379
  password: ""
  db: 0

weather:
  enabled: true
  key_prefix: weather:openmeteo:jeri
  refresh_interval: 1h
  cache_ttl: 240h
  low_wind_max_knots: 12
  open_meteo:
    latitude: -2.795
    longitude: -40.514
    forecast_days: 7
    wind_speed_unit: kn
    timezone: auto
```

See `weather.example.yaml` for the full standalone block shape.

## Runtime

When enabled, runtime starts an immediate refresh and then runs the collector
every `refresh_interval`. Future endpoints can call:

```go
forecast, err := app.Weather.GetDayOrRefresh(ctx, date)
```

If Redis has no key for that day, `GetDayOrRefresh` fetches Open-Meteo, rewrites
the daily Redis keys, and then returns the requested day.
