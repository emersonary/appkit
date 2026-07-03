package weather

import (
	"testing"
	"time"
)

func TestBuildWindSlotsAggregatesMinMaxDirectionAndStars(t *testing.T) {
	loc := time.FixedZone("test", -3*60*60)
	day := DayForecast{
		Date:      "2026-07-03",
		UpdatedAt: time.Date(2026, 7, 3, 14, 0, 0, 0, time.UTC),
		LowWindHours: []LowWindHour{
			{
				Time:     time.Date(2026, 7, 3, 6, 0, 0, 0, loc),
				Wind10m:  WindSample{SpeedKnots: 11, DirectionDeg: 90},
				Wind80m:  WindSample{SpeedKnots: 12},
				Wind120m: WindSample{SpeedKnots: 14},
			},
			{
				Time:     time.Date(2026, 7, 3, 7, 0, 0, 0, loc),
				Wind10m:  WindSample{SpeedKnots: 14, DirectionDeg: 100},
				Wind80m:  WindSample{SpeedKnots: 16},
				Wind120m: WindSample{SpeedKnots: 18},
			},
			{
				Time:     time.Date(2026, 7, 3, 8, 0, 0, 0, loc),
				Wind10m:  WindSample{SpeedKnots: 12, DirectionDeg: 110},
				Wind80m:  WindSample{SpeedKnots: 14},
				Wind120m: WindSample{SpeedKnots: 16},
			},
		},
	}

	forecast := BuildWindSlots(day)
	if len(forecast.Slots) != 4 {
		t.Fatalf("expected 4 slots, got %d", len(forecast.Slots))
	}
	morning := forecast.Slots[0]
	if morning.MinKnots != 12.7 || morning.MaxKnots != 16.7 {
		t.Fatalf("unexpected min/max: %+v", morning)
	}
	if morning.DirectionCompass != "E" {
		t.Fatalf("expected E direction, got %q (%f)", morning.DirectionCompass, morning.DirectionDeg)
	}
	if morning.Stars != 1 {
		t.Fatalf("expected 1 star, got %d", morning.Stars)
	}
}

func TestStarsForMinWind(t *testing.T) {
	tests := []struct {
		min  float64
		want int
	}{
		{9.9, 0},
		{10, 0},
		{11, 1},
		{14, 1},
		{15, 2},
		{18, 2},
		{19, 3},
		{35, 5},
	}
	for _, tt := range tests {
		if got := StarsForMinWind(tt.min); got != tt.want {
			t.Fatalf("StarsForMinWind(%f): got %d want %d", tt.min, got, tt.want)
		}
	}
}
