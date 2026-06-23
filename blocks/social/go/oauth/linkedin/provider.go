package linkedin

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/emersonary/appkit/social/oauth"
)

const (
	PlatformID     = "li"
	defaultScope   = "openid profile w_member_social"
	authURL        = "https://www.linkedin.com/oauth/v2/authorization"
	tokenURL       = "https://www.linkedin.com/oauth/v2/accessToken"
	userInfoURL    = "https://api.linkedin.com/v2/userinfo"
	maxBodyBytes   = 1 << 20
)

type Provider struct {
	httpClient *http.Client
}

func NewProvider() *Provider {
	return &Provider{httpClient: &http.Client{Timeout: 30 * time.Second}}
}

func NewProviderWithClient(client *http.Client) *Provider {
	if client == nil {
		client = &http.Client{Timeout: 30 * time.Second}
	}
	return &Provider{httpClient: client}
}

func (p *Provider) PlatformID() string { return PlatformID }

func (p *Provider) DefaultScope() string { return defaultScope }

func (p *Provider) AuthorizeURL(cfg oauth.PlatformOAuthConfig, redirectURI, state string) (string, error) {
	q := url.Values{}
	q.Set("response_type", "code")
	q.Set("client_id", cfg.ResolvedClientID())
	q.Set("redirect_uri", redirectURI)
	q.Set("scope", cfg.ResolvedScope(defaultScope))
	q.Set("state", state)
	return authURL + "?" + q.Encode(), nil
}

func (p *Provider) ExchangeCode(ctx context.Context, cfg oauth.PlatformOAuthConfig, redirectURI, code string) (oauth.TokenResult, error) {
	form := url.Values{}
	form.Set("grant_type", "authorization_code")
	form.Set("code", code)
	form.Set("redirect_uri", redirectURI)
	form.Set("client_id", cfg.ResolvedClientID())
	form.Set("client_secret", cfg.ResolvedClientSecret())

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return oauth.TokenResult{}, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return oauth.TokenResult{}, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(io.LimitReader(resp.Body, maxBodyBytes))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return oauth.TokenResult{}, fmt.Errorf("token exchange %d: %s", resp.StatusCode, string(body))
	}

	var out struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int64  `json:"expires_in"`
	}
	if err := json.Unmarshal(body, &out); err != nil {
		return oauth.TokenResult{}, err
	}
	if strings.TrimSpace(out.AccessToken) == "" {
		return oauth.TokenResult{}, fmt.Errorf("empty access_token")
	}
	return oauth.TokenResult{
		AccessToken: out.AccessToken,
		ExpiresIn:   out.ExpiresIn,
		Scopes:      cfg.ResolvedScope(defaultScope),
	}, nil
}

func (p *Provider) ResolveAccountID(ctx context.Context, accessToken string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, userInfoURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(io.LimitReader(resp.Body, maxBodyBytes))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("userinfo %d: %s", resp.StatusCode, string(body))
	}

	var info struct {
		Sub string `json:"sub"`
	}
	if err := json.Unmarshal(body, &info); err != nil {
		return "", err
	}
	sub := strings.TrimSpace(info.Sub)
	if sub == "" {
		return "", fmt.Errorf("empty linkedin sub")
	}
	if strings.HasPrefix(sub, "urn:li:") {
		return sub, nil
	}
	return "urn:li:person:" + sub, nil
}
