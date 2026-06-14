package currency

import "time"

const BaseCurrencyCode = "USD"

type Currency struct {
	Code             string
	Name             string
	Symbol           string
	IsBase           bool
	WebsiteLanguages []string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type ExchangeRate struct {
	CurrencyCode string
	Rate         float64
	Source       string
	FetchedAt    time.Time
}

type ExchangeRateHistory struct {
	ID           string
	CurrencyCode string
	Rate         float64
	Source       string
	RecordedAt   time.Time
}

type Snapshot struct {
	BaseCurrency string
	Rates        map[string]float64
	Source       string
	FetchedAt    time.Time
}

type SyncResult struct {
	Updated int
	Skipped int
	Source  string
}
