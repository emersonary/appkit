package oauth

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const stateTTL = 15 * time.Minute

type oauthState struct {
	TenantID   string `json:"tenant_id"`
	PlatformID string `json:"platform_id"`
	Language   string `json:"language"`
	Nonce      string `json:"nonce"`
	ExpiresAt  int64  `json:"exp"`
}

// CallbackLogger receives OAuth callback failures for host app logging.
type CallbackLogger interface {
	Warn(msg string, keysAndValues ...any)
}

// Manager coordinates publishing OAuth across platforms for multiple tenants.
type Manager struct {
	store       ConnectionStore
	tenants     TenantOAuthResolver
	providers   map[string]Provider
	apiPublic   string
	frontend    string
	stateSecret []byte
	httpClient  *http.Client
	log         CallbackLogger
}

type ManagerOptions struct {
	Store         ConnectionStore
	Tenants       TenantOAuthResolver
	Providers     []Provider
	APIPublicURL  string
	FrontendURL   string
	StateSecret   string
	Logger        CallbackLogger
}

func NewManager(opts ManagerOptions) (*Manager, error) {
	if opts.Store == nil {
		return nil, errors.New("social oauth: connection store required")
	}
	if opts.Tenants == nil {
		return nil, errors.New("social oauth: tenant resolver required")
	}
	if strings.TrimSpace(opts.StateSecret) == "" {
		return nil, errors.New("social oauth: state secret required")
	}

	providers := make(map[string]Provider, len(opts.Providers))
	for _, p := range opts.Providers {
		if p == nil {
			continue
		}
		id := strings.TrimSpace(p.PlatformID())
		if id == "" {
			continue
		}
		providers[id] = p
	}

	return &Manager{
		store:       opts.Store,
		tenants:     opts.Tenants,
		providers:   providers,
		apiPublic:   trimRightSlash(opts.APIPublicURL),
		frontend:    trimRightSlash(opts.FrontendURL),
		stateSecret: []byte(opts.StateSecret),
		httpClient:  &http.Client{Timeout: 30 * time.Second},
		log:         opts.Logger,
	}, nil
}

func (m *Manager) OAuthConfigured(tenantID, platformID string) bool {
	_, ok := m.tenants.OAuthConfig(tenantID, platformID)
	return ok
}

func (m *Manager) Status(ctx context.Context, tenantID, platformID, language string) (ConnectionStatus, error) {
	oauthConfigured := m.OAuthConfigured(tenantID, platformID)
	conn, valid, err := m.store.ValidAt(ctx, tenantID, platformID, language, time.Now())
	if err != nil {
		return ConnectionStatus{}, err
	}
	if valid {
		st := connectionStatus(conn, true)
		st.OAuthConfigured = oauthConfigured
		return st, nil
	}
	conn, err = m.store.Get(ctx, tenantID, platformID, language)
	if err != nil {
		return ConnectionStatus{Connected: false, OAuthConfigured: oauthConfigured}, nil
	}
	st := connectionStatus(conn, false)
	st.Expired = true
	st.DaysLeft = 0
	st.OAuthConfigured = oauthConfigured
	return st, nil
}

func (m *Manager) Disconnect(ctx context.Context, tenantID, platformID, language string) error {
	return m.store.Delete(ctx, tenantID, platformID, language)
}

func (m *Manager) AuthorizeURL(ctx context.Context, tenantID, platformID, language string) (string, error) {
	provider, oauthCfg, redirectURI, err := m.resolveOAuth(tenantID, platformID)
	if err != nil {
		return "", err
	}
	language = strings.TrimSpace(language)
	if language == "" {
		return "", errors.New("language required for social oauth")
	}

	nonce, err := randomNonce()
	if err != nil {
		return "", err
	}
	state, err := m.signState(oauthState{
		TenantID:   tenantID,
		PlatformID: platformID,
		Language:   language,
		Nonce:      nonce,
		ExpiresAt:  time.Now().Add(stateTTL).Unix(),
	})
	if err != nil {
		return "", err
	}
	return provider.AuthorizeURL(oauthCfg, redirectURI, state)
}

func (m *Manager) StateCookieName(platformID, language string) string {
	return "social_oauth_state_" + strings.TrimSpace(platformID) + "_" + strings.TrimSpace(language)
}

func (m *Manager) SetStateCookie(w http.ResponseWriter, platformID, language, authorizeURL string) error {
	u, err := url.Parse(authorizeURL)
	if err != nil {
		return err
	}
	state := u.Query().Get("state")
	if state == "" {
		return errors.New("missing state in authorize url")
	}
	name := m.StateCookieName(platformID, language)
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    state,
		Path:     "/",
		MaxAge:   int(stateTTL.Seconds()),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	return nil
}

func (m *Manager) HandleCallback(ctx context.Context, platformID string, w http.ResponseWriter, r *http.Request) {
	platformID = strings.TrimSpace(platformID)

	if errCode := strings.TrimSpace(r.URL.Query().Get("error")); errCode != "" {
		detail := strings.TrimSpace(r.URL.Query().Get("error_description"))
		if detail == "" {
			detail = errCode
		} else {
			detail = errCode + ": " + detail
		}
		m.failCallback(w, r, platformID, "", "provider_error", detail)
		return
	}

	code := strings.TrimSpace(r.URL.Query().Get("code"))
	stateParam := strings.TrimSpace(r.URL.Query().Get("state"))
	if code == "" {
		m.failCallback(w, r, platformID, "", "missing_code", "callback missing code and error parameters")
		return
	}

	st, err := m.parseState(stateParam)
	if err != nil || st.PlatformID != platformID {
		m.failCallback(w, r, platformID, "", "invalid_state", "oauth state invalid or expired")
		return
	}

	cookieName := m.StateCookieName(platformID, st.Language)
	cookie, cookieErr := r.Cookie(cookieName)
	if cookieErr == nil && cookie.Value != "" {
		if cookie.Value != stateParam {
			m.failCallback(w, r, platformID, st.Language, "invalid_state", "oauth state cookie mismatch")
			return
		}
		http.SetCookie(w, &http.Cookie{Name: cookieName, Value: "", Path: "/", MaxAge: -1, HttpOnly: true, SameSite: http.SameSiteLaxMode})
	}
	// When the redirect URI is on another host (e.g. emersonary.dev) the state cookie set at
	// /start is not sent here; the HMAC-signed state parameter is sufficient.

	provider, oauthCfg, redirectURI, err := m.resolveOAuth(st.TenantID, platformID)
	if err != nil {
		m.failCallback(w, r, platformID, st.Language, "oauth_not_configured", err.Error())
		return
	}

	tokenResp, err := provider.ExchangeCode(ctx, oauthCfg, redirectURI, code)
	if err != nil {
		m.failCallback(w, r, platformID, st.Language, "token_exchange_failed", err.Error())
		return
	}

	accountID, err := provider.ResolveAccountID(ctx, tokenResp.AccessToken)
	if err != nil {
		m.failCallback(w, r, platformID, st.Language, "profile_failed", err.Error())
		return
	}

	expiresIn := tokenResp.ExpiresIn
	if expiresIn <= 0 {
		expiresIn = 5184000
	}
	expiresAt := time.Now().Add(time.Duration(expiresIn) * time.Second)
	scopes := tokenResp.Scopes
	if scopes == "" {
		scopes = oauthCfg.ResolvedScope(provider.DefaultScope())
	}

	if err := m.store.Upsert(ctx, Connection{
		TenantID:    st.TenantID,
		PlatformID:  platformID,
		Language:    st.Language,
		AccountID:   accountID,
		AccessToken: tokenResp.AccessToken,
		ExpiresAt:   expiresAt,
		Scopes:      scopes,
	}); err != nil {
		m.failCallback(w, r, platformID, st.Language, "store_failed", err.Error())
		return
	}

	target := m.frontend + "/admin?social=" + platformID + "_connected&lang=" + url.QueryEscape(st.Language)
	http.Redirect(w, r, target, http.StatusFound)
}

func (m *Manager) ValidConnection(ctx context.Context, tenantID, platformID, language string) (Connection, error) {
	conn, valid, err := m.store.ValidAt(ctx, tenantID, platformID, language, time.Now())
	if err != nil {
		return Connection{}, err
	}
	if valid {
		return conn, nil
	}
	if _, err := m.store.Get(ctx, tenantID, platformID, language); err == nil {
		return Connection{}, ErrReconnectRequired
	}
	return Connection{}, ErrReconnectRequired
}

func (m *Manager) resolveOAuth(tenantID, platformID string) (Provider, PlatformOAuthConfig, string, error) {
	provider, ok := m.providers[platformID]
	if !ok {
		return nil, PlatformOAuthConfig{}, "", fmt.Errorf("oauth provider %q not registered", platformID)
	}
	oauthCfg, ok := m.tenants.OAuthConfig(tenantID, platformID)
	if !ok || !oauthCfg.Configured() {
		return nil, PlatformOAuthConfig{}, "", fmt.Errorf("oauth not configured for tenant %q platform %q", tenantID, platformID)
	}
	redirectURI := strings.TrimSpace(oauthCfg.RedirectURI)
	if redirectURI == "" {
		redirectURI = m.apiPublic + "/auth/social/" + platformID + "/callback"
	}
	return provider, oauthCfg, redirectURI, nil
}

func (m *Manager) redirectError(w http.ResponseWriter, r *http.Request, platformID, code string) {
	m.failCallback(w, r, platformID, "", code, "")
}

func (m *Manager) failCallback(w http.ResponseWriter, r *http.Request, platformID, language, reason, detail string) {
	detail = sanitizeCallbackDetail(detail)
	m.logCallbackError(platformID, reason, detail, r.URL.RawQuery)

	target := m.frontend + "/admin?social=" + url.QueryEscape(platformID) + "_error&reason=" + url.QueryEscape(reason)
	if language != "" {
		target += "&lang=" + url.QueryEscape(language)
	}
	if detail != "" {
		target += "&detail=" + url.QueryEscape(detail)
	}
	http.Redirect(w, r, target, http.StatusFound)
}

func (m *Manager) logCallbackError(platformID, reason, detail, rawQuery string) {
	msg := fmt.Sprintf("social oauth callback failed platform=%s reason=%s", platformID, reason)
	if m.log != nil {
		m.log.Warn(msg, "detail", detail, "query", rawQuery)
		return
	}
	log.Printf("%s detail=%q query=%q", msg, detail, rawQuery)
}

func sanitizeCallbackDetail(detail string) string {
	detail = strings.Join(strings.Fields(detail), " ")
	if len(detail) > 400 {
		return detail[:400] + "…"
	}
	return detail
}

// RedirectStartError sends the browser back to the admin UI when /start fails.
func (m *Manager) RedirectStartError(w http.ResponseWriter, r *http.Request, platformID, language string, err error) {
	reason := "oauth_start_failed"
	detail := ""
	if err != nil {
		detail = err.Error()
		if strings.Contains(detail, "oauth not configured") {
			reason = "oauth_not_configured"
		}
	}
	m.failCallback(w, r, platformID, language, reason, detail)
}

func connectionStatus(conn Connection, connected bool) ConnectionStatus {
	exp := conn.ExpiresAt.UTC()
	days := int(time.Until(conn.ExpiresAt).Hours() / 24)
	if days < 0 {
		days = 0
	}
	return ConnectionStatus{
		Connected: connected,
		Expired:   !connected,
		AccountID: conn.AccountID,
		ExpiresAt: &exp,
		DaysLeft:  days,
	}
}

func (m *Manager) signState(st oauthState) (string, error) {
	payload, err := json.Marshal(st)
	if err != nil {
		return "", err
	}
	encoded := base64.RawURLEncoding.EncodeToString(payload)
	sig := sign(encoded, m.stateSecret)
	return encoded + "." + sig, nil
}

func (m *Manager) parseState(raw string) (oauthState, error) {
	parts := strings.Split(raw, ".")
	if len(parts) != 2 {
		return oauthState{}, errors.New("invalid state format")
	}
	if !hmac.Equal([]byte(sign(parts[0], m.stateSecret)), []byte(parts[1])) {
		return oauthState{}, errors.New("invalid state signature")
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return oauthState{}, err
	}
	var st oauthState
	if err := json.Unmarshal(payload, &st); err != nil {
		return oauthState{}, err
	}
	if time.Now().Unix() > st.ExpiresAt {
		return oauthState{}, errors.New("state expired")
	}
	if strings.TrimSpace(st.TenantID) == "" || strings.TrimSpace(st.PlatformID) == "" || strings.TrimSpace(st.Language) == "" {
		return oauthState{}, errors.New("missing tenant, platform, or language")
	}
	return st, nil
}

func sign(payload string, secret []byte) string {
	mac := hmac.New(sha256.New, secret)
	_, _ = mac.Write([]byte(payload))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func randomNonce() (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func trimRightSlash(v string) string {
	return strings.TrimRight(strings.TrimSpace(v), "/")
}

// ReadBody is a test helper for providers.
func ReadBody(r io.Reader, limit int64) ([]byte, error) {
	return io.ReadAll(io.LimitReader(r, limit))
}
