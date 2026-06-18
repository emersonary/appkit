package tenants

import (
	"fmt"
	"regexp"
	"strings"
)

var identPattern = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)
var slugPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
var feedIDPattern = regexp.MustCompile(`^[a-z][a-z0-9]*$`)

func validateIdent(name string) error {
	if !identPattern.MatchString(name) {
		return fmt.Errorf("invalid identifier %q", name)
	}
	return nil
}

func validateSlug(slug string) error {
	slug = strings.TrimSpace(strings.ToLower(slug))
	if !slugPattern.MatchString(slug) {
		return fmt.Errorf("invalid slug %q", slug)
	}
	return nil
}

func validateFeedID(id string) error {
	id = strings.TrimSpace(strings.ToLower(id))
	if !feedIDPattern.MatchString(id) {
		return fmt.Errorf("invalid feed id %q", id)
	}
	return nil
}

func quoteIdent(name string) string {
	return `"` + strings.ReplaceAll(name, `"`, `""`) + `"`
}

func qualifiedName(schema, object string) string {
	return quoteIdent(schema) + "." + quoteIdent(object)
}

func normalizeSlug(slug string) string {
	return strings.TrimSpace(strings.ToLower(slug))
}
