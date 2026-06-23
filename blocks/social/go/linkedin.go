package social

import (
	"context"
	"fmt"
	"net/http"
	"strings"
)

type linkedInClient struct {
	basePlatform
	api       *httpAPIClient
	authorURN string
}

const defaultLinkedInAPIVersion = "202606"

func applyLinkedInAPIHeaders(r *http.Request) {
	r.Header.Set("LinkedIn-Version", defaultLinkedInAPIVersion)
	r.Header.Set("X-Restli-Protocol-Version", "2.0.0")
}

func newLinkedInClient(cfg PlatformConfig, templates *TemplateRenderer, defaultDispatch DispatchMode, httpClient *httpAPIClient) *linkedInClient {
	token := cfg.resolvedAccessToken()
	client := httpClient
	if client == nil {
		client = newHTTPAPIClient(cfg.APIBaseURL, cfg.Timeout, func(r *http.Request) {
			r.Header.Set("Authorization", "Bearer "+token)
			applyLinkedInAPIHeaders(r)
		})
	}
	author := strings.TrimSpace(cfg.AccountID)
	author = linkedInAuthorURN(author)
	return &linkedInClient{
		basePlatform: basePlatform{id: PlatformLinkedIn, cfg: cfg, templates: templates, defaults: defaultDispatch},
		api:          client,
		authorURN:    author,
	}
}

func (c *linkedInClient) GetAccountInfo(ctx context.Context) (AccountInfo, error) {
	var resp struct {
		ID       string `json:"id"`
		LocalizedName string `json:"localizedName"`
	}
	path := "/rest/organizations/" + strings.TrimPrefix(c.authorURN, "urn:li:organization:")
	if err := c.api.doJSON(ctx, httpMethodGet, path, nil, nil, &resp); err != nil {
		return AccountInfo{}, err
	}
	return AccountInfo{
		PlatformID:  c.id,
		AccountID:   c.authorURN,
		DisplayName: resp.LocalizedName,
	}, nil
}

func (c *linkedInClient) CreatePost(ctx context.Context, req CreatePostRequest) (CreatePostResult, error) {
	formatted := req.Formatted
	if formatted.PlatformID == "" {
		var err error
		formatted, err = c.FormatPost(ctx, req.Input)
		if err != nil {
			return CreatePostResult{}, err
		}
	}

	mode := resolveDispatch(req, c.DefaultDispatch())
	if mode == DispatchClient {
		return buildClientJob(c.id, formatted, "Open LinkedIn and create a post with the formatted caption and link.", map[string]any{
			"author":  c.authorURN,
			"text":    formatted.Caption,
			"link":    formatted.LinkURL,
		}), nil
	}

	var resp struct {
		ID string `json:"id"`
	}
	body := map[string]any{
		"author":         c.authorURN,
		"commentary":     formatted.Caption,
		"visibility":     "PUBLIC",
		"lifecycleState": "PUBLISHED",
		"distribution": map[string]any{
			"feedDistribution":               "MAIN_FEED",
			"targetEntities":                 []any{},
			"thirdPartyDistributionChannels": []any{},
		},
	}
	if formatted.LinkURL != "" {
		body["content"] = map[string]any{
			"article": map[string]any{
				"source": formatted.LinkURL,
				"title":  formatted.Title,
			},
		}
	}
	if err := c.api.doJSON(ctx, httpMethodPost, "/rest/posts", nil, body, &resp); err != nil {
		return CreatePostResult{}, wrapErr(ErrPublishFailed, "linkedin.post", err)
	}

	return CreatePostResult{
		DispatchMode: DispatchServer,
		PostID:       resp.ID,
		PublishedURL: fmt.Sprintf("https://www.linkedin.com/feed/update/%s", resp.ID),
	}, nil
}

func (c *linkedInClient) GetPost(ctx context.Context, postID string) (PostInfo, error) {
	postID = strings.TrimSpace(postID)
	if postID == "" {
		return PostInfo{}, ErrInvalidRequest.With("post_id", "required")
	}
	var resp struct {
		ID         string `json:"id"`
		Commentary string `json:"commentary"`
	}
	if err := c.api.doJSON(ctx, httpMethodGet, "/rest/posts/"+postID, nil, nil, &resp); err != nil {
		return PostInfo{}, err
	}
	return PostInfo{
		PlatformID:   c.id,
		PostID:       resp.ID,
		Caption:      resp.Commentary,
		PublishedURL: fmt.Sprintf("https://www.linkedin.com/feed/update/%s", resp.ID),
		Status:       "published",
	}, nil
}

func linkedInAuthorURN(accountID string) string {
	accountID = strings.TrimSpace(accountID)
	if accountID == "" {
		return ""
	}
	if strings.HasPrefix(accountID, "urn:li:") {
		return accountID
	}
	return "urn:li:organization:" + accountID
}
