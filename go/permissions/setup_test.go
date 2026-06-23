package permissions

import (
	"strings"
	"testing"
)

func TestSetupDefaultsProfiles(t *testing.T) {
	setup := Setup{Schema: "account"}
	setup.normalize()
	if len(setup.Profiles) != 2 {
		t.Fatalf("expected default profiles, got %d", len(setup.Profiles))
	}
	if err := setup.Validate(); err != nil {
		t.Fatalf("Validate: %v", err)
	}
}

func TestSetupInputResolve(t *testing.T) {
	setup, err := (SetupInput{
		Schema:         "account",
		DefaultProfile: "member",
		Groups: []GroupConfig{
			{ID: "manage", Name: "Manage", RoutePrefix: "/admin/manage"},
		},
		Categories: []CategoryConfig{
			{ID: "content", Name: "Content", Group: "manage"},
		},
		Permissions: []PermissionConfig{
			{ID: "blog_post", Name: "Posts", Category: "content", BeAction: AllActions},
		},
		Profiles: []ProfileConfig{
			{ID: "member", Name: "Member"},
			{ID: "admin", Name: "Admin", Superuser: true},
		},
	}).Resolve()
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if setup.DefaultProfile != "member" {
		t.Fatalf("default profile: got %q", setup.DefaultProfile)
	}
}

func TestBuildSchemaSQLIncludesTables(t *testing.T) {
	setup := Setup{
		Schema:         "account",
		DefaultProfile: "member",
		Groups:         []GroupConfig{{ID: "manage", Name: "Manage"}},
		Categories:     []CategoryConfig{{ID: "content", Name: "Content", Group: "manage"}},
		Permissions:    []PermissionConfig{{ID: "blog_post", Name: "Posts", Category: "content", BeAction: AllActions}},
		Profiles:       []ProfileConfig{{ID: "member", Name: "Member"}},
	}
	setup.normalize()

	sqlText := buildSchemaSQL(setup)
	for _, fragment := range []string{
		`"account"."permission_groups"`,
		`"account"."permission_categories"`,
		`"account"."permissions"`,
		`"account"."profiles"`,
		`"account"."profile_permissions"`,
		`be_action`,
		`id_profile`,
		`'blog_post'`,
	} {
		if !strings.Contains(sqlText, fragment) {
			t.Fatalf("expected schema SQL to contain %q", fragment)
		}
	}
}

func TestPermissionConfigRejectsDotInID(t *testing.T) {
	setup := Setup{
		Schema:         "account",
		DefaultProfile: "member",
		Groups:         []GroupConfig{{ID: "manage", Name: "Manage"}},
		Categories:     []CategoryConfig{{ID: "content", Name: "Content", Group: "manage"}},
		Permissions:    []PermissionConfig{{ID: "blog.post", Name: "Posts", Category: "content", BeAction: AllActions}},
		Profiles:       []ProfileConfig{{ID: "member", Name: "Member"}},
	}
	setup.normalize()

	err := setup.Validate()
	if err == nil {
		t.Fatal("expected validation error for dotted permission id")
	}
	if !strings.Contains(err.Error(), "must not contain '.'") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPermissionConfigValidate(t *testing.T) {
	if err := (PermissionConfig{ID: "blog_post", BeAction: 1}).Validate(); err != nil {
		t.Fatalf("valid id: %v", err)
	}
	if err := (PermissionConfig{ID: "blog.post", BeAction: 1}).Validate(); err == nil {
		t.Fatal("expected error for dotted id")
	}
	if err := (PermissionConfig{ID: "blog_post", Parent: "blog.post", BeAction: 1}).Validate(); err == nil {
		t.Fatal("expected error for dotted parent")
	}
}

func TestMaskAllows(t *testing.T) {
	if !MaskAllows(5, ActionList) {
		t.Fatal("expected list in mask 5")
	}
	if MaskAllows(5, ActionCreate) {
		t.Fatal("did not expect create in mask 5")
	}
}

func TestWire_Disabled(t *testing.T) {
	svc, err := Wire(t.Context(), nil, SetupInput{Enabled: false}, WireOptions{})
	if err != nil {
		t.Fatalf("Wire: %v", err)
	}
	if svc != nil {
		t.Fatal("expected nil service when disabled")
	}
}
