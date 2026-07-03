package weather

import (
	"math"
	"time"
)

type WindSlotID string

const (
	WindSlotMorning   WindSlotID = "morning"
	WindSlotMidday    WindSlotID = "midday"
	WindSlotAfternoon WindSlotID = "afternoon"
	WindSlotLateDay   WindSlotID = "lateDay"
)

type WindSlot struct {
	ID               WindSlotID `json:"id"`
	TimeRange        string     `json:"time_range"`
	MinKnots         float64    `json:"min_knots"`
	MaxKnots         float64    `json:"max_knots"`
	DirectionDeg     float64    `json:"direction_deg"`
	DirectionCompass string     `json:"direction_compass"`
	Stars            int        `json:"stars"`
	HourCount        int        `json:"hour_count"`
}

type WindSlotsForecast struct {
	Date      string     `json:"date"`
	UpdatedAt time.Time  `json:"updated_at"`
	Slots     []WindSlot `json:"slots"`
}

type windSlotWindow struct {
	id    WindSlotID
	label string
	start int
	end   int
}

var defaultWindSlotWindows = []windSlotWindow{
	{id: WindSlotMorning, label: "06:00 - 09:00", start: 6, end: 9},
	{id: WindSlotMidday, label: "09:00 - 12:00", start: 9, end: 12},
	{id: WindSlotAfternoon, label: "12:00 - 15:00", start: 12, end: 15},
	{id: WindSlotLateDay, label: "15:00 - 18:00", start: 15, end: 18},
}

func BuildWindSlots(day DayForecast) WindSlotsForecast {
	slots := make([]WindSlot, 0, len(defaultWindSlotWindows))
	for _, window := range defaultWindSlotWindows {
		slots = append(slots, buildWindSlot(day, window))
	}
	return WindSlotsForecast{
		Date:      day.Date,
		UpdatedAt: day.UpdatedAt,
		Slots:     slots,
	}
}

func buildWindSlot(day DayForecast, window windSlotWindow) WindSlot {
	slot := WindSlot{
		ID:        window.id,
		TimeRange: window.label,
	}

	minKnots := math.Inf(1)
	maxKnots := math.Inf(-1)
	var sinSum, cosSum float64

	for _, hour := range day.LowWindHours {
		h := hour.Time.Hour()
		if h < window.start || h >= window.end {
			continue
		}

		speed := EstimateWindGuru(
			hour.Wind10m.SpeedKnots,
			hour.Wind80m.SpeedKnots,
			hour.Wind120m.SpeedKnots,
			hour.Wind180m.SpeedKnots,
		)
		if speed < minKnots {
			minKnots = speed
		}
		if speed > maxKnots {
			maxKnots = speed
		}

		radians := hour.Wind10m.DirectionDeg * math.Pi / 180
		sinSum += math.Sin(radians)
		cosSum += math.Cos(radians)
		slot.HourCount++
	}

	if slot.HourCount == 0 {
		return slot
	}

	slot.MinKnots = round1(minKnots)
	slot.MaxKnots = round1(maxKnots)
	slot.DirectionDeg = round1(normalizeDegrees(math.Atan2(sinSum, cosSum) * 180 / math.Pi))
	slot.DirectionCompass = compassDirection(slot.DirectionDeg)
	slot.Stars = StarsForMinWind(slot.MinKnots)
	return slot
}

// StarsForMinWind maps minimum slot wind to a 0-5 rating.
// 10kn or below shows no stars; 11-14 = 1, 15-18 = 2, etc.
func StarsForMinWind(minKnots float64) int {
	stars := int(math.Ceil((minKnots - 10) / 4))
	if stars < 0 {
		return 0
	}
	if stars > 5 {
		return 5
	}
	return stars
}

func normalizeDegrees(deg float64) float64 {
	for deg < 0 {
		deg += 360
	}
	for deg >= 360 {
		deg -= 360
	}
	return deg
}

func compassDirection(deg float64) string {
	directions := []string{"N", "NNE", "NE", "ENE", "E", "ESE", "SE", "SSE", "S", "SSW", "SW", "WSW", "W", "WNW", "NW", "NNW"}
	idx := int(math.Round(deg/22.5)) % len(directions)
	return directions[idx]
}

func round1(v float64) float64 {
	return math.Round(v*10) / 10
}
