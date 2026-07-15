package resource

import "testing"

func TestValidateEmailAndURL(t *testing.T) {
	schema := Schema{Fields: []Field{
		{Key: "email", Editable: true, Validations: []ValidationRule{{Kind: ValidationEmail}}},
		{Key: "website_url", Editable: true, Validations: []ValidationRule{{Kind: ValidationURL}}},
	}}

	if err := ValidateValues(schema, map[string]string{
		"email":       "not-an-email",
		"website_url": "ftp://example.com",
	}); err == nil {
		t.Fatal("expected validation error")
	}

	if err := ValidateValues(schema, map[string]string{
		"email":       "hello@example.com",
		"website_url": "https://example.com",
	}); err != nil {
		t.Fatalf("unexpected validation error: %v", err)
	}
}

func TestCountryFieldIncludesOptions(t *testing.T) {
	schema, err := Describe(struct {
		Country *string `db:"country_code" resource:"label=Country;kind=country"`
	}{})
	if err != nil {
		t.Fatalf("Describe: %v", err)
	}
	if len(schema.Fields) != 1 || len(schema.Fields[0].Options) == 0 {
		t.Fatalf("expected country options, got %+v", schema.Fields)
	}
}

func TestValidateCountryCode(t *testing.T) {
	schema, err := Describe(struct {
		Country *string `db:"country_code" resource:"kind=country"`
	}{})
	if err != nil {
		t.Fatalf("Describe: %v", err)
	}
	if err := ValidateValues(schema, map[string]string{"country_code": "ZZ"}); err == nil {
		t.Fatal("expected invalid country")
	}
	if err := ValidateValues(schema, map[string]string{"country_code": "PT"}); err != nil {
		t.Fatalf("PT should be valid: %v", err)
	}
}
