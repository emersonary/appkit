package weather

import "time"

const openMeteoSource = "open-meteo"

type WindSample struct {
	SpeedKnots   float64 `json:"speed_knots"`
	DirectionDeg float64 `json:"direction_deg"`
}

type LowWindHour struct {
	Time          time.Time  `json:"time"`
	Wind10m       WindSample `json:"wind_10m"`
	Wind80m       WindSample `json:"wind_80m"`
	Wind120m      WindSample `json:"wind_120m"`
	Wind180m      WindSample `json:"wind_180m"`
	Gusts10mKnots float64    `json:"gusts_10m_knots"`
	TemperatureC  float64    `json:"temperature_2m_c"`
}

type DayForecast struct {
	Date         string        `json:"date"`
	Source       string        `json:"source"`
	Latitude     float64       `json:"latitude"`
	Longitude    float64       `json:"longitude"`
	Timezone     string        `json:"timezone"`
	UpdatedAt    time.Time     `json:"updated_at"`
	LowWindHours []LowWindHour `json:"low_wind_hours"`
}

type openMeteoResponse struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Timezone  string  `json:"timezone"`
	Hourly    struct {
		Time              []string  `json:"time"`
		WindSpeed10m      []float64 `json:"wind_speed_10m"`
		WindSpeed80m      []float64 `json:"wind_speed_80m"`
		WindSpeed120m     []float64 `json:"wind_speed_120m"`
		WindSpeed180m     []float64 `json:"wind_speed_180m"`
		WindDirection10m  []float64 `json:"wind_direction_10m"`
		WindDirection80m  []float64 `json:"wind_direction_80m"`
		WindDirection120m []float64 `json:"wind_direction_120m"`
		WindDirection180m []float64 `json:"wind_direction_180m"`
		WindGusts10m      []float64 `json:"wind_gusts_10m"`
		Temperature2m     []float64 `json:"temperature_2m"`
	} `json:"hourly"`
}
