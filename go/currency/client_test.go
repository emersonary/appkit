package currency

import "testing"

func TestParseRatesOpenER(t *testing.T) {
	body := []byte(`{"result":"success","base_code":"USD","rates":{"EUR":0.92,"BRL":5.1}}`)
	rates, base, err := parseRates(body)
	if err != nil {
		t.Fatal(err)
	}
	if base != "USD" {
		t.Fatalf("base=%s", base)
	}
	if rates["EUR"] != 0.92 || rates["BRL"] != 5.1 {
		t.Fatalf("rates=%v", rates)
	}
}

func TestParseRatesFrankfurter(t *testing.T) {
	body := []byte(`{"amount":1,"base":"USD","date":"2026-06-05","rates":{"JPY":157.2}}`)
	rates, base, err := parseRates(body)
	if err != nil {
		t.Fatal(err)
	}
	if base != "USD" {
		t.Fatalf("base=%s", base)
	}
	if rates["JPY"] != 157.2 {
		t.Fatalf("rates=%v", rates)
	}
}
