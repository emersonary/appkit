package resource

import (
	"encoding/json"
	"fmt"
	"net/mail"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

var emailPattern = regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`)

// ValidateValues checks submitted values against schema rules for editable fields.
func ValidateValues(schema Schema, values map[string]string) error {
	for _, field := range schema.Fields {
		if !field.Editable {
			continue
		}
		raw := values[field.Key]
		for _, rule := range field.Validations {
			if err := validateRule(field, raw, rule); err != nil {
				return err
			}
		}
		if field.Kind == KindCountry && strings.TrimSpace(raw) != "" && !isAllowedCountry(raw) {
			return fmt.Errorf("resource: field %q has invalid country code", field.Key)
		}
		if field.Kind == KindLocation && field.LocationMode == LocationModeManual {
			if err := validateLocationField(field, raw); err != nil {
				return err
			}
		}
	}
	return nil
}

func validateLocationField(field Field, raw string) error {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	var payload struct {
		Lat *float64 `json:"lat"`
		Lng *float64 `json:"lng"`
	}
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		return fmt.Errorf("resource: field %q has invalid location payload", field.Key)
	}
	if payload.Lat == nil || payload.Lng == nil {
		return fmt.Errorf("resource: field %q requires lat and lng", field.Key)
	}
	if *payload.Lat < -90 || *payload.Lat > 90 {
		return fmt.Errorf("resource: field %q latitude out of range", field.Key)
	}
	if *payload.Lng < -180 || *payload.Lng > 180 {
		return fmt.Errorf("resource: field %q longitude out of range", field.Key)
	}
	return nil
}

func validateRule(field Field, raw string, rule ValidationRule) error {
	trimmed := strings.TrimSpace(raw)
	switch rule.Kind {
	case ValidationRequired:
		if trimmed == "" {
			return fmt.Errorf("resource: required field %q is empty", field.Key)
		}
	case ValidationEmail:
		if trimmed == "" {
			return nil
		}
		if _, err := mail.ParseAddress(trimmed); err != nil && !emailPattern.MatchString(trimmed) {
			return fmt.Errorf("resource: field %q must be a valid email", field.Key)
		}
	case ValidationURL:
		if trimmed == "" {
			return nil
		}
		parsed, err := url.Parse(trimmed)
		if err != nil || parsed.Host == "" {
			return fmt.Errorf("resource: field %q must be a valid URL", field.Key)
		}
		scheme := strings.ToLower(parsed.Scheme)
		if scheme != "http" && scheme != "https" {
			return fmt.Errorf("resource: field %q must use http or https", field.Key)
		}
	case ValidationMaxLength:
		limit, err := strconv.Atoi(rule.Param)
		if err != nil || limit <= 0 {
			return fmt.Errorf("resource: invalid max length rule for field %q", field.Key)
		}
		if len(trimmed) > limit {
			return fmt.Errorf("resource: field %q exceeds max length of %d", field.Key, limit)
		}
	case ValidationPattern:
		if trimmed == "" || rule.Param == "" {
			return nil
		}
		re, err := regexp.Compile(rule.Param)
		if err != nil {
			return fmt.Errorf("resource: invalid pattern rule for field %q", field.Key)
		}
		if !re.MatchString(trimmed) {
			return fmt.Errorf("resource: field %q has invalid format", field.Key)
		}
	}
	return nil
}

func buildValidations(field Field, meta fieldMeta) []ValidationRule {
	rules := make([]ValidationRule, 0, 4)
	add := func(kind ValidationKind, param string) {
		for _, existing := range rules {
			if existing.Kind == kind && existing.Param == param {
				return
			}
		}
		rules = append(rules, ValidationRule{Kind: kind, Param: param})
	}

	if meta.required || field.Required {
		add(ValidationRequired, "")
	}
	for _, kind := range meta.validations {
		add(kind, meta.validationParams[kind])
	}
	switch field.Kind {
	case KindEmail:
		add(ValidationEmail, "")
	case KindURL:
		add(ValidationURL, "")
	}
	if meta.maxLength > 0 {
		add(ValidationMaxLength, strconv.Itoa(meta.maxLength))
	}
	return rules
}

func finalizeField(field *Field, meta fieldMeta) {
	field.Validations = buildValidations(*field, meta)
	if field.Kind == KindCountry {
		field.Options = CountryOptions()
	}
}
