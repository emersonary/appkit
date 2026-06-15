package currency

import (
	"sort"
	"strings"
)

type CatalogEntry struct {
	Name   string
	Symbol string
}

func NormalizeCode(code string) string {
	return strings.ToUpper(strings.TrimSpace(code))
}

func LookupISO4217(code string) (CatalogEntry, bool) {
	entry, ok := ISO4217Catalog[NormalizeCode(code)]
	return entry, ok
}

func AllISO4217Codes() []string {
	codes := make([]string, 0, len(ISO4217Catalog))
	for code := range ISO4217Catalog {
		codes = append(codes, code)
	}

	sort.Strings(codes)
	return codes
}

func validateISO4217Code(code string) error {
	normalized := NormalizeCode(code)
	if len(normalized) != 3 {
		return ErrUnknownISO4217.With("code", code)
	}

	if _, ok := ISO4217Catalog[normalized]; !ok {
		return ErrUnknownISO4217.With("code", normalized)
	}

	return nil
}
