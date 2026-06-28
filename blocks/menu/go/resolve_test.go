package menu

import (
	"testing"

	"github.com/emersonary/appkit/permissions"
)

func testPermissionSetup() permissions.Setup {
	return permissions.Setup{
		DefaultProfile: "member",
		Permissions: []permissions.PermissionConfig{
			{ID: "posts", Name: "Posts", Category: "blog", BeAction: permissions.AllActions, RouteName: "/admin/posts", Icon: "posts"},
			{ID: "posts_drafts", Name: "Drafts", Category: "blog", Parent: "posts", BeAction: permissions.ActionList, RouteName: "/admin/posts/drafts", Icon: "draft"},
			{ID: "posts_published", Name: "Published", Category: "blog", Parent: "posts", BeAction: permissions.ActionList, RouteName: "/admin/posts/published", Icon: "published"},
			{ID: "settings_general", Name: "General", Category: "settings", BeAction: permissions.ActionList, RouteName: "/admin/settings", Icon: "settings"},
		},
		Profiles: []permissions.ProfileConfig{
			{ID: "member", Name: "Member"},
			{ID: "admin", Name: "Admin", Superuser: true},
		},
	}
}

func testMenuSetup() Setup {
	return Setup{
		Sidebar: SidebarConfig{
			DefaultMenu: "posts_drafts",
		},
		Menus: []MenuConfig{
			{
				ID:   "manage",
				Name: "Manage",
				Permissions: []string{
					"posts",
				},
			},
		},
	}
}

func TestResolveLayoutExpandsDescendants(t *testing.T) {
	permSetup := testPermissionSetup()
	tree, err := permissions.PermissionConfigList(permSetup.Permissions).Tree()
	if err != nil {
		t.Fatal(err)
	}

	grants := []permissions.FlatPermission{
		{Permission: permissions.Permission{IDPermission: "posts", Name: "Posts"}, GrantedActions: permissions.ActionList},
		{Permission: permissions.Permission{IDPermission: "posts_drafts", Name: "Drafts"}, GrantedActions: permissions.ActionList},
	}

	layout, err := ResolveLayout(testMenuSetup(), tree, grants, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(layout.Menus) != 1 {
		t.Fatalf("expected 1 menu, got %d", len(layout.Menus))
	}
	if len(layout.Menus[0].Items) != 1 {
		t.Fatalf("expected 1 root item, got %d", len(layout.Menus[0].Items))
	}
	if len(layout.Menus[0].Items[0].Children) != 1 {
		t.Fatalf("expected expanded child, got %d children", len(layout.Menus[0].Items[0].Children))
	}
	if layout.DefaultSelectedPermissionID != "posts_drafts" {
		t.Fatalf("expected default posts_drafts, got %q", layout.DefaultSelectedPermissionID)
	}
}

func TestValidateAgainstPermissionsRejectsUnknownID(t *testing.T) {
	setup := testMenuSetup()
	setup.Menus[0].Permissions = []string{"missing_permission"}

	if err := setup.ValidateAgainstPermissions(testPermissionSetup()); err == nil {
		t.Fatal("expected validation error for unknown permission id")
	}
}
