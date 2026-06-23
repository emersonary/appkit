package social

import (
	"embed"
	"fmt"
	"strings"
)

//go:embed post-templates/*.template.txt
var templateFS embed.FS

// TemplateRenderer loads and renders per-platform caption templates.
type TemplateRenderer struct {
	bodies map[PlatformID]string
}

// NewTemplateRenderer loads embedded templates for all platforms.
func NewTemplateRenderer() (*TemplateRenderer, error) {
	bodies := make(map[PlatformID]string, len(DefaultPlatforms))
	for _, id := range DefaultPlatforms {
		path := fmt.Sprintf("post-templates/%s.template.txt", id)
		data, err := templateFS.ReadFile(path)
		if err != nil {
			return nil, wrapErr(ErrLoadConfig, "templates", fmt.Errorf("%s: %w", path, err))
		}
		bodies[id] = extractTemplateBody(string(data))
	}
	return &TemplateRenderer{bodies: bodies}, nil
}

func extractTemplateBody(raw string) string {
	var lines []string
	for _, line := range strings.Split(raw, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			lines = append(lines, "")
			continue
		}
		// Comment lines are documentation only — never part of the published caption.
		if strings.HasPrefix(trimmed, "#") {
			continue
		}
		lines = append(lines, line)
	}

	body := strings.TrimSpace(strings.Join(lines, "\n"))
	for strings.Contains(body, "\n\n\n") {
		body = strings.ReplaceAll(body, "\n\n\n", "\n\n")
	}
	return body
}

// Render applies template fields for a platform.
func (r *TemplateRenderer) Render(id PlatformID, fields TemplateFields) (string, error) {
	body, ok := r.bodies[id]
	if !ok {
		return "", ErrPlatformNotFound.With("platform", string(id))
	}
	out := body
	for key, value := range fields {
		out = strings.ReplaceAll(out, "{{"+key+"}}", value)
	}
	// Remove leftover optional placeholders.
	for {
		start := strings.Index(out, "{{")
		if start < 0 {
			break
		}
		end := strings.Index(out[start:], "}}")
		if end < 0 {
			break
		}
		out = out[:start] + out[start+end+2:]
	}
	return collapseCaption(out), nil
}

func collapseCaption(raw string) string {
	var lines []string
	for _, line := range strings.Split(raw, "\n") {
		if strings.TrimSpace(line) == "" {
			if len(lines) > 0 && lines[len(lines)-1] == "" {
				continue
			}
			lines = append(lines, "")
			continue
		}
		lines = append(lines, line)
	}
	return strings.TrimSpace(strings.Join(lines, "\n"))
}

func brandLinkLine(language, brand, articleURL string) string {
	switch strings.ToLower(strings.TrimSpace(language)) {
	case "pt":
		return fmt.Sprintf("Leia no %s: %s", brand, articleURL)
	case "es":
		return fmt.Sprintf("Lee en %s: %s", brand, articleURL)
	case "fr":
		return fmt.Sprintf("Lire sur %s : %s", brand, articleURL)
	case "de":
		return fmt.Sprintf("Lesen auf %s: %s", brand, articleURL)
	default:
		return fmt.Sprintf("Read on %s: %s", brand, articleURL)
	}
}

// BuildFields maps PostInput into template variables with sensible defaults.
func BuildFields(input PostInput, cfg PlatformConfig) TemplateFields {
	intro := strings.TrimSpace(input.IntroText)
	title := strings.TrimSpace(input.Title)
	articleURL := strings.TrimSpace(input.ArticleURL)
	brand := strings.TrimSpace(input.SourceBrand)
	sourceURL := strings.TrimSpace(input.SourceURL)
	hashtags := strings.TrimSpace(input.Hashtags)
	hero := strings.TrimSpace(input.HeroImageURL)
	video := strings.TrimSpace(input.VideoURL)

	excerpt := intro
	if idx := strings.Index(excerpt, "\n"); idx >= 0 {
		excerpt = strings.TrimSpace(excerpt[:idx])
	}
	if len(excerpt) > 160 {
		excerpt = strings.TrimSpace(excerpt[:157]) + "..."
	}

	fields := TemplateFields{
		"title":                    title,
		"intro_text":               intro,
		"article_url":              articleURL,
		"source_brand":             brand,
		"source_url":               sourceURL,
		"link_line":                brandLinkLine(input.Language, brand, articleURL),
		"hero_image_url":           hero,
		"video_url":                video,
		"hashtags":                 hashtags,
		"link_preview_title":       title,
		"link_preview_description": excerpt,
		"article_excerpt":          excerpt,
		"cover_image_url":          hero,
		"board_name":               strings.TrimSpace(cfg.BoardName),
		"thread_reply_2":           fmt.Sprintf("Leia completo no %s: %s", brand, articleURL),
		"first_comment":            fmt.Sprintf("%s", articleURL),
	}
	if fields["board_name"] == "" {
		fields["board_name"] = "Blog"
	}
	return fields
}
