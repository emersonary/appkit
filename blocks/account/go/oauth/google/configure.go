package google

import (
	"strings"

	"github.com/emersonary/appkit/accounts"
)

// Configure registers the Google OAuth provider on svc when config and secrets allow it.
// Returns true when Google OAuth is enabled and registered.
func Configure(svc *accounts.Service) bool {
	cfg := svc.Config()
	secrets := svc.Secrets()
	google := cfg.OAuth.Google
	if !google.EnabledWithSecrets(secrets) {
		return false
	}

	svc.RegisterProvider(New(Config{
		ClientID:     firstNonEmptyCredential(secrets.GoogleClientID, google.ClientID),
		ClientSecret: firstNonEmptyCredential(secrets.GoogleClientSecret, google.ClientSecret),
		RedirectURL:  google.RedirectURL,
	}))
	return true
}

func firstNonEmptyCredential(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}
