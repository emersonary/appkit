package google

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/emersonary/appkit/accounts"
	"github.com/emersonary/appkit/accounts/oauth"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type Config struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

type Provider struct {
	oauth *oauth2.Config
}

type userInfo struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
}

func New(cfg Config) *Provider {
	return &Provider{
		oauth: &oauth2.Config{
			ClientID:     cfg.ClientID,
			ClientSecret: cfg.ClientSecret,
			RedirectURL:  cfg.RedirectURL,
			Scopes: []string{
				"https://www.googleapis.com/auth/userinfo.email",
				"https://www.googleapis.com/auth/userinfo.profile",
			},
			Endpoint: google.Endpoint,
		},
	}
}

func (p *Provider) Name() string {
	return oauth.ProviderGoogle
}

func (p *Provider) AuthCodeURL(state string) string {
	return p.oauth.AuthCodeURL(state, oauth2.AccessTypeOnline)
}

func (p *Provider) ExchangeUser(ctx context.Context, code string) (oauth.Profile, error) {
	token, err := p.oauth.Exchange(ctx, code)
	if err != nil {
		return oauth.Profile{}, fmt.Errorf("exchange code: %w", err)
	}

	client := p.oauth.Client(ctx, token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return oauth.Profile{}, fmt.Errorf("fetch userinfo: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return oauth.Profile{}, fmt.Errorf("userinfo status %d: %s", resp.StatusCode, string(body))
	}

	var info userInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return oauth.Profile{}, fmt.Errorf("decode userinfo: %w", err)
	}
	if info.Email == "" || !info.VerifiedEmail {
		return oauth.Profile{}, fmt.Errorf("google email not verified")
	}

	firstName := info.GivenName
	lastName := info.FamilyName
	if firstName == "" && info.Name != "" {
		firstName, lastName = accounts.SplitFullName(info.Name)
	}

	return oauth.Profile{
		ProviderUserID: info.ID,
		Email:          info.Email,
		FirstName:      firstName,
		LastName:       lastName,
		AvatarURL:      info.Picture,
		EmailVerified:  true,
	}, nil
}
