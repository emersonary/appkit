package social

import (
	"context"
	"fmt"
	"strings"
)

type instagramClient struct {
	graphClient
}

func newInstagramClient(cfg PlatformConfig, templates *TemplateRenderer, defaultDispatch DispatchMode, httpClient *httpAPIClient) *instagramClient {
	return &instagramClient{graphClient: *newGraphClient(PlatformInstagram, cfg, templates, defaultDispatch, httpClient)}
}

func (c *instagramClient) CreatePost(ctx context.Context, req CreatePostRequest) (CreatePostResult, error) {
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
		return buildClientJob(c.id, formatted, "Open Instagram and publish carousel or single image with the caption.", map[string]any{
			"account_id": c.accountID,
			"caption":    formatted.Caption,
			"image_urls": formatted.MediaURLs,
		}), nil
	}

	if len(formatted.MediaURLs) == 0 {
		return CreatePostResult{}, ErrInvalidRequest.With("hero_image_url", "instagram requires at least one image URL for server publish")
	}

	createQuery := queryAccessToken(c.accessToken)
	var createResp struct {
		ID string `json:"id"`
	}
	createBody := map[string]any{
		"caption":   formatted.Caption,
		"image_url": formatted.MediaURLs[0],
	}
	if err := c.api.doJSON(ctx, httpMethodPost, c.graphPath(c.accountID+"/media"), createQuery, createBody, &createResp); err != nil {
		return CreatePostResult{}, wrapErr(ErrPublishFailed, "instagram.create_media", err)
	}

	publishQuery := queryAccessToken(c.accessToken)
	var publishResp struct {
		ID string `json:"id"`
	}
	publishBody := map[string]any{"creation_id": createResp.ID}
	if err := c.api.doJSON(ctx, httpMethodPost, c.graphPath(c.accountID+"/media_publish"), publishQuery, publishBody, &publishResp); err != nil {
		return CreatePostResult{}, wrapErr(ErrPublishFailed, "instagram.publish", err)
	}

	return CreatePostResult{
		DispatchMode: DispatchServer,
		PostID:       publishResp.ID,
		PublishedURL: fmt.Sprintf("https://www.instagram.com/p/%s/", publishResp.ID),
	}, nil
}

func (c *instagramClient) GetPost(ctx context.Context, postID string) (PostInfo, error) {
	postID = strings.TrimSpace(postID)
	if postID == "" {
		return PostInfo{}, ErrInvalidRequest.With("post_id", "required")
	}
	var resp struct {
		ID      string `json:"id"`
		Caption string `json:"caption"`
	}
	q := queryAccessToken(c.accessToken)
	q.Set("fields", "id,caption")
	if err := c.api.doJSON(ctx, httpMethodGet, c.graphPath(postID), q, nil, &resp); err != nil {
		return PostInfo{}, err
	}
	return PostInfo{
		PlatformID:   c.id,
		PostID:       resp.ID,
		Caption:      resp.Caption,
		PublishedURL: fmt.Sprintf("https://www.instagram.com/p/%s/", resp.ID),
		Status:       "published",
	}, nil
}
