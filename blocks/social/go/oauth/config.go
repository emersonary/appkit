package oauth

import (
	"os"
	"strings"
)

// PlatformOAuthConfig holds per-tenant OAuth app credentials for one platform.
// Secrets are resolved from ClientSecret or ClientSecretEnv (tenant-specific env var).
type PlatformOAuthConfig struct {
	ClientID        string
	ClientIDEnv     string
	ClientSecret    string
	ClientSecretEnv string
	Scope           string
	RedirectURI     string
}

func (c PlatformOAuthConfig) ResolvedClientID() string {
	if s := strings.TrimSpace(c.ClientID); s != "" {
		return s
	}
	if env := strings.TrimSpace(c.ClientIDEnv); env != "" {
		return strings.TrimSpace(os.Getenv(env))
	}
	return ""
}

func (c PlatformOAuthConfig) ResolvedScope(defaultScope string) string {
	if s := strings.TrimSpace(c.Scope); s != "" {
		return s
	}
	return defaultScope
}

func (c PlatformOAuthConfig) ResolvedClientSecret() string {
	if s := strings.TrimSpace(c.ClientSecret); s != "" {
		return s
	}
	if env := strings.TrimSpace(c.ClientSecretEnv); env != "" {
		return strings.TrimSpace(os.Getenv(env))
	}
	return ""
}

func (c PlatformOAuthConfig) Configured() bool {
	return c.ResolvedClientID() != "" && c.ResolvedClientSecret() != ""
}
