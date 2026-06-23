package social

import (
	"context"
	"fmt"
	"strings"
)

type facebookClient struct {
	graphClient
}

func newFacebookClient(cfg PlatformConfig, templates *TemplateRenderer, defaultDispatch DispatchMode, httpClient *httpAPIClient) *facebookClient {
	return &facebookClient{graphClient: *newGraphClient(PlatformFacebook, cfg, templates, defaultDispatch, httpClient)}
}

func (c *facebookClient) CreatePost(ctx context.Context, req CreatePostRequest) (CreatePostResult, error) {
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
		return buildClientJob(c.id, formatted, "Open Facebook Page composer and paste the caption with link preview.", map[string]any{
			"page_id": c.accountID,
			"message": formatted.Caption,
			"link":    formatted.LinkURL,
		}), nil
	}

	q := queryAccessToken(c.accessToken)
	var resp struct {
		ID string `json:"id"`
	}
	body := map[string]any{
		"message": formatted.Caption,
		"link":    formatted.LinkURL,
	}
	if err := c.api.doJSON(ctx, httpMethodPost, c.graphPath(c.accountID+"/feed"), q, body, &resp); err != nil {
		return CreatePostResult{}, wrapErr(ErrPublishFailed, "facebook.feed", err)
	}

	return CreatePostResult{
		DispatchMode: DispatchServer,
		PostID:       resp.ID,
		PublishedURL: fmt.Sprintf("https://www.facebook.com/%s", resp.ID),
	}, nil
}

func (c *facebookClient) GetPost(ctx context.Context, postID string) (PostInfo, error) {
	postID = strings.TrimSpace(postID)
	if postID == "" {
		return PostInfo{}, ErrInvalidRequest.With("post_id", "required")
	}
	var resp struct {
		ID      string `json:"id"`
		Message string `json:"message"`
	}
	q := queryAccessToken(c.accessToken)
	q.Set("fields", "id,message")
	if err := c.api.doJSON(ctx, httpMethodGet, c.graphPath(postID), q, nil, &resp); err != nil {
		return PostInfo{}, err
	}
	return PostInfo{
		PlatformID:   c.id,
		PostID:       resp.ID,
		Caption:      resp.Message,
		PublishedURL: fmt.Sprintf("https://www.facebook.com/%s", resp.ID),
		Status:       "published",
	}, nil
}
