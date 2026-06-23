package permissions

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

// validatePermissionID rejects dots — tree paths are derived from id_parent, not dotted ids.
func validatePermissionID(id string) error {
	if strings.Contains(id, ".") {
		return fmt.Errorf("permission id %q must not contain '.'", id)
	}
	return validateIdent(id)
}

func quoteIdent(name string) string {
	return `"` + strings.ReplaceAll(name, `"`, `""`) + `"`
}

func qualifiedName(schema, object string) string {
	return quoteIdent(schema) + "." + quoteIdent(object)
}

func quoteLiteral(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "''") + "'"
}
