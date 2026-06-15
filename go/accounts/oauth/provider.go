package oauth

import "context"

const ProviderGoogle = "google"
const ProviderFacebook = "facebook"
const ProviderApple = "apple"

type Profile struct {
	ProviderUserID string
	Email          string
	FirstName      string
	LastName       string
	AvatarURL      string
	EmailVerified  bool
}

type Provider interface {
	Name() string
	AuthCodeURL(state string) string
	ExchangeUser(ctx context.Context, code string) (Profile, error)
}
