package language

import (
	"sort"
	"strings"
)

type CatalogEntry struct {
	Name       string
	NativeName string
	Direction  string
}

func NormalizeCode(code string) string {
	code = strings.ToLower(strings.TrimSpace(code))
	if code == "" {
		return ""
	}

	// Regional tags collapse to the primary ISO 639-1 subtag.
	if i := strings.IndexByte(code, '-'); i > 0 {
		code = code[:i]
	}

	return code
}

func LookupCatalog(code string) (CatalogEntry, bool) {
	entry, ok := LanguageCatalog[NormalizeCode(code)]
	return entry, ok
}

func AllCatalogCodes() []string {
	codes := make([]string, 0, len(LanguageCatalog))
	for code := range LanguageCatalog {
		codes = append(codes, code)
	}

	sort.Strings(codes)
	return codes
}

func validateCatalogCode(code string) error {
	normalized := NormalizeCode(code)
	if len(normalized) != 2 {
		return ErrUnknownCatalogCode.With("code", code)
	}

	if _, ok := LanguageCatalog[normalized]; !ok {
		return ErrUnknownCatalogCode.With("code", normalized)
	}

	return nil
}

func normalizeCodeList(codes []string) []string {
	if len(codes) == 0 {
		return nil
	}

	out := make([]string, 0, len(codes))
	for _, code := range codes {
		out = append(out, NormalizeCode(code))
	}

	sort.Strings(out)
	return out
}
