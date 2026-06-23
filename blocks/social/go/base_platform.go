package social

import (
	"context"
	"strings"
)

type basePlatform struct {
	id        PlatformID
	cfg       PlatformConfig
	templates *TemplateRenderer
	defaults  DispatchMode
}

func (b *basePlatform) PlatformID() PlatformID {
	return b.id
}

func (b *basePlatform) DefaultDispatch() DispatchMode {
	return b.defaults
}

func (b *basePlatform) FormatPost(_ context.Context, input PostInput) (FormattedPost, error) {
	if err := validatePostInput(input); err != nil {
		return FormattedPost{}, err
	}
	fields := BuildFields(input, b.cfg)
	caption, err := b.templates.Render(b.id, fields)
	if err != nil {
		return FormattedPost{}, err
	}
	media := []string{}
	if hero := strings.TrimSpace(input.HeroImageURL); hero != "" {
		media = append(media, hero)
	}
	return FormattedPost{
		PlatformID: b.id,
		Caption:    caption,
		Title:      strings.TrimSpace(input.Title),
		LinkURL:    strings.TrimSpace(input.ArticleURL),
		MediaURLs:  media,
		VideoURL:   strings.TrimSpace(input.VideoURL),
		Fields:     fields,
	}, nil
}

func validatePostInput(input PostInput) error {
	if strings.TrimSpace(input.IntroText) == "" {
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

func resolveDispatch(req CreatePostRequest, defaultMode DispatchMode) DispatchMode {
	if req.DispatchMode != "" {
		return req.DispatchMode
	}
	return defaultMode
}

func buildClientJob(id PlatformID, formatted FormattedPost, instructions string, payload map[string]any) CreatePostResult {
	return CreatePostResult{
		DispatchMode: DispatchClient,
		ClientJob: &ClientPublishJob{
			PlatformID:   id,
			Caption:      formatted.Caption,
			Title:        formatted.Title,
			LinkURL:      formatted.LinkURL,
			MediaURLs:    append([]string(nil), formatted.MediaURLs...),
			VideoURL:     formatted.VideoURL,
			Instructions: instructions,
			Payload:      payload,
		},
	}
}
