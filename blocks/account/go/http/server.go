package accounthttp

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/url"

	"github.com/emersonary/appkit/accounts"
	"github.com/emersonary/appkit/accounts/oauth"
)

type Server struct {
	svc      *accounts.Service
	provider oauth.Provider
}

func New(svc *accounts.Service) *Server {
	var provider oauth.Provider
	if p, ok := svc.OAccountProvider(oauth.ProviderGoogle); ok {
		provider = p
	}
	return &Server{svc: svc, provider: provider}
}

func (s *Server) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /account/verify-email", s.handleVerifyEmail)
	mux.HandleFunc("GET /account/google", s.handleGoogleLogin)
	mux.HandleFunc("GET /account/google/callback", s.handleGoogleCallback)
}

func (s *Server) handleVerifyEmail(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	result, err := s.svc.VerifyEmail(r.Context(), token)
	cfg := s.svc.Config()

	if err != nil {
		redirect := fmt.Sprintf("%s/verify-email?verified=invalid", cfg.URLs.FrontendURL)
		http.Redirect(w, r, redirect, http.StatusFound)
		return
	}

	status := "verified"
	if result.AlreadyVerified {
		status = "already"
	}
	redirect := fmt.Sprintf(
		"%s/verify-email?verified=%s&email=%s",
		cfg.URLs.FrontendURL,
		status,
		url.QueryEscape(result.Email),
	)
	http.Redirect(w, r, redirect, http.StatusFound)
}

func (s *Server) handleGoogleLogin(w http.ResponseWriter, r *http.Request) {
	if !s.svc.Config().RegistrationEnabled() {
		http.Error(w, "account registration is disabled", http.StatusForbidden)
		return
	}
	if s.provider == nil {
		http.Error(w, "google oauth not configured", http.StatusServiceUnavailable)
		return
	}

	state, err := randomState()
	if err != nil {
		http.Error(w, "failed to create oauth state", http.StatusInternalServerError)
		return
	}

	cookieName := s.svc.Config().OAuth.StateCookieName
	http.SetCookie(w, &http.Cookie{
		Name:     cookieName,
		Value:    state,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   600,
	})

	http.Redirect(w, r, s.provider.AuthCodeURL(state), http.StatusFound)
}

func (s *Server) handleGoogleCallback(w http.ResponseWriter, r *http.Request) {
	if !s.svc.Config().RegistrationEnabled() {
		s.redirectWithError(w, r, "registration_disabled")
		return
	}
	if s.provider == nil {
		http.Error(w, "google oauth not configured", http.StatusServiceUnavailable)
		return
	}

	cfg := s.svc.Config()
	if errMsg := r.URL.Query().Get("error"); errMsg != "" {
		s.redirectWithError(w, r, errMsg)
		return
	}

	state := r.URL.Query().Get("state")
	cookie, err := r.Cookie(cfg.OAuth.StateCookieName)
	if err != nil || cookie.Value == "" || cookie.Value != state {
		http.Error(w, "invalid oauth state", http.StatusBadRequest)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     cfg.OAuth.StateCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
	})

	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "missing code", http.StatusBadRequest)
		return
	}

	profile, err := s.provider.ExchangeUser(r.Context(), code)
	if err != nil {
		s.redirectWithError(w, r, "google_login_failed")
		return
	}

	var firstName *string
	if profile.FirstName != "" {
		firstName = &profile.FirstName
	}
	var lastName *string
	if profile.LastName != "" {
		lastName = &profile.LastName
	}
	var avatarURL *string
	if profile.AvatarURL != "" {
		avatarURL = &profile.AvatarURL
	}

	session, err := s.svc.LoginOAuth(
		r.Context(),
		s.provider.Name(),
		profile.ProviderUserID,
		profile.Email,
		firstName,
		lastName,
		avatarURL,
	)
	if err != nil {
		s.redirectWithError(w, r, "login_failed")
		return
	}

	redirectURL := fmt.Sprintf(
		"%s/account/callback?access_token=%s",
		cfg.URLs.FrontendURL,
		url.QueryEscape(session.AccessToken),
	)
	http.Redirect(w, r, redirectURL, http.StatusFound)
}

func (s *Server) redirectWithError(w http.ResponseWriter, r *http.Request, code string) {
	redirectURL := fmt.Sprintf(
		"%s/account/callback?error=%s",
		s.svc.Config().URLs.FrontendURL,
		url.QueryEscape(code),
	)
	http.Redirect(w, r, redirectURL, http.StatusFound)
}

func randomState() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
