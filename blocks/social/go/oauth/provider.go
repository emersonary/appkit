package oauth

import "context"

// TokenResult is returned after exchanging an authorization code.
type TokenResult struct {
	AccessToken string
	ExpiresIn   int64
	Scopes      string
}

// Provider implements OAuth for one publishing platform (LinkedIn, Meta, …).
type Provider interface {
	PlatformID() string
	DefaultScope() string
	AuthorizeURL(cfg PlatformOAuthConfig, redirectURI, state string) (string, error)
	ExchangeCode(ctx context.Context, cfg PlatformOAuthConfig, redirectURI, code string) (TokenResult, error)
	ResolveAccountID(ctx context.Context, accessToken string) (string, error)
}
