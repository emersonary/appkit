package weather

import (
	"sort"
	"time"
)

func BuildDailyForecasts(resp openMeteoResponse, cfg AppConfig, updatedAt time.Time) ([]DayForecast, error) {
	cfg.normalize()
	location := time.UTC
	if resp.Timezone != "" {
		if loc, err := time.LoadLocation(resp.Timezone); err == nil {
			location = loc
		}
	}

	daysByDate := make(map[string]*DayForecast)
	for i, rawTime := range resp.Hourly.Time {
		if !hasHourlyIndex(resp, i) {
			continue
		}
		t, err := parseOpenMeteoTime(rawTime, location)
		if err != nil {
			return nil, err
		}
		date := t.Format("2006-01-02")
		day := daysByDate[date]
		if day == nil {
			day = &DayForecast{
				Date:      date,
				Source:    openMeteoSource,
				Latitude:  resp.Latitude,
				Longitude: resp.Longitude,
				Timezone:  resp.Timezone,
				UpdatedAt: updatedAt,
			}
			daysByDate[date] = day
		}

		wind10m := resp.Hourly.WindSpeed10m[i]
		gust10m := resp.Hourly.WindGusts10m[i]
		if !isLowWind(wind10m, gust10m, cfg) {
			continue
		}

		day.LowWindHours = append(day.LowWindHours, LowWindHour{
			Time: t,
			Wind10m: WindSample{
				SpeedKnots:   wind10m,
				DirectionDeg: resp.Hourly.WindDirection10m[i],
			},
			Wind80m: WindSample{
				SpeedKnots:   resp.Hourly.WindSpeed80m[i],
				DirectionDeg: resp.Hourly.WindDirection80m[i],
			},
			Wind120m: WindSample{
				SpeedKnots:   resp.Hourly.WindSpeed120m[i],
				DirectionDeg: resp.Hourly.WindDirection120m[i],
			},
			Wind180m: WindSample{
				SpeedKnots:   resp.Hourly.WindSpeed180m[i],
				DirectionDeg: resp.Hourly.WindDirection180m[i],
			},
			Gusts10mKnots: gust10m,
			TemperatureC:  resp.Hourly.Temperature2m[i],
		})
	}

	dates := make([]string, 0, len(daysByDate))
	for date := range daysByDate {
		dates = append(dates, date)
	}
	sort.Strings(dates)

	out := make([]DayForecast, 0, len(dates))
	for _, date := range dates {
		day := daysByDate[date]
		sort.Slice(day.LowWindHours, func(i, j int) bool {
			return day.LowWindHours[i].Time.Before(day.LowWindHours[j].Time)
		})
		out = append(out, *day)
	}
	return out, nil
}

func hasHourlyIndex(resp openMeteoResponse, i int) bool {
	return i < len(resp.Hourly.WindSpeed10m) &&
		i < len(resp.Hourly.WindSpeed80m) &&
		i < len(resp.Hourly.WindSpeed120m) &&
		i < len(resp.Hourly.WindSpeed180m) &&
		i < len(resp.Hourly.WindDirection10m) &&
		i < len(resp.Hourly.WindDirection80m) &&
		i < len(resp.Hourly.WindDirection120m) &&
		i < len(resp.Hourly.WindDirection180m) &&
		i < len(resp.Hourly.WindGusts10m) &&
		i < len(resp.Hourly.Temperature2m)
}

func isLowWind(speed10m, gust10m float64, cfg AppConfig) bool {
	if speed10m > cfg.LowWindMaxKnots {
		return false
	}
	return cfg.LowGustMaxKnots <= 0 || gust10m <= cfg.LowGustMaxKnots
}

func parseOpenMeteoTime(raw string, loc *time.Location) (time.Time, error) {
	if loc == nil {
		loc = time.UTC
	}
	if t, err := time.ParseInLocation("2006-01-02T15:04", raw, loc); err == nil {
		return t, nil
	}
	return time.Parse(time.RFC3339, raw)
}
