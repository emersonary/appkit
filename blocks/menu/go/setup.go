package menu

import (
	"fmt"
	"os"
	"strings"

	"github.com/emersonary/appkit/permissions"
	"gopkg.in/yaml.v3"
)

type SidebarConfig struct {
	Floating         bool   `yaml:"floating" mapstructure:"floating" json:"floating"`
	HideWhenSelected bool   `yaml:"hide_when_selected" mapstructure:"hide_when_selected" json:"hide_when_selected"`
	Locked           bool   `yaml:"locked" mapstructure:"locked" json:"locked"`
	DefaultMenu      string `yaml:"default_menu" mapstructure:"default_menu" json:"default_menu"`
}

type MenuConfig struct {
	ID          string   `yaml:"id" mapstructure:"id" json:"id"`
	Name        string   `yaml:"name" mapstructure:"name" json:"name"`
	Icon        string   `yaml:"icon" mapstructure:"icon" json:"icon"`
	SortOrder   int      `yaml:"sort_order" mapstructure:"sort_order" json:"sort_order"`
	Permissions []string `yaml:"permissions" mapstructure:"permissions" json:"permissions"`
}

// Setup is the validated menu definition.
type Setup struct {
	Sidebar SidebarConfig `yaml:"sidebar" mapstructure:"sidebar" json:"sidebar"`
	Menus   []MenuConfig  `yaml:"menus" mapstructure:"menus" json:"menus"`
}

func LoadSetup(path string) (Setup, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Setup{}, wrapErr(ErrLoadSetup, "path", fmt.Errorf("%s: %w", path, err))
	}

	var setup Setup
	if err := yaml.Unmarshal(data, &setup); err != nil {
		return Setup{}, wrapErr(ErrLoadSetup, "parse", err)
	}

	setup.normalize()
	if err := setup.Validate(); err != nil {
		return Setup{}, err
	}
	return setup, nil
}

func (s *Setup) normalize() {
	for i := range s.Menus {
		s.Menus[i].ID = strings.TrimSpace(s.Menus[i].ID)
		s.Menus[i].Name = strings.TrimSpace(s.Menus[i].Name)
		s.Menus[i].Icon = strings.TrimSpace(s.Menus[i].Icon)
		clean := make([]string, 0, len(s.Menus[i].Permissions))
		for _, fullID := range s.Menus[i].Permissions {
			if trimmed := strings.TrimSpace(fullID); trimmed != "" {
				clean = append(clean, trimmed)
			}
		}
		s.Menus[i].Permissions = clean
	}
	s.Sidebar.DefaultMenu = strings.TrimSpace(s.Sidebar.DefaultMenu)
}

func (s Setup) Validate() error {
	if len(s.Menus) == 0 {
		return invalidSetup("menus", "at least one menu is required")
	}

	menuIDs := make(map[string]struct{}, len(s.Menus))
	for _, menu := range s.Menus {
		if menu.ID == "" {
			return invalidSetup("menus", "menu id is required")
		}
		if _, dup := menuIDs[menu.ID]; dup {
			return invalidSetupf("menus", "duplicate menu id %q", menu.ID)
		}
		menuIDs[menu.ID] = struct{}{}
		if menu.Name == "" {
			return invalidSetupf("menus.%s.name", menu.ID, "name is required")
		}
		if len(menu.Permissions) == 0 {
			return invalidSetupf("menus.%s.permissions", menu.ID, "at least one permission full_id is required")
		}
	}

	if s.Sidebar.DefaultMenu == "" {
		return invalidSetup("sidebar.default_menu", "default_menu permission id is required")
	}

	return nil
}

// ValidateAgainstPermissions checks menu full_ids and default_menu against the permissions catalog.
func (s Setup) ValidateAgainstPermissions(permSetup permissions.Setup) error {
	tree, err := permissions.PermissionConfigList(permSetup.Permissions).Tree()
	if err != nil {
		return invalidSetup("permissions", err.Error())
	}

	permissionIDs := make(map[string]struct{}, len(permSetup.Permissions))
	for _, p := range permSetup.Permissions {
		permissionIDs[strings.TrimSpace(p.ID)] = struct{}{}
	}

	for _, menu := range s.Menus {
		for _, fullID := range menu.Permissions {
			if tree.FindByFullID(fullID) == nil {
				return invalidSetupf("menus.%s.permissions", menu.ID, "unknown permission full_id %q", fullID)
			}
		}
	}

	if _, ok := permissionIDs[s.Sidebar.DefaultMenu]; !ok {
		return invalidSetupf("sidebar.default_menu", "unknown permission id %q", s.Sidebar.DefaultMenu)
	}

	return nil
}
