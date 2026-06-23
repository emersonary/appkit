package social

import (
	"context"
	"strings"
)

type tiktokClient struct {
	basePlatform
	api         *httpAPIClient
	accessToken string
	accountID   string
}

func newTikTokClient(cfg PlatformConfig, templates *TemplateRenderer, defaultDispatch DispatchMode, httpClient *httpAPIClient) *tiktokClient {
	token := cfg.resolvedAccessToken()
	client := httpClient
	if client == nil {
		client = newHTTPAPIClient(cfg.APIBaseURL, cfg.Timeout, bearerAuth(token))
	}
	return &tiktokClient{
		basePlatform: basePlatform{id: PlatformTikTok, cfg: cfg, templates: templates, defaults: defaultDispatch},
		api:          client,
		accessToken:  token,
		accountID:    strings.TrimSpace(cfg.AccountID),
	}
}

func (c *tiktokClient) GetAccountInfo(ctx context.Context) (AccountInfo, error) {
	var resp struct {
		Data struct {
			User struct {
				OpenID      string `json:"open_id"`
				DisplayName string `json:"display_name"`
			} `json:"user"`
		} `json:"data"`
	}
	if err := c.api.doJSON(ctx, httpMethodPost, "/v2/user/info/", nil, map[string]any{
		"fields": []string{"open_id", "display_name"},
	}, &resp); err != nil {
		return AccountInfo{}, err
	}
	return AccountInfo{
		PlatformID:  c.id,
		AccountID:   resp.Data.User.OpenID,
		DisplayName: resp.Data.User.DisplayName,
	}, nil
}

func (c *tiktokClient) CreatePost(ctx context.Context, req CreatePostRequest) (CreatePostResult, error) {
	formatted := req.Formatted
	if formatted.PlatformID == "" {
		var err error
		formatted, err = c.FormatPost(ctx, req.Input)
		if err != nil {
			return CreatePostResult{}, err
		}
	}

	mode := resolveDispatch(req, c.DefaultDispatch())
	if mode == DispatchClient || formatted.VideoURL == "" {
		return buildClientJob(c.id, formatted, "Upload video in TikTok app with caption; pin article link in comment if needed.", map[string]any{
			"open_id":    c.accountID,
			"caption":    formatted.Caption,
			"video_url":  formatted.VideoURL,
			"cover_url":  firstMedia(formatted.MediaURLs),
		}), nil
	}

	var resp struct {
		Data struct {
			PublishID string `json:"publish_id"`
		} `json:"data"`
	}
	body := map[string]any{
		"post_info": map[string]any{
			"title":       formatted.Caption,
			"privacy_level": "PUBLIC_TO_EVERYONE",
		},
		"source_info": map[string]any{
			"source":    "PULL_FROM_URL",
			"video_url": formatted.VideoURL,
		},
	}
	if err := c.api.doJSON(ctx, httpMethodPost, "/v2/post/publish/video/init/", nil, body, &resp); err != nil {
		return CreatePostResult{}, wrapErr(ErrPublishFailed, "tiktok.publish", err)
	}

	return CreatePostResult{
		DispatchMode: DispatchServer,
		PostID:       resp.Data.PublishID,
	}, nil
}

func (c *tiktokClient) GetPost(ctx context.Context, postID string) (PostInfo, error) {
	postID = strings.TrimSpace(postID)
	if postID == "" {
		return PostInfo{}, ErrInvalidRequest.With("post_id", "required")
	}
	var resp struct {
		Data struct {
			Videos []struct {
				ID    string `json:"id"`
				Title string `json:"title"`
			} `json:"videos"`
		} `json:"data"`
	}
	if err := c.api.doJSON(ctx, httpMethodPost, "/v2/video/query/", nil, map[string]any{
		"filters": map[string]any{"video_ids": []string{postID}},
	}, &resp); err != nil {
		return PostInfo{}, err
	}
	if len(resp.Data.Videos) == 0 {
		return PostInfo{}, ErrAPIFailed.With("post_id", "not found")
	}
	v := resp.Data.Videos[0]
	return PostInfo{
		PlatformID: c.id,
		PostID:     v.ID,
		Caption:    v.Title,
		Status:     "published",
	}, nil
}
