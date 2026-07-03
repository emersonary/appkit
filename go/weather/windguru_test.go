package weather

import "testing"

func TestEstimateWindGuruUses80mAnd120mWeights(t *testing.T) {
	got := EstimateWindGuru(5, 10, 20, 30)
	want := 13.5
	if got != want {
		t.Fatalf("EstimateWindGuru() = %f, want %f", got, want)
	}
}
