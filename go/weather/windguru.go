package weather

// EstimateWindGuru converts Open-Meteo tower-level samples into the wind speed
// estimate shown in the UI. The full signature keeps room for future formulas.
func EstimateWindGuru(w10, w80, w120, w180 float64) float64 {
	return 0.65*w80 + 0.35*w120
}
