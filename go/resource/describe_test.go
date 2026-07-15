package resource

import (
	"testing"
	"time"
)

type sampleProfile struct {
	TenantID    string     `db:"tenant_id" resource:"-"`
	DisplayName *string    `db:"display_name" resource:"label=Display name;kind=text;section=identity;order=1;required"`
	Email       *string    `db:"email" resource:"label=Email;kind=email;section=contact;order=1"`
	UpdatedAt   time.Time  `db:"updated_at" resource:"readonly;kind=datetime;section=meta;order=99"`
	HiddenField string     `db:"hidden_field" resource:"hidden"`
}

func TestDescribeBusinessLikeStruct(t *testing.T) {
	schema, err := Describe(sampleProfile{}, WithName("tenant_business_profile"), WithLabel("Business profile"))
	if err != nil {
		t.Fatalf("Describe: %v", err)
	}
	if schema.Name != "tenant_business_profile" {
		t.Fatalf("name = %q", schema.Name)
	}
	if len(schema.Fields) != 2 {
		t.Fatalf("expected 2 visible fields, got %d", len(schema.Fields))
	}
	var displayName *Field
	for i := range schema.Fields {
		if schema.Fields[i].Key == "display_name" {
			displayName = &schema.Fields[i]
			break
		}
	}
	if displayName == nil || !displayName.Required {
		t.Fatalf("display_name field missing or not required: %+v", displayName)
	}
}

func TestBuildEditPayloadAndApplyPatch(t *testing.T) {
	name := "Maria's Salon"
	email := "maria@example.com"
	updated := time.Date(2026, 7, 5, 12, 0, 0, 0, time.UTC)
	profile := sampleProfile{
		TenantID:    "tenant-1",
		DisplayName: &name,
		Email:       &email,
		UpdatedAt:   updated,
	}

	payload, err := BuildEditPayload(profile, WithName("tenant_business_profile"), WithLabel("Business profile"))
	if err != nil {
		t.Fatalf("BuildEditPayload: %v", err)
	}
	if payload.Values["display_name"] != name {
		t.Fatalf("display_name = %q", payload.Values["display_name"])
	}
	if _, ok := payload.Values["updated_at"]; ok {
		t.Fatal("updated_at should not appear in edit payload values")
	}

	newName := "Maria's Studio"
	if err := ApplyPatch(&profile, map[string]string{
		"display_name": newName,
		"email":        email,
	}); err != nil {
		t.Fatalf("ApplyPatch: %v", err)
	}
	if profile.DisplayName == nil || *profile.DisplayName != newName {
		t.Fatalf("patched display name = %+v", profile.DisplayName)
	}
}

func TestApplyPatchRequiredValidation(t *testing.T) {
	profile := sampleProfile{}
	err := ApplyPatch(&profile, map[string]string{"email": "a@b.com"})
	if err == nil {
		t.Fatal("expected required display_name error")
	}
}

func TestBuildNewEditPayloadMarksRecordNew(t *testing.T) {
	payload, err := BuildNewEditPayload(sampleProfile{}, WithName("tenant_social_link"), WithLabel("Social link"))
	if err != nil {
		t.Fatalf("BuildNewEditPayload: %v", err)
	}
	if payload.RecordState != RecordStateNew {
		t.Fatalf("record state = %q", payload.RecordState)
	}
	if payload.Values["display_name"] != "" {
		t.Fatalf("expected empty display_name, got %q", payload.Values["display_name"])
	}
}
