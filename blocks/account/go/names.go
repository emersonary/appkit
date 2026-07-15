package accounts

import "strings"

// SplitFullName splits a single display string into first and last name.
func SplitFullName(name string) (first, last string) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", ""
	}
	parts := strings.Fields(name)
	if len(parts) == 1 {
		return parts[0], ""
	}
	return parts[0], strings.Join(parts[1:], " ")
}

func trimStringPtr(v *string) *string {
	if v == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*v)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

// normalizeLanguageCode stores a short BCP 47 primary language tag or nil when unset/invalid.
func normalizeLanguageCode(code *string) *string {
	code = trimStringPtr(code)
	if code == nil {
		return nil
	}
	primary := strings.ToLower(strings.Split(*code, "-")[0])
	if primary == "" || len(primary) > 8 {
		return nil
	}
	for _, r := range primary {
		if r < 'a' || r > 'z' {
			return nil
		}
	}
	return &primary
}
