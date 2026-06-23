package social

import (
	"context"
	"fmt"
	"strings"
)

type threadsClient struct {
	graphClient
}

func newThreadsClient(cfg PlatformConfig, templates *TemplateRenderer, defaultDispatch DispatchMode, httpClient *httpAPIClient) *threadsClient {
	return &threadsClient{graphClient: *newGraphClient(PlatformThreads, cfg, templates, defaultDispatch, httpClient)}
}

func (c *threadsClient) CreatePost(ctx context.Context, req CreatePostRequest) (CreatePostResult, error) {
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
		return buildClientJob(c.id, formatted, "Open Threads and publish text (500 char limit).", map[string]any{
			"account_id": c.accountID,
			"text":       formatted.Caption,
		}), nil
	}

	q := queryAccessToken(c.accessToken)
	var resp struct {
		ID string `json:"id"`
	}
	body := map[string]any{
		"media_type": "TEXT",
		"text":       formatted.Caption,
	}
	if err := c.api.doJSON(ctx, httpMethodPost, c.graphPath(c.accountID+"/threads"), q, body, &resp); err != nil {
		return CreatePostResult{}, wrapErr(ErrPublishFailed, "threads.create", err)
	}

	return CreatePostResult{
		DispatchMode: DispatchServer,
		PostID:       resp.ID,
		PublishedURL: fmt.Sprintf("https://www.threads.net/@me/post/%s", resp.ID),
	}, nil
}

func (c *threadsClient) GetPost(ctx context.Context, postID string) (PostInfo, error) {
	postID = strings.TrimSpace(postID)
	if postID == "" {
		return PostInfo{}, ErrInvalidRequest.With("post_id", "required")
	}
	var resp struct {
		ID   string `json:"id"`
		Text string `json:"text"`
	}
	q := queryAccessToken(c.accessToken)
	q.Set("fields", "id,text")
	if err := c.api.doJSON(ctx, httpMethodGet, c.graphPath(postID), q, nil, &resp); err != nil {
		return PostInfo{}, err
	}
	return PostInfo{
		PlatformID:   c.id,
		PostID:       resp.ID,
		Caption:      resp.Text,
		PublishedURL: fmt.Sprintf("https://www.threads.net/@me/post/%s", resp.ID),
		Status:       "published",
	}, nil
}
