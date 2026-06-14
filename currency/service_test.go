package currency

import "testing"

func TestConvertViaUSD(t *testing.T) {
	svc, err := NewService(nil, Config{
		Schema:       "public",
		BaseCurrency: BaseCurrencyCode,
		Currencies:   []string{BaseCurrencyCode, "EUR", "BRL"},
	}, Options{})
	if err != nil {
		t.Fatal(err)
	}

	got, err := svc.convertWithRates(100, "EUR", "BRL", map[string]float64{
		"EUR": 0.5,
		"BRL": 5.0,
	})
	if err != nil {
		t.Fatal(err)
	}

	want := 1000.0
	if got != want {
		t.Fatalf("got %v want %v", got, want)
	}
}

func TestConvertSameCurrency(t *testing.T) {
	svc, err := NewService(nil, Config{
		Schema:       "public",
		BaseCurrency: BaseCurrencyCode,
		Currencies:   []string{BaseCurrencyCode, "EUR"},
	}, Options{})
	if err != nil {
		t.Fatal(err)
	}

	_, err = svc.convertWithRates(10, "USD", "USD", nil)
	if err != ErrSameCurrency {
		t.Fatalf("expected ErrSameCurrency, got %v", err)
	}
}

func (s *Service) convertWithRates(amount float64, fromCode, toCode string, rates map[string]float64) (float64, error) {
	if amount <= 0 {
		return 0, ErrInvalidAmount
	}

	fromCode = NormalizeCode(fromCode)
	toCode = NormalizeCode(toCode)

	if err := s.ValidateCode(fromCode); err != nil {
		return 0, err
	}
	if err := s.ValidateCode(toCode); err != nil {
		return 0, err
	}

	if fromCode == toCode {
		return 0, ErrSameCurrency
	}

	usdAmount := amount
	if fromCode != s.cfg.BaseCurrency {
		rate, ok := rates[fromCode]
		if !ok || rate <= 0 {
			return 0, ErrRateNotFound.With("code", fromCode)
		}

		usdAmount = amount / rate
	}

	if toCode == s.cfg.BaseCurrency {
		return usdAmount, nil
	}

	rate, ok := rates[toCode]
	if !ok || rate <= 0 {
		return 0, ErrRateNotFound.With("code", toCode)
	}

	return usdAmount * rate, nil
}
