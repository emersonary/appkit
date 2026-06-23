package menu

import (
	"sort"
	"strings"

	"github.com/emersonary/appkit/permissions"
)

// ResolvedMenuItem is one visible navigation node.
type ResolvedMenuItem struct {
	FullID       string
	PermissionID string
	Name         string
	RouteName    string
	Icon         string
	Children     []ResolvedMenuItem
}

// ResolvedMenu is a named menu with filtered items for an account.
type ResolvedMenu struct {
	ID    string
	Name  string
	Items []ResolvedMenuItem
}

// Layout is the full resolved navigation for an account.
type Layout struct {
	Sidebar                      SidebarConfig
	Menus                        []ResolvedMenu
	DefaultSelectedPermissionID  string
}

func ResolveLayout(setup Setup, tree *permissions.PermissionTree, grants []permissions.FlatPermission) (Layout, error) {
	grantMap := buildGrantMap(grants)
	menus := make([]ResolvedMenu, 0, len(setup.Menus))
	sortedMenus := append([]MenuConfig(nil), setup.Menus...)
	sort.Slice(sortedMenus, func(i, j int) bool {
		return sortedMenus[i].SortOrder < sortedMenus[j].SortOrder
	})

	for _, menuCfg := range sortedMenus {
		roots, err := menuRoots(menuCfg, tree)
		if err != nil {
			return Layout{}, err
		}
		items := make([]ResolvedMenuItem, 0, len(roots))
		for _, root := range roots {
			if item := filterMenuNode(root, grantMap); item != nil {
				items = append(items, *item)
			}
		}
		menus = append(menus, ResolvedMenu{
			ID:    menuCfg.ID,
			Name:  menuCfg.Name,
			Items: items,
		})
	}

	return Layout{
		Sidebar:                     setup.Sidebar,
		Menus:                       menus,
		DefaultSelectedPermissionID: findDefaultPermissionID(setup.Sidebar.DefaultMenu, menus),
	}, nil
}

func buildGrantMap(grants []permissions.FlatPermission) map[string]int {
	out := make(map[string]int, len(grants))
	for _, grant := range grants {
		out[grant.IDPermission] = grant.GrantedActions
	}
	return out
}

func menuRoots(menu MenuConfig, tree *permissions.PermissionTree) ([]*permissions.PermissionNode, error) {
	included := make(map[string]*permissions.PermissionNode)
	for _, fullID := range menu.Permissions {
		node := tree.FindByFullID(fullID)
		if node == nil {
			return nil, invalidSetupf("menus.%s.permissions", menu.ID, "unknown permission full_id %q", fullID)
		}
		collectSubtree(node, included)
	}

	roots := make([]*permissions.PermissionNode, 0)
	for _, node := range included {
		parentMissing := node.ParentNode == nil
		if !parentMissing {
			_, parentIncluded := included[node.ParentNode.FullID]
			parentMissing = !parentIncluded
		}
		if parentMissing {
			roots = append(roots, node)
		}
	}

	sort.Slice(roots, func(i, j int) bool {
		return permissionSortOrder(roots[i]) < permissionSortOrder(roots[j])
	})
	return roots, nil
}

func collectSubtree(node *permissions.PermissionNode, included map[string]*permissions.PermissionNode) {
	if node == nil {
		return
	}
	included[node.FullID] = node
	for _, child := range node.Children {
		collectSubtree(child, included)
	}
}

func filterMenuNode(node *permissions.PermissionNode, grants map[string]int) *ResolvedMenuItem {
	if node == nil || node.Permission == nil {
		return nil
	}

	children := make([]ResolvedMenuItem, 0, len(node.Children))
	for _, child := range node.Children {
		if item := filterMenuNode(child, grants); item != nil {
			children = append(children, *item)
		}
	}

	permID := strings.TrimSpace(node.Permission.IDPermission)
	selfVisible := isMenuVisible(permID, grants)
	if !selfVisible && len(children) == 0 {
		return nil
	}

	return &ResolvedMenuItem{
		FullID:       node.FullID,
		PermissionID: permID,
		Name:         node.Name,
		RouteName:    node.Permission.RouteName,
		Icon:         node.Permission.Icon,
		Children:     children,
	}
}

func isMenuVisible(permissionID string, grants map[string]int) bool {
	mask, ok := grants[permissionID]
	return ok && permissions.MaskAllows(mask, permissions.ActionList)
}

func findDefaultPermissionID(defaultID string, menus []ResolvedMenu) string {
	defaultID = strings.TrimSpace(defaultID)
	if defaultID == "" {
		return ""
	}
	for _, menu := range menus {
		if id := findPermissionIDInItems(defaultID, menu.Items); id != "" {
			return id
		}
	}
	return ""
}

func findPermissionIDInItems(target string, items []ResolvedMenuItem) string {
	for _, item := range items {
		if item.PermissionID == target {
			return item.PermissionID
		}
		if id := findPermissionIDInItems(target, item.Children); id != "" {
			return id
		}
	}
	return ""
}

func permissionSortOrder(node *permissions.PermissionNode) int {
	if node == nil || node.Permission == nil {
		return 0
	}
	return node.Permission.SortOrder
}
