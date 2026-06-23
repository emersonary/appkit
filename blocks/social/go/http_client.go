package social

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type httpAPIClient struct {
	baseURL    string
	httpClient *http.Client
	authHeader func(*http.Request)
}

func newHTTPAPIClient(baseURL string, timeout time.Duration, authHeader func(*http.Request)) *httpAPIClient {
	if timeout <= 0 {
		timeout = defaultPlatformTimeout
	}
	return &httpAPIClient{
		baseURL:    strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{Timeout: timeout},
		authHeader: authHeader,
	}
}

func (c *httpAPIClient) withHTTPClient(client *http.Client) *httpAPIClient {
	if client != nil {
		c.httpClient = client
	}
	return c
}

func (c *httpAPIClient) doJSON(ctx context.Context, method, path string, query url.Values, body any, out any) error {
	var bodyReader io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			return wrapErr(ErrAPIFailed, "marshal", err)
		}
		bodyReader = bytes.NewReader(raw)
	}

	reqURL := c.baseURL + path
	if len(query) > 0 {
		reqURL += "?" + query.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, method, reqURL, bodyReader)
	if err != nil {
		return wrapErr(ErrAPIFailed, "request", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.authHeader != nil {
		c.authHeader(req)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return wrapErr(ErrAPIFailed, "http", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return wrapErr(ErrAPIFailed, "read", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return ErrAPIFailed.With("status", fmt.Sprintf("%d: %s", resp.StatusCode, strings.TrimSpace(string(respBody))))
	}

	if out == nil || len(respBody) == 0 {
		return nil
	}
	if err := json.Unmarshal(respBody, out); err != nil {
		return wrapErr(ErrAPIFailed, "decode", err)
	}
	return nil
}

func bearerAuth(token string) func(*http.Request) {
	return func(req *http.Request) {
		req.Header.Set("Authorization", "Bearer "+token)
	}
}

func queryAccessToken(token string) url.Values {
	return url.Values{"access_token": {token}}
}
