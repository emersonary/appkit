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

func (c *linkedInClient) FormatPost(ctx context.Context, input PostInput) (FormattedPost, error) {
	if err := validateLinkedInPostInput(input); err != nil {
		return FormattedPost{}, err
	}
	mode := resolveLinkedInMode(input)
	fields := BuildFields(input, c.cfg)
	caption, err := c.templates.RenderWithContext(c.id, fields, TemplateContext{
		Platform: c.id,
		Filters:  map[string]string{"type": linkedInModeFilter(mode)},
	})
	if err != nil {
		return FormattedPost{}, err
	}
	media := []string{}
	if hero := resolveHeroMediaURL(input); hero != "" {
		media = append(media, hero)
	}
	return FormattedPost{
		PlatformID:         c.id,
		Caption:            caption,
		Title:              strings.TrimSpace(input.Title),
		LinkURL:            strings.TrimSpace(input.ArticleURL),
		MediaURLs:          media,
		VideoURL:           strings.TrimSpace(input.VideoURL),
		IncludeArticleLink: mode == linkedInModeLink,
		Fields:             fields,
	}, nil
}

func validateLinkedInPostInput(input PostInput) error {
	if strings.TrimSpace(input.IntroText) == "" && strings.TrimSpace(input.BodyHTML) == "" {
		return ErrInvalidRequest.With("intro_text", "required")
	}
	if strings.TrimSpace(input.ArticleURL) == "" {
		return ErrInvalidRequest.With("article_url", "required")
	}
	if strings.TrimSpace(input.SourceBrand) == "" {
		return ErrInvalidRequest.With("source_brand", "required")
	}
	return nil
}

func linkedInIncludeArticleLink(formatted FormattedPost) bool {
	if formatted.IncludeArticleLink {
		return strings.TrimSpace(formatted.LinkURL) != ""
	}
	// Legacy formatted payloads saved before IncludeArticleLink existed.
	if strings.TrimSpace(formatted.LinkURL) == "" {
		return false
	}
	return strings.Contains(formatted.Caption, formatted.LinkURL)
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
	if linkedInIncludeArticleLink(formatted) {
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
