package social

import (
	"context"
	"fmt"
	"strings"
)

type pinterestClient struct {
	basePlatform
	api         *httpAPIClient
	accessToken string
	boardID     string
}

func newPinterestClient(cfg PlatformConfig, templates *TemplateRenderer, defaultDispatch DispatchMode, httpClient *httpAPIClient) *pinterestClient {
	token := cfg.resolvedAccessToken()
	client := httpClient
	if client == nil {
		client = newHTTPAPIClient(cfg.APIBaseURL, cfg.Timeout, bearerAuth(token))
	}
	return &pinterestClient{
		basePlatform: basePlatform{id: PlatformPinterest, cfg: cfg, templates: templates, defaults: defaultDispatch},
		api:          client,
		accessToken:  token,
		boardID:      strings.TrimSpace(cfg.BoardID),
	}
}

func (c *pinterestClient) GetAccountInfo(ctx context.Context) (AccountInfo, error) {
	var resp struct {
		Username string `json:"username"`
	}
	if err := c.api.doJSON(ctx, httpMethodGet, "/user_account", nil, nil, &resp); err != nil {
		return AccountInfo{}, err
	}
	return AccountInfo{
		PlatformID: c.id,
		Username:   resp.Username,
	}, nil
}

func (c *pinterestClient) CreatePost(ctx context.Context, req CreatePostRequest) (CreatePostResult, error) {
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
		return buildClientJob(c.id, formatted, "Open Pinterest pin builder with vertical image and destination link.", map[string]any{
			"board_id":    c.boardID,
			"title":       formatted.Title,
			"description": formatted.Caption,
			"link":        formatted.LinkURL,
			"image_url":   firstMedia(formatted.MediaURLs),
		}), nil
	}

	if firstMedia(formatted.MediaURLs) == "" {
		return CreatePostResult{}, ErrInvalidRequest.With("hero_image_url", "pinterest requires hero_image_url for server publish")
	}

	var resp struct {
		ID string `json:"id"`
	}
	body := map[string]any{
		"board_id":    c.boardID,
		"title":       formatted.Title,
		"description": formatted.Caption,
		"link":        formatted.LinkURL,
		"media_source": map[string]any{
			"source_type": "image_url",
			"url":         firstMedia(formatted.MediaURLs),
		},
	}
	if err := c.api.doJSON(ctx, httpMethodPost, "/pins", nil, body, &resp); err != nil {
		return CreatePostResult{}, wrapErr(ErrPublishFailed, "pinterest.pin", err)
	}

	return CreatePostResult{
		DispatchMode: DispatchServer,
		PostID:       resp.ID,
		PublishedURL: fmt.Sprintf("https://www.pinterest.com/pin/%s/", resp.ID),
	}, nil
}

func (c *pinterestClient) GetPost(ctx context.Context, postID string) (PostInfo, error) {
	postID = strings.TrimSpace(postID)
	if postID == "" {
		return PostInfo{}, ErrInvalidRequest.With("post_id", "required")
	}
	var resp struct {
		ID          string `json:"id"`
		Description string `json:"description"`
	}
	if err := c.api.doJSON(ctx, httpMethodGet, "/pins/"+postID, nil, nil, &resp); err != nil {
		return PostInfo{}, err
	}
	return PostInfo{
		PlatformID:   c.id,
		PostID:       resp.ID,
		Caption:      resp.Description,
		PublishedURL: fmt.Sprintf("https://www.pinterest.com/pin/%s/", resp.ID),
		Status:       "published",
	}, nil
}

func firstMedia(urls []string) string {
	if len(urls) == 0 {
		return ""
	}
	return strings.TrimSpace(urls[0])
}
