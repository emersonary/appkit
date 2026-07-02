package ai

import "regexp"

var translationHTMLTagRE = regexp.MustCompile(`(?i)<[a-z][^>]*>`)

func translationLooksLikeHTML(text string) bool {
	return translationHTMLTagRE.MatchString(text)
}
