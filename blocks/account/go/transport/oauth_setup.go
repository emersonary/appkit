package transport

import (
	"github.com/emersonary/appkit/accounts"
	"github.com/emersonary/appkit/accounts/oauth"
	googleoauth "github.com/emersonary/appkit/accounts/oauth/google"
)

// OAuthProviders reports which OAuth providers are active after configuration.
type OAuthProviders struct {
	Google   bool
	Facebook bool
	Apple    bool
}

func configureOAuthProviders(svc *accounts.Service) OAuthProviders {
	var out OAuthProviders
	if !svc.Config().RegistrationEnabled() || !svc.Config().OAuthEnabled() {
		return out
	}

	if svc.Config().OAuthProviderEnabled(oauth.ProviderGoogle) {
		out.Google = googleoauth.Configure(svc)
	}

	// Reserved for future providers.
	if svc.Config().OAuthProviderEnabled(oauth.ProviderFacebook) {
		out.Facebook = false
	}
	if svc.Config().OAuthProviderEnabled(oauth.ProviderApple) {
		out.Apple = false
	}

	return out
}
