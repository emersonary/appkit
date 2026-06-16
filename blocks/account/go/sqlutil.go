package accounts

import (
	"fmt"
	"regexp"
	"strings"
)

var identPattern = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)

func validateIdent(name string) error {
	if !identPattern.MatchString(name) {
		return fmt.Errorf("invalid identifier %q", name)
	}
	return nil
}

func quoteIdent(name string) string {
	return `"` + strings.ReplaceAll(name, `"`, `""`) + `"`
}

func qualifiedName(schema, object string) string {
	return quoteIdent(schema) + "." + quoteIdent(object)
}
