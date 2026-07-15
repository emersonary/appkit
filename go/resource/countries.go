package resource

import "sort"

// CountryOptions returns ISO 3166-1 alpha-2 choices sorted by label.
func CountryOptions() []FieldOption {
	options := make([]FieldOption, 0, len(iso3166Alpha2))
	for code, label := range iso3166Alpha2 {
		options = append(options, FieldOption{Value: code, Label: label})
	}
	sort.Slice(options, func(i, j int) bool {
		return options[i].Label < options[j].Label
	})
	return options
}

func isAllowedCountry(code string) bool {
	_, ok := iso3166Alpha2[code]
	return ok
}
