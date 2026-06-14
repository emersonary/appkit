package currency

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const DefaultAPIURL = "https://open.er-api.com/v6/latest/USD"

type Client struct {
	apiURL string
	http   *http.Client
}

func NewClient(apiURL string) *Client {
	url := strings.TrimSpace(apiURL)
	if url == "" {
		url = DefaultAPIURL
	}

	return &Client{
		apiURL: url,
		http: &http.Client{
			Timeout: 20 * time.Second,
		},
	}
}

func (c *Client) FetchUSDRates(ctx context.Context) (Snapshot, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.apiURL, nil)
	if err != nil {
		return Snapshot{}, err
	}

	res, err := c.http.Do(req)
	if err != nil {
		return Snapshot{}, wrapErr(ErrFeedFetch, "request", err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(io.LimitReader(res.Body, 1<<20))
	if err != nil {
		return Snapshot{}, wrapErr(ErrFeedFetch, "read", err)
	}
	if res.StatusCode != http.StatusOK {
		return Snapshot{}, ErrFeedFetch.With("status", fmt.Sprintf("%d: %s", res.StatusCode, strings.TrimSpace(string(body))))
	}

	rates, base, err := parseRates(body)
	if err != nil {
		return Snapshot{}, wrapErr(ErrFeedFetch, "parse", err)
	}
	if base == "" {
		base = BaseCurrencyCode
	}
	if base != BaseCurrencyCode {
		return Snapshot{}, ErrFeedFetch.With("base", base)
	}
	if len(rates) == 0 {
		return Snapshot{}, ErrFeedFetch.With("rates", "empty")
	}

	return Snapshot{
		BaseCurrency: base,
		Rates:        rates,
		Source:       c.apiURL,
		FetchedAt:    time.Now().UTC(),
	}, nil
}

func parseRates(body []byte) (map[string]float64, string, error) {
	var openER struct {
		Result   string             `json:"result"`
		BaseCode string             `json:"base_code"`
		Rates    map[string]float64 `json:"rates"`
		Error    string             `json:"error-type"`
	}
	if err := json.Unmarshal(body, &openER); err == nil && len(openER.Rates) > 0 &&
		(openER.BaseCode != "" || openER.Result == "success") {
		if openER.Result != "" && openER.Result != "success" {
			return nil, "", fmt.Errorf("currency feed error: %s", openER.Error)
		}

		return openER.Rates, openER.BaseCode, nil
	}

	var frankfurter struct {
		Base  string             `json:"base"`
		Rates map[string]float64 `json:"rates"`
	}
	if err := json.Unmarshal(body, &frankfurter); err != nil {
		return nil, "", fmt.Errorf("currency feed parse: %w", err)
	}
	if len(frankfurter.Rates) == 0 {
		return nil, "", fmt.Errorf("currency feed parse: empty rates")
	}

	return frankfurter.Rates, frankfurter.Base, nil
}
