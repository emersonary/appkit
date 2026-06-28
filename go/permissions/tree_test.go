package permissions

import (
	"strings"
	"testing"
)

func TestPermissionConfigListTree(t *testing.T) {
	tree, err := PermissionConfigList{
		{ID: "blog_post", Name: "Posts", Category: "content", BeAction: AllActions, SortOrder: 1},
		{ID: "blog_post_drafts", Name: "Drafts", Category: "content", Parent: "blog_post", BeAction: ActionList, SortOrder: 2},
	}.Tree()
	if err != nil {
		t.Fatalf("Tree: %v", err)
	}

	posts := tree.FindByFullID("Posts")
	if posts == nil {
		t.Fatal("expected Posts node")
	}
	if posts.Permission.IDPermission != "blog_post" {
		t.Fatalf("posts id: got %q", posts.Permission.IDPermission)
	}
	if len(posts.Children) != 1 {
		t.Fatalf("expected 1 child, got %d", len(posts.Children))
	}

	drafts := tree.FindByFullID("Posts.Drafts")
	if drafts == nil {
		t.Fatal("expected Posts.Drafts node")
	}
	if drafts.ParentNode != posts {
		t.Fatal("expected drafts parent to be posts")
	}
	if drafts.FullID != "Posts.Drafts" {
		t.Fatalf("full id: got %q", drafts.FullID)
	}

	if byID := tree.FindByID("blog_post_drafts"); byID != drafts {
		t.Fatalf("FindByID: expected drafts node, got %+v", byID)
	}
	if tree.FindByID("missing") != nil {
		t.Fatal("expected nil for unknown permission id")
	}
}

func TestPermissionTreeFindByFullIDMissing(t *testing.T) {
	tree, err := PermissionConfigList{
		{ID: "trip", Name: "Trips", BeAction: AllActions},
	}.Tree()
	if err != nil {
		t.Fatalf("Tree: %v", err)
	}
	if got := tree.FindByFullID("Trips.Nope"); got != nil {
		t.Fatalf("expected nil, got %+v", got)
	}
}

func TestPermissionTreeDuplicateFullID(t *testing.T) {
	_, err := PermissionConfigList{
		{ID: "a", Name: "Posts", BeAction: AllActions},
		{ID: "b", Name: "Posts", BeAction: AllActions},
	}.Tree()
	if err == nil {
		t.Fatal("expected duplicate full id error")
	}
	if !strings.Contains(err.Error(), "duplicate permission full id") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPermissionTreeUnknownParent(t *testing.T) {
	_, err := PermissionConfigList{
		{ID: "child", Name: "Child", Parent: "missing", BeAction: AllActions},
	}.Tree()
	if err == nil {
		t.Fatal("expected unknown parent error")
	}
}

func TestPermissionTreeCycle(t *testing.T) {
	_, err := NewPermissionTree([]Permission{
		{IDPermission: "a", Name: "A", IDParent: strPtr("b")},
		{IDPermission: "b", Name: "B", IDParent: strPtr("a")},
	})
	if err == nil {
		t.Fatal("expected cycle error")
	}
}

func strPtr(s string) *string { return &s }
