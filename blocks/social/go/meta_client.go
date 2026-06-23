package social

import (
	"context"
	"strings"
)

type graphClient struct {
	basePlatform
	api         *httpAPIClient
	accessToken string
	accountID   string
}

func newGraphClient(id PlatformID, cfg PlatformConfig, templates *TemplateRenderer, defaultDispatch DispatchMode, httpClient *httpAPIClient) *graphClient {
	token := cfg.resolvedAccessToken()
	client := httpClient
	if client == nil {
		client = newHTTPAPIClient(cfg.APIBaseURL, cfg.Timeout, nil)
	}
	return &graphClient{
		basePlatform: basePlatform{id: id, cfg: cfg, templates: templates, defaults: defaultDispatch},
		api:          client,
		accessToken:  token,
		accountID:    strings.TrimSpace(cfg.AccountID),
	}
}

func (c *graphClient) graphPath(path string) string {
	return "/" + strings.TrimLeft(path, "/")
}

func (c *graphClient) GetAccountInfo(ctx context.Context) (AccountInfo, error) {
	var resp struct {
		ID       string `json:"id"`
		Name     string `json:"name"`
		Username string `json:"username"`
	}
	q := queryAccessToken(c.accessToken)
	q.Set("fields", "id,name,username")
	if err := c.api.doJSON(ctx, httpMethodGet, c.graphPath(c.accountID), q, nil, &resp); err != nil {
		return AccountInfo{}, err
	}
	return AccountInfo{
		PlatformID:  c.id,
		AccountID:   resp.ID,
		DisplayName: resp.Name,
		Username:    resp.Username,
	}, nil
}

const httpMethodGet = "GET"
const httpMethodPost = "POST"
