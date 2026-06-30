package social

import (
	"fmt"
	"strings"
)

// TemplateContext selects session blocks during caption rendering.
type TemplateContext struct {
	Platform PlatformID
	Filters  map[string]string
}

// TemplateSession is a conditional block inside a platform template file.
type TemplateSession struct {
	Filters map[string]string
	Body    string
}

// ParsedTemplate is a platform caption template split into global and session parts.
type ParsedTemplate struct {
	Global   string
	Sessions []TemplateSession
}

func parseTemplate(raw string) (ParsedTemplate, error) {
	var out ParsedTemplate
	var globalLines []string
	var sessionLines []string
	var current *TemplateSession
	inSession := false

	flushSession := func() error {
		if current == nil {
			return nil
		}
		current.Body = strings.TrimSpace(strings.Join(sessionLines, "\n"))
		out.Sessions = append(out.Sessions, *current)
		current = nil
		sessionLines = nil
		inSession = false
		return nil
	}

	for _, line := range strings.Split(raw, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "#") {
			continue
		}

		if strings.HasPrefix(trimmed, "@session") {
			if inSession {
				return ParsedTemplate{}, fmt.Errorf("nested @session is not allowed")
			}
			if err := flushSession(); err != nil {
				return ParsedTemplate{}, err
			}
			filters, err := parseSessionFilters(strings.TrimSpace(strings.TrimPrefix(trimmed, "@session")))
			if err != nil {
				return ParsedTemplate{}, err
			}
			current = &TemplateSession{Filters: filters}
			inSession = true
			continue
		}

		if trimmed == "@end" || strings.HasPrefix(trimmed, "@end ") {
			if !inSession {
				return ParsedTemplate{}, fmt.Errorf("@end without matching @session")
			}
			if err := flushSession(); err != nil {
				return ParsedTemplate{}, err
			}
			continue
		}

		if inSession {
			sessionLines = append(sessionLines, line)
		} else {
			globalLines = append(globalLines, line)
		}
	}

	if inSession {
		return ParsedTemplate{}, fmt.Errorf("unclosed @session")
	}
	if err := flushSession(); err != nil {
		return ParsedTemplate{}, err
	}

	out.Global = strings.TrimSpace(strings.Join(globalLines, "\n"))
	return out, nil
}

func parseSessionFilters(raw string) (map[string]string, error) {
	if raw == "" {
		return map[string]string{}, nil
	}
	out := make(map[string]string)
	for _, part := range strings.Fields(raw) {
		key, value, ok := strings.Cut(part, "=")
		if !ok || strings.TrimSpace(key) == "" {
			return nil, fmt.Errorf("invalid session filter %q", part)
		}
		out[strings.TrimSpace(key)] = strings.TrimSpace(value)
	}
	return out, nil
}

func sessionMatches(session TemplateSession, ctx TemplateContext) bool {
	if len(session.Filters) == 0 {
		return true
	}
	for key, want := range session.Filters {
		got, ok := ctx.Filters[key]
		if !ok || got != want {
			return false
		}
	}
	return true
}

func composeTemplateBody(parsed ParsedTemplate, ctx TemplateContext) string {
	var parts []string
	if strings.TrimSpace(parsed.Global) != "" {
		parts = append(parts, parsed.Global)
	}
	for _, session := range parsed.Sessions {
		if !sessionMatches(session, ctx) {
			continue
		}
		if body := strings.TrimSpace(session.Body); body != "" {
			parts = append(parts, body)
		}
	}
	return strings.TrimSpace(strings.Join(parts, "\n\n"))
}

func linkedInModeFilter(mode linkedInMode) string {
	if mode == linkedInModeLink {
		return "link"
	}
	return "native"
}
