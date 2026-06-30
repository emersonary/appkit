package social

import (
	"html"
	"regexp"
	"strings"
	"unicode/utf8"
)

const linkedInShortPostRunes = 1200

type linkedInMode int

const (
	linkedInModeLink linkedInMode = iota
	linkedInModeNative
)

var (
	linkedInRichContentRE = regexp.MustCompile(`(?i)<(img|pre|code|video|iframe)\b`)
	linkedInBlockBreakRE  = regexp.MustCompile(`(?i)</(p|div|h[1-6]|li|br|tr|blockquote)>`)
	linkedInTagRE         = regexp.MustCompile(`<[^>]*>`)
	linkedInSpaceRE       = regexp.MustCompile(`[ \t]+`)
)

func resolveLinkedInMode(input PostInput) linkedInMode {
	plain := htmlToPlainText(input.BodyHTML)
	if linkedInPlainTextRunes(plain) <= linkedInShortPostRunes {
		return linkedInModeNative
	}
	if linkedInHasRichContent(input.BodyHTML) {
		return linkedInModeLink
	}
	return linkedInModeNative
}

// LinkedInModeForInput returns "link" or "native" for the given post input.
func LinkedInModeForInput(input PostInput) string {
	if resolveLinkedInMode(input) == linkedInModeLink {
		return "link"
	}
	return "native"
}

func linkedInHasRichContent(bodyHTML string) bool {
	return linkedInRichContentRE.MatchString(bodyHTML)
}

func linkedInPlainTextRunes(text string) int {
	return utf8.RuneCountInString(strings.TrimSpace(text))
}

func htmlToPlainText(raw string) string {
	if strings.TrimSpace(raw) == "" {
		return ""
	}
	text := linkedInBlockBreakRE.ReplaceAllString(raw, "\n")
	text = linkedInTagRE.ReplaceAllString(text, "")
	text = html.UnescapeString(text)
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")

	var lines []string
	for _, line := range strings.Split(text, "\n") {
		line = linkedInSpaceRE.ReplaceAllString(strings.TrimSpace(line), " ")
		if line == "" {
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

func resolveHeroMediaURL(input PostInput) string {
	if cover := strings.TrimSpace(input.CoverImageURL); cover != "" {
		return cover
	}
	return strings.TrimSpace(input.HeroImageURL)
}
