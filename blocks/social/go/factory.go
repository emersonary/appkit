package social

import (
	"fmt"
	"net/http"
)

func buildPlatformClient(platformID PlatformID, cfg PlatformConfig, templates *TemplateRenderer, defaultDispatch DispatchMode, httpClient *http.Client) (PlatformClient, error) {
	dispatch, err := cfg.resolvedDispatch(defaultDispatch)
	if err != nil {
		return nil, err
	}

	var api *httpAPIClient
	if httpClient != nil {
		token := cfg.resolvedAccessToken()
		api = newHTTPAPIClient(cfg.APIBaseURL, cfg.Timeout, nil).withHTTPClient(httpClient)
		switch cfg.driver(platformID) {
		case "instagram", "facebook", "threads":
			// graph uses query token
		case "pinterest", "tiktok", "linkedin", "youtube":
			api.authHeader = bearerAuth(token)
		}
		if cfg.driver(platformID) == "linkedin" {
			prev := api.authHeader
			api.authHeader = func(r *http.Request) {
				if prev != nil {
					prev(r)
				}
				applyLinkedInAPIHeaders(r)
			}
		}
		if cfg.driver(platformID) == "youtube" {
			prev := api.authHeader
			api.authHeader = func(r *http.Request) {
				if prev != nil {
					prev(r)
				}
				r.Header.Set("Authorization", "Bearer "+cfg.resolvedAccessToken())
			}
		}
	}

	switch cfg.driver(platformID) {
	case "instagram":
		return newInstagramClient(cfg, templates, dispatch, api), nil
	case "facebook":
		return newFacebookClient(cfg, templates, dispatch, api), nil
	case "threads":
		return newThreadsClient(cfg, templates, dispatch, api), nil
	case "pinterest":
		return newPinterestClient(cfg, templates, dispatch, api), nil
	case "tiktok":
		return newTikTokClient(cfg, templates, dispatch, api), nil
	case "linkedin":
		return newLinkedInClient(cfg, templates, dispatch, api), nil
	case "youtube":
		return newYouTubeClient(cfg, templates, dispatch, api), nil
	default:
		return nil, invalidConfigf("platforms.%s.driver", string(platformID), "unsupported driver %q", cfg.driver(platformID))
	}
}

func buildAllPlatformClients(cfg SocialConfig, templates *TemplateRenderer, httpClient *http.Client) (map[PlatformID]PlatformClient, error) {
	defaultDispatch, err := cfg.defaultDispatch()
	if err != nil {
		return nil, err
	}

	clients := make(map[PlatformID]PlatformClient)
	for rawID, platformCfg := range cfg.Platforms {
		if !platformCfg.isEnabled() {
			continue
		}
		platformID, err := ParsePlatformID(rawID)
		if err != nil {
			return nil, err
		}
		client, err := buildPlatformClient(platformID, platformCfg, templates, defaultDispatch, httpClient)
		if err != nil {
			return nil, fmt.Errorf("platform %s: %w", platformID, err)
		}
		clients[platformID] = client
	}
	return clients, nil
}

// NewPlatformClient builds one platform client from config (e.g. with a fresh OAuth token).
func NewPlatformClient(platformID PlatformID, cfg PlatformConfig, templates *TemplateRenderer, defaultDispatch DispatchMode, httpClient *http.Client) (PlatformClient, error) {
	return buildPlatformClient(platformID, cfg, templates, defaultDispatch, httpClient)
}
