package social

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type youtubeClient struct {
	basePlatform
	api         *httpAPIClient
	accessToken string
	channelID   string
	postType    string
}

func newYouTubeClient(cfg PlatformConfig, templates *TemplateRenderer, defaultDispatch DispatchMode, httpClient *httpAPIClient) *youtubeClient {
	token := cfg.resolvedAccessToken()
	client := httpClient
	if client == nil {
		client = newHTTPAPIClient(cfg.APIBaseURL, cfg.Timeout, func(r *http.Request) {
			r.Header.Set("Authorization", "Bearer "+token)
		})
	}
	return &youtubeClient{
		basePlatform: basePlatform{id: PlatformYouTube, cfg: cfg, templates: templates, defaults: defaultDispatch},
		api:          client,
		accessToken:  token,
		channelID:    strings.TrimSpace(cfg.AccountID),
		postType:     strings.TrimSpace(cfg.YouTubePostType),
	}
}

func (c *youtubeClient) GetAccountInfo(ctx context.Context) (AccountInfo, error) {
	q := url.Values{}
	q.Set("part", "snippet,statistics")
	q.Set("id", c.channelID)
	var resp struct {
		Items []struct {
			ID      string `json:"id"`
			Snippet struct {
				Title string `json:"title"`
			} `json:"snippet"`
			Statistics struct {
				SubscriberCount string `json:"subscriberCount"`
			} `json:"statistics"`
		} `json:"items"`
	}
	if err := c.api.doJSON(ctx, httpMethodGet, "/channels", q, nil, &resp); err != nil {
		return AccountInfo{}, err
	}
	if len(resp.Items) == 0 {
		return AccountInfo{}, ErrAPIFailed.With("channel_id", "not found")
	}
	item := resp.Items[0]
	return AccountInfo{
		PlatformID:  c.id,
		AccountID:   item.ID,
		DisplayName: item.Snippet.Title,
	}, nil
}

func (c *youtubeClient) CreatePost(ctx context.Context, req CreatePostRequest) (CreatePostResult, error) {
	formatted := req.Formatted
	if formatted.PlatformID == "" {
		var err error
		formatted, err = c.FormatPost(ctx, req.Input)
		if err != nil {
			return CreatePostResult{}, err
		}
	}

	mode := resolveDispatch(req, c.DefaultDispatch())
	contentKind := strings.ToLower(strings.TrimSpace(req.Input.ContentKind))
	useVideo := contentKind == "video" || formatted.VideoURL != "" || c.postType == "video"

	if mode == DispatchClient || !useVideo {
		instructions := "Open YouTube Community tab and publish the formatted text with optional image."
		if useVideo {
			instructions = "Upload video in YouTube Studio with title, description, and thumbnail."
		}
		return buildClientJob(c.id, formatted, instructions, map[string]any{
			"channel_id": c.channelID,
			"post_type":  c.postType,
			"title":      formatted.Title,
			"caption":    formatted.Caption,
			"video_url":  formatted.VideoURL,
			"image_url":  firstMedia(formatted.MediaURLs),
		}), nil
	}

	q := url.Values{}
	q.Set("part", "snippet,status")
	var resp struct {
		ID string `json:"id"`
	}
	body := map[string]any{
		"snippet": map[string]any{
			"title":       formatted.Title,
			"description": formatted.Caption,
			"channelId":   c.channelID,
		},
		"status": map[string]any{
			"privacyStatus": "public",
		},
	}
	if err := c.api.doJSON(ctx, httpMethodPost, "/videos", q, body, &resp); err != nil {
		return CreatePostResult{}, wrapErr(ErrPublishFailed, "youtube.video", err)
	}

	return CreatePostResult{
		DispatchMode: DispatchServer,
		PostID:       resp.ID,
		PublishedURL: fmt.Sprintf("https://www.youtube.com/watch?v=%s", resp.ID),
	}, nil
}

func (c *youtubeClient) GetPost(ctx context.Context, postID string) (PostInfo, error) {
	postID = strings.TrimSpace(postID)
	if postID == "" {
		return PostInfo{}, ErrInvalidRequest.With("post_id", "required")
	}
	q := url.Values{}
	q.Set("part", "snippet")
	q.Set("id", postID)
	var resp struct {
		Items []struct {
			ID      string `json:"id"`
			Snippet struct {
				Title       string `json:"title"`
				Description string `json:"description"`
			} `json:"snippet"`
		} `json:"items"`
	}
	if err := c.api.doJSON(ctx, httpMethodGet, "/videos", q, nil, &resp); err != nil {
		return PostInfo{}, err
	}
	if len(resp.Items) == 0 {
		return PostInfo{}, ErrAPIFailed.With("post_id", "not found")
	}
	item := resp.Items[0]
	return PostInfo{
		PlatformID:   c.id,
		PostID:       item.ID,
		Caption:      item.Snippet.Description,
		PublishedURL: fmt.Sprintf("https://www.youtube.com/watch?v=%s", item.ID),
		Status:       "published",
	}, nil
}
