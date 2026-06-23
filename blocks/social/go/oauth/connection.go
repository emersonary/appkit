package oauth

import "time"

// Connection stores a tenant OAuth token for publishing on one platform and language.
type Connection struct {
	ID          string
	TenantID    string
	PlatformID  string
	Language    string
	AccountID   string
	AccessToken string
	ExpiresAt   time.Time
	Scopes      string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (c Connection) ValidAt(now time.Time) bool {
	return c.AccessToken != "" && now.Before(c.ExpiresAt)
}

// ConnectionStatus is returned by the OAuth manager status endpoint.
type ConnectionStatus struct {
	OAuthConfigured bool       `json:"oauth_configured"`
	Connected       bool       `json:"connected"`
	Expired         bool       `json:"expired"`
	AccountID       string     `json:"account_id,omitempty"`
	ExpiresAt       *time.Time `json:"expires_at,omitempty"`
	DaysLeft        int        `json:"days_left"`
}

// TenantOAuthResolver returns per-tenant OAuth app credentials (client_id, tenant-scoped secret env).
type TenantOAuthResolver interface {
	OAuthConfig(tenantID, platformID string) (PlatformOAuthConfig, bool)
}
