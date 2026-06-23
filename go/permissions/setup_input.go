package permissions

import (
	"os"
	"strings"
)

// SetupInput is how the application declares permissions (main YAML permissions node).
type SetupInput struct {
	Enabled        bool                      `mapstructure:"enabled" json:"enabled"`
	ConfigPath     string                    `mapstructure:"config_path" json:"config_path,omitempty"`
	Schema         string                    `mapstructure:"schema" json:"schema"`
	AccountsSchema string                    `mapstructure:"accounts_schema" json:"accounts_schema"`
	DefaultProfile string                    `mapstructure:"default_profile" json:"default_profile"`
	SkipSeed       bool                      `mapstructure:"skip_seed" json:"skip_seed"`
	Groups         []GroupConfig             `mapstructure:"groups" json:"groups"`
	Categories     []CategoryConfig          `mapstructure:"categories" json:"categories"`
	Permissions    []PermissionConfig        `mapstructure:"permissions" json:"permissions"`
	Profiles       []ProfileConfig           `mapstructure:"profiles" json:"profiles"`
	ProfilePerms   []ProfilePermissionConfig `mapstructure:"profile_permissions" json:"profile_permissions"`
}

func (in *SetupInput) ApplyDefaults() {
	if strings.TrimSpace(in.Schema) == "" {
		in.Schema = "account"
	}
	if strings.TrimSpace(in.DefaultProfile) == "" {
		in.DefaultProfile = defaultProfileID
	}
}

// Resolve returns a validated Setup from inline declaration.
func (in SetupInput) Resolve() (Setup, error) {
	in.ApplyDefaults()

	setup := Setup{
		Schema:         in.Schema,
		AccountsSchema: in.AccountsSchema,
		DefaultProfile: in.DefaultProfile,
		SkipSeed:       in.SkipSeed,
		Groups:         append([]GroupConfig(nil), in.Groups...),
		Categories:     append([]CategoryConfig(nil), in.Categories...),
		Permissions:    append([]PermissionConfig(nil), in.Permissions...),
		Profiles:       append([]ProfileConfig(nil), in.Profiles...),
		ProfilePerms:   append([]ProfilePermissionConfig(nil), in.ProfilePerms...),
	}

	setup.normalize()
	if err := setup.Validate(); err != nil {
		return setup, err
	}

	return setup, nil
}

// ResolveSetup loads an external setup file when config_path is set, otherwise resolves inline input.
func ResolveSetup(input SetupInput) (Setup, error) {
	input.ApplyDefaults()

	path := strings.TrimSpace(input.ConfigPath)
	if path != "" {
		if _, err := os.Stat(path); err == nil {
			setup, err := LoadSetup(path)
			if err != nil {
				return Setup{}, err
			}
			mergeSetupInput(&setup, input)
			return setup, nil
		}
	}

	return input.Resolve()
}

func mergeSetupInput(setup *Setup, input SetupInput) {
	if input.Schema != "" {
		setup.Schema = input.Schema
	}
	if input.AccountsSchema != "" {
		setup.AccountsSchema = input.AccountsSchema
	}
	if input.DefaultProfile != "" {
		setup.DefaultProfile = input.DefaultProfile
	}
	setup.SkipSeed = input.SkipSeed || setup.SkipSeed
	if len(input.Groups) > 0 {
		setup.Groups = append([]GroupConfig(nil), input.Groups...)
	}
	if len(input.Categories) > 0 {
		setup.Categories = append([]CategoryConfig(nil), input.Categories...)
	}
	if len(input.Permissions) > 0 {
		setup.Permissions = append([]PermissionConfig(nil), input.Permissions...)
	}
	if len(input.Profiles) > 0 {
		setup.Profiles = append([]ProfileConfig(nil), input.Profiles...)
	}
	if len(input.ProfilePerms) > 0 {
		setup.ProfilePerms = append([]ProfilePermissionConfig(nil), input.ProfilePerms...)
	}
}
