package menu

import (
	"os"
	"strings"
)

// AppConfig is the application-level menu block config (main YAML menu node).
type AppConfig struct {
	Enabled    bool           `mapstructure:"enabled" json:"enabled"`
	ConfigPath string         `mapstructure:"config_path" json:"config_path,omitempty"`
	Sidebar    SidebarConfig  `mapstructure:"sidebar" json:"sidebar"`
	Menus      []MenuConfig   `mapstructure:"menus" json:"menus"`
}

// ResolveSetup loads an external file when config_path is set, otherwise resolves inline input.
func ResolveSetup(input AppConfig) (Setup, error) {
	path := strings.TrimSpace(input.ConfigPath)
	if path != "" {
		if _, err := os.Stat(path); err == nil {
			setup, err := LoadSetup(path)
			if err != nil {
				return Setup{}, err
			}
			mergeAppConfig(&setup, input)
			setup.normalize()
			if err := setup.Validate(); err != nil {
				return Setup{}, err
			}
			return setup, nil
		}
	}

	setup := Setup{
		Sidebar: input.Sidebar,
		Menus:   append([]MenuConfig(nil), input.Menus...),
	}
	setup.normalize()
	if err := setup.Validate(); err != nil {
		return Setup{}, err
	}
	return setup, nil
}

func mergeAppConfig(setup *Setup, input AppConfig) {
	setup.Sidebar = input.Sidebar
	if len(input.Menus) > 0 {
		setup.Menus = append([]MenuConfig(nil), input.Menus...)
	}
}
