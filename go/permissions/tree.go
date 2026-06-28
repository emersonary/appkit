package permissions

import (
	"fmt"
	"sort"
	"strings"
)

// PermissionNode is one node in a permission tree. FullID is the dotted path of
// ancestor and own display names (e.g. "Posts.Drafts").
type PermissionNode struct {
	FullID     string
	Name       string
	ParentNode *PermissionNode
	Permission *Permission
	Children   []*PermissionNode
}

// PermissionTree is a forest of permission nodes indexed by FullID and config id.
type PermissionTree struct {
	Roots    []*PermissionNode
	byFullID map[string]*PermissionNode
	byID     map[string]*PermissionNode
}

// FindByFullID returns the node for a dotted full id path, or nil when not found.
func (t *PermissionTree) FindByFullID(fullID string) *PermissionNode {
	if t == nil {
		return nil
	}
	return t.byFullID[strings.TrimSpace(fullID)]
}

// FindByID returns the node for a permission config id, or nil when not found.
func (t *PermissionTree) FindByID(id string) *PermissionNode {
	if t == nil {
		return nil
	}
	return t.byID[strings.TrimSpace(id)]
}

// PermissionConfigList is a []PermissionConfig with tree-building helpers.
type PermissionConfigList []PermissionConfig

// Tree builds a PermissionTree from config rows (parent links + dotted FullID from names).
func (list PermissionConfigList) Tree() (*PermissionTree, error) {
	return NewPermissionTreeFromConfigs([]PermissionConfig(list))
}

// NewPermissionTreeFromConfigs builds a tree from YAML/config permission rows.
func NewPermissionTreeFromConfigs(configs []PermissionConfig) (*PermissionTree, error) {
	perms := make([]Permission, 0, len(configs))
	for _, cfg := range configs {
		perms = append(perms, permissionFromConfig(cfg))
	}
	return NewPermissionTree(perms)
}

// NewPermissionTree builds a tree from persisted permission rows.
func NewPermissionTree(perms []Permission) (*PermissionTree, error) {
	if len(perms) == 0 {
		return &PermissionTree{
			byFullID: make(map[string]*PermissionNode),
			byID:     make(map[string]*PermissionNode),
		}, nil
	}

	byID := make(map[string]*PermissionNode, len(perms))
	for i := range perms {
		p := perms[i]
		id := strings.TrimSpace(p.IDPermission)
		if id == "" {
			return nil, fmt.Errorf("permission id required")
		}
		if _, dup := byID[id]; dup {
			return nil, fmt.Errorf("duplicate permission id %q", id)
		}
		name := strings.TrimSpace(p.Name)
		if name == "" {
			name = id
		}
		perm := perms[i]
		byID[id] = &PermissionNode{
			Name:       name,
			Permission: &perm,
		}
	}

	roots := make([]*PermissionNode, 0)
	for i := range perms {
		p := perms[i]
		id := strings.TrimSpace(p.IDPermission)
		node := byID[id]

		parentID := ""
		if p.IDParent != nil {
			parentID = strings.TrimSpace(*p.IDParent)
		}
		if parentID == "" {
			roots = append(roots, node)
			continue
		}
		if parentID == id {
			return nil, fmt.Errorf("permission %q cannot be its own parent", id)
		}
		parent, ok := byID[parentID]
		if !ok {
			return nil, fmt.Errorf("unknown parent %q for permission %q", parentID, id)
		}
		if pathHasID(parent, id) {
			return nil, fmt.Errorf("permission cycle involving %q", id)
		}
		node.ParentNode = parent
		parent.Children = append(parent.Children, node)
	}

	for _, root := range roots {
		sortPermissionChildren(root)
	}
	sort.Slice(roots, func(i, j int) bool {
		return permissionSortOrder(roots[i]) < permissionSortOrder(roots[j])
	})

	tree := &PermissionTree{
		Roots:    roots,
		byFullID: make(map[string]*PermissionNode, len(perms)),
		byID:     byID,
	}
	for _, root := range roots {
		if err := assignPermissionFullIDs(root, tree.byFullID); err != nil {
			return nil, err
		}
	}
	return tree, nil
}

func permissionFromConfig(cfg PermissionConfig) Permission {
	id := strings.TrimSpace(cfg.ID)
	parent := strings.TrimSpace(cfg.Parent)
	var idParent *string
	if parent != "" {
		idParent = &parent
	}
	enabled := true
	if cfg.Enabled != nil {
		enabled = *cfg.Enabled
	}
	return Permission{
		IDPermission:         id,
		Name:                 cfg.Name,
		IDPermissionCategory: strings.TrimSpace(cfg.Category),
		IDParent:             idParent,
		BeAction:             cfg.BeAction,
		RouteName:            cfg.RouteName,
		Icon:                 cfg.Icon,
		Enabled:              enabled,
		SortOrder:            cfg.SortOrder,
	}
}

func assignPermissionFullIDs(node *PermissionNode, index map[string]*PermissionNode) error {
	node.FullID = permissionNodeFullID(node)
	if node.FullID == "" {
		return fmt.Errorf("permission full id required for %q", node.Permission.IDPermission)
	}
	if _, dup := index[node.FullID]; dup {
		return fmt.Errorf("duplicate permission full id %q", node.FullID)
	}
	index[node.FullID] = node

	for _, child := range node.Children {
		if err := assignPermissionFullIDs(child, index); err != nil {
			return err
		}
	}
	return nil
}

func permissionNodeFullID(node *PermissionNode) string {
	segments := make([]string, 0, 4)
	for n := node; n != nil; n = n.ParentNode {
		seg := strings.TrimSpace(n.Name)
		if seg == "" && n.Permission != nil {
			seg = strings.TrimSpace(n.Permission.IDPermission)
		}
		segments = append([]string{seg}, segments...)
	}
	return strings.Join(segments, ".")
}

func pathHasID(node *PermissionNode, id string) bool {
	for n := node; n != nil; n = n.ParentNode {
		if n.Permission != nil && strings.TrimSpace(n.Permission.IDPermission) == id {
			return true
		}
	}
	return false
}

func sortPermissionChildren(node *PermissionNode) {
	sort.Slice(node.Children, func(i, j int) bool {
		return permissionSortOrder(node.Children[i]) < permissionSortOrder(node.Children[j])
	})
	for _, child := range node.Children {
		sortPermissionChildren(child)
	}
}

func permissionSortOrder(node *PermissionNode) int {
	if node == nil || node.Permission == nil {
		return 0
	}
	return node.Permission.SortOrder
}
