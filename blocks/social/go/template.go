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
	templates map[PlatformID]ParsedTemplate
}

// NewTemplateRenderer loads embedded templates for all platforms.
func NewTemplateRenderer() (*TemplateRenderer, error) {
	templates := make(map[PlatformID]ParsedTemplate, len(DefaultPlatforms))
	for _, id := range DefaultPlatforms {
		path := fmt.Sprintf("post-templates/%s.template.txt", id)
		data, err := templateFS.ReadFile(path)
		if err != nil {
			return nil, wrapErr(ErrLoadConfig, "templates", fmt.Errorf("%s: %w", path, err))
		}
		parsed, err := parseTemplate(string(data))
		if err != nil {
			return nil, wrapErr(ErrLoadConfig, "templates", fmt.Errorf("%s: %w", path, err))
		}
		templates[id] = parsed
	}
	return &TemplateRenderer{templates: templates}, nil
}

// Render applies template fields for a platform using an empty context (global content only).
func (r *TemplateRenderer) Render(id PlatformID, fields TemplateFields) (string, error) {
	return r.RenderWithContext(id, fields, TemplateContext{Platform: id})
}

// RenderWithContext applies fields and renders global content plus matching session blocks.
func (r *TemplateRenderer) RenderWithContext(id PlatformID, fields TemplateFields, ctx TemplateContext) (string, error) {
	parsed, ok := r.templates[id]
	if !ok {
		return "", ErrPlatformNotFound.With("platform", string(id))
	}
	body := composeTemplateBody(parsed, ctx)
	out := body
	for key, value := range fields {
		out = strings.ReplaceAll(out, "{{"+key+"}}", value)
	}
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
	hero := resolveHeroMediaURL(input)
	cover := strings.TrimSpace(input.CoverImageURL)
	if cover == "" {
		cover = hero
	}
	video := strings.TrimSpace(input.VideoURL)
	linkLine := brandLinkLine(input.Language, brand, articleURL)

	fullText := htmlToPlainText(input.BodyHTML)
	if fullText == "" {
		fullText = intro
	}

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
		"full_text":                fullText,
		"article_url":              articleURL,
		"source_brand":             brand,
		"source_url":               sourceURL,
		"link_line":                linkLine,
		"read_on_line":             linkLine,
		"hero_image_url":           hero,
		"video_url":                video,
		"hashtags":                 hashtags,
		"link_preview_title":       title,
		"link_preview_description": excerpt,
		"article_excerpt":          excerpt,
		"cover_image_url":          cover,
		"board_name":               strings.TrimSpace(cfg.BoardName),
		"thread_reply_2":           fmt.Sprintf("Leia completo no %s: %s", brand, articleURL),
		"first_comment":            fmt.Sprintf("%s", articleURL),
	}
	if fields["board_name"] == "" {
		fields["board_name"] = "Blog"
	}
	return fields
}
