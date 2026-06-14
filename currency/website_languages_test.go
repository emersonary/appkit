package currency

import "testing"

func TestCurrencyWebsiteLanguagesCoversISO4217Catalog(t *testing.T) {
	for code := range ISO4217Catalog {
		if _, ok := CurrencyWebsiteLanguages[code]; !ok {
			t.Fatalf("expected website languages for ISO4217 code %s", code)
		}
	}

	if len(CurrencyWebsiteLanguages) != len(ISO4217Catalog) {
		t.Fatalf("got %d website language entries, want %d", len(CurrencyWebsiteLanguages), len(ISO4217Catalog))
	}
}

func TestWebsiteLanguagesForCurrency(t *testing.T) {
	if got := WebsiteLanguagesForCurrency("BRL"); len(got) != 1 || got[0] != "pt" {
		t.Fatalf("BRL languages = %v, want [pt]", got)
	}

	if got := WebsiteLanguagesForCurrency("USD"); len(got) == 0 {
		t.Fatalf("USD languages = %v, want non-empty", got)
	}

	if got := WebsiteLanguagesForCurrency(" eur "); len(got) < 4 {
		t.Fatalf("EUR languages = %v, want at least 4 entries", got)
	}

	if got := WebsiteLanguagesForCurrency("XTS"); len(got) != 0 {
		t.Fatalf("XTS languages = %v, want empty slice", got)
	}
}
