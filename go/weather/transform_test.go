package weather

import (
	"testing"
	"time"
)

func TestBuildDailyForecastsSplitsLowWindByDay(t *testing.T) {
	var resp openMeteoResponse
	resp.Latitude = -2.795
	resp.Longitude = -40.514
	resp.Timezone = "America/Fortaleza"
	resp.Hourly.Time = []string{
		"2026-07-03T08:00",
		"2026-07-03T09:00",
		"2026-07-04T08:00",
	}
	resp.Hourly.WindSpeed10m = []float64{8, 14, 6}
	resp.Hourly.WindSpeed80m = []float64{10, 16, 8}
	resp.Hourly.WindSpeed120m = []float64{11, 17, 9}
	resp.Hourly.WindSpeed180m = []float64{12, 18, 10}
	resp.Hourly.WindDirection10m = []float64{120, 130, 140}
	resp.Hourly.WindDirection80m = []float64{121, 131, 141}
	resp.Hourly.WindDirection120m = []float64{122, 132, 142}
	resp.Hourly.WindDirection180m = []float64{123, 133, 143}
	resp.Hourly.WindGusts10m = []float64{10, 16, 9}
	resp.Hourly.Temperature2m = []float64{26, 27, 25}

	cfg := AppConfig{Enabled: true, LowWindMaxKnots: 12}
	days, err := BuildDailyForecasts(resp, cfg, time.Date(2026, 7, 3, 14, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}

	if len(days) != 2 {
		t.Fatalf("expected 2 days, got %d", len(days))
	}
	if days[0].Date != "2026-07-03" || len(days[0].LowWindHours) != 1 {
		t.Fatalf("unexpected first day: %+v", days[0])
	}
	if days[1].Date != "2026-07-04" || len(days[1].LowWindHours) != 1 {
		t.Fatalf("unexpected second day: %+v", days[1])
	}
	if days[0].LowWindHours[0].Wind10m.DirectionDeg != 120 {
		t.Fatalf("expected direction 120, got %f", days[0].LowWindHours[0].Wind10m.DirectionDeg)
	}
}

func TestBuildDailyForecastsAppliesOptionalGustLimit(t *testing.T) {
	var resp openMeteoResponse
	resp.Timezone = "America/Fortaleza"
	resp.Hourly.Time = []string{"2026-07-03T08:00"}
	resp.Hourly.WindSpeed10m = []float64{8}
	resp.Hourly.WindSpeed80m = []float64{8}
	resp.Hourly.WindSpeed120m = []float64{8}
	resp.Hourly.WindSpeed180m = []float64{8}
	resp.Hourly.WindDirection10m = []float64{120}
	resp.Hourly.WindDirection80m = []float64{120}
	resp.Hourly.WindDirection120m = []float64{120}
	resp.Hourly.WindDirection180m = []float64{120}
	resp.Hourly.WindGusts10m = []float64{18}
	resp.Hourly.Temperature2m = []float64{26}

	cfg := AppConfig{Enabled: true, LowWindMaxKnots: 12, LowGustMaxKnots: 16}
	days, err := BuildDailyForecasts(resp, cfg, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if len(days) != 1 {
		t.Fatalf("expected day key with empty low-wind hours, got %+v", days)
	}
	if len(days[0].LowWindHours) != 0 {
		t.Fatalf("expected gust-limited hour to be excluded, got %+v", days[0].LowWindHours)
	}
}
