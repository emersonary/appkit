package weather

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const hourlyParams = "wind_speed_10m,wind_speed_80m,wind_speed_120m,wind_speed_180m,wind_direction_10m,wind_direction_80m,wind_direction_120m,wind_direction_180m,wind_gusts_10m,temperature_2m"

type OpenMeteoClient struct {
	cfg        OpenMeteoConfig
	httpClient *http.Client
}

func NewOpenMeteoClient(cfg OpenMeteoConfig, httpClient *http.Client) *OpenMeteoClient {
	cfg.normalize()
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 30 * time.Second}
	}
	return &OpenMeteoClient{cfg: cfg, httpClient: httpClient}
}

func (c *OpenMeteoClient) FetchForecast(ctx context.Context) (openMeteoResponse, error) {
	endpoint, err := url.Parse(c.cfg.BaseURL)
	if err != nil {
		return openMeteoResponse{}, wrapErr(ErrFetchForecast, "base_url", err)
	}

	q := endpoint.Query()
	q.Set("latitude", strconv.FormatFloat(c.cfg.Latitude, 'f', -1, 64))
	q.Set("longitude", strconv.FormatFloat(c.cfg.Longitude, 'f', -1, 64))
	q.Set("hourly", hourlyParams)
	q.Set("wind_speed_unit", strings.TrimSpace(c.cfg.WindSpeedUnit))
	q.Set("timezone", strings.TrimSpace(c.cfg.Timezone))
	q.Set("forecast_days", strconv.Itoa(c.cfg.ForecastDays))
	endpoint.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return openMeteoResponse{}, wrapErr(ErrFetchForecast, "request", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return openMeteoResponse{}, wrapErr(ErrFetchForecast, "http", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return openMeteoResponse{}, wrapErr(ErrFetchForecast, "read", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return openMeteoResponse{}, ErrFetchForecast.With("status", fmt.Sprintf("%d: %s", resp.StatusCode, truncate(body, 512)))
	}

	var forecast openMeteoResponse
	if err := json.Unmarshal(body, &forecast); err != nil {
		return openMeteoResponse{}, wrapErr(ErrFetchForecast, "decode", err)
	}
	return forecast, nil
}

func truncate(body []byte, max int) string {
	s := strings.TrimSpace(string(body))
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}
