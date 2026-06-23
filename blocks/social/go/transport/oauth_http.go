package transport

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/emersonary/appkit/social/oauth"
)

// AdminScopeResolver extracts the tenant (project) id from admin requests.
type AdminScopeResolver interface {
	TenantFromBearer(r *http.Request) (tenantID string, ok bool)
	TenantFromTokenQuery(r *http.Request) (tenantID string, ok bool)
}

// OAuthHandler serves generic publishing OAuth routes per platform and language.
type OAuthHandler struct {
	Manager *oauth.Manager
	Scope   AdminScopeResolver
}

func (h OAuthHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /auth/social/{platform}/{language}/authorize", h.authorize)
	mux.HandleFunc("GET /auth/social/{platform}/{language}/start", h.start)
	mux.HandleFunc("GET /auth/social/{platform}/callback", h.callback)
	mux.HandleFunc("GET /auth/social/{platform}/{language}/status", h.status)
	mux.HandleFunc("DELETE /auth/social/{platform}/{language}", h.disconnect)
}

func (h OAuthHandler) authorize(w http.ResponseWriter, r *http.Request) {
	platformID, language := routePlatformLanguage(r)
	tenantID, ok := h.Scope.TenantFromBearer(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	if h.Manager == nil {
		http.Error(w, "social oauth unavailable", http.StatusServiceUnavailable)
		return
	}

	url, err := h.Manager.AuthorizeURL(r.Context(), tenantID, platformID, language)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := h.Manager.SetStateCookie(w, platformID, language, url); err != nil {
		http.Error(w, "oauth state error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"url": url})
}

func (h OAuthHandler) start(w http.ResponseWriter, r *http.Request) {
	platformID, language := routePlatformLanguage(r)
	tenantID, ok := h.Scope.TenantFromTokenQuery(r)
	if !ok {
		http.Error(w, "unauthorized: invalid or expired admin token", http.StatusUnauthorized)
		return
	}
	if h.Manager == nil {
		http.Error(w, "social oauth unavailable", http.StatusServiceUnavailable)
		return
	}

	url, err := h.Manager.AuthorizeURL(r.Context(), tenantID, platformID, language)
	if err != nil {
		h.Manager.RedirectStartError(w, r, platformID, language, err)
		return
	}
	if err := h.Manager.SetStateCookie(w, platformID, language, url); err != nil {
		http.Error(w, "oauth state error", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, url, http.StatusFound)
}

func (h OAuthHandler) callback(w http.ResponseWriter, r *http.Request) {
	platformID := strings.TrimSpace(r.PathValue("platform"))
	if h.Manager == nil {
		http.Error(w, "social oauth unavailable", http.StatusServiceUnavailable)
		return
	}
	h.Manager.HandleCallback(r.Context(), platformID, w, r)
}

func (h OAuthHandler) status(w http.ResponseWriter, r *http.Request) {
	platformID, language := routePlatformLanguage(r)
	tenantID, ok := h.Scope.TenantFromBearer(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	if h.Manager == nil {
		http.Error(w, "social oauth unavailable", http.StatusServiceUnavailable)
		return
	}
	st, err := h.Manager.Status(r.Context(), tenantID, platformID, language)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(st)
}

func (h OAuthHandler) disconnect(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	platformID, language := routePlatformLanguage(r)
	tenantID, ok := h.Scope.TenantFromBearer(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	if h.Manager == nil {
		http.Error(w, "social oauth unavailable", http.StatusServiceUnavailable)
		return
	}
	if err := h.Manager.Disconnect(r.Context(), tenantID, platformID, language); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func routePlatformLanguage(r *http.Request) (platformID, language string) {
	return strings.TrimSpace(r.PathValue("platform")), strings.TrimSpace(r.PathValue("language"))
}

// Block wires social transport on the host HTTP mux.
type Block struct {
	OAuth OAuthHandler
}

type Mount struct {
	HTTPMux *http.ServeMux
}

func New(oauthHandler OAuthHandler, mount *Mount) *Block {
	b := &Block{OAuth: oauthHandler}
	if mount != nil && mount.HTTPMux != nil {
		b.OAuth.Register(mount.HTTPMux)
	}
	return b
}
