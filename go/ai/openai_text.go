package ai

import "strings"

// cleanModelText strips common LLM wrappers (markdown fences, quotes).
func cleanModelText(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	if strings.HasPrefix(raw, "```") {
		lines := strings.Split(raw, "\n")
		if len(lines) >= 2 {
			start := 1
			end := len(lines)
			if strings.TrimSpace(lines[len(lines)-1]) == "```" {
				end = len(lines) - 1
			}
			raw = strings.TrimSpace(strings.Join(lines[start:end], "\n"))
		}
	}
	return strings.Trim(raw, `"`)
}
