package permissions

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

const defaultProfileID = "member"

type GroupConfig struct {
	ID          string `yaml:"id"`
	Name        string `yaml:"name"`
	SortOrder   int    `yaml:"sort_order"`
	RoutePrefix string `yaml:"route_prefix"`
}

type CategoryConfig struct {
	ID        string `yaml:"id"`
	Name      string `yaml:"name"`
	Group     string `yaml:"group"`
	SortOrder int    `yaml:"sort_order"`
}

type PermissionConfig struct {
	// ID must be a lowercase identifier without '.' (tree paths are built from parent links).
	ID        string `yaml:"id"`
	Name      string `yaml:"name"`
	Category  string `yaml:"category"`
	Parent    string `yaml:"parent"`
	BeAction  int    `yaml:"be_action"`
	RouteName string `yaml:"route_name"`
	Icon      string `yaml:"icon"`
	Enabled   *bool  `yaml:"enabled"`
	SortOrder int    `yaml:"sort_order"`
}

func (p PermissionConfig) Validate() error {
	id := strings.TrimSpace(p.ID)
	if id == "" {
		return fmt.Errorf("id required")
	}
	if err := validatePermissionID(id); err != nil {
		return err
	}
	parent := strings.TrimSpace(p.Parent)
	if parent != "" {
		if err := validatePermissionID(parent); err != nil {
			return fmt.Errorf("parent: %w", err)
		}
		if parent == id {
			return fmt.Errorf("permission %q cannot be its own parent", id)
		}
	}
	if p.BeAction < 0 {
		return fmt.Errorf("invalid be_action for %q", id)
	}
	return nil
}

type ProfileConfig struct {
	ID        string `yaml:"id"`
	Name      string `yaml:"name"`
	Superuser bool   `yaml:"superuser"`
	All       bool   `yaml:"all"` // alias for superuser
}

type ProfilePermissionConfig struct {
	Profile        string `yaml:"profile"`
	Permission     string `yaml:"permission"`
	GrantedActions *int   `yaml:"granted_actions"`
}

func (pp ProfilePermissionConfig) Validate() error {
	perm := strings.TrimSpace(pp.Permission)
	if perm == "" {
		return fmt.Errorf("permission required")
	}
	return validatePermissionID(perm)
}

// Setup is the validated RBAC permission system definition.
type Setup struct {
	Schema         string                    `yaml:"schema"`
	AccountsSchema string                    `yaml:"accounts_schema"`
	DefaultProfile string                    `yaml:"default_profile"`
	SkipSeed       bool                      `yaml:"skip_seed"`
	Groups         []GroupConfig             `yaml:"groups"`
	Categories     []CategoryConfig          `yaml:"categories"`
	Permissions    []PermissionConfig        `yaml:"permissions"`
	Profiles       []ProfileConfig           `yaml:"profiles"`
	ProfilePerms   []ProfilePermissionConfig `yaml:"profile_permissions"`
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
	s.Schema = strings.TrimSpace(s.Schema)
	if s.Schema == "" {
		s.Schema = "account"
	}
	s.AccountsSchema = strings.TrimSpace(s.AccountsSchema)
	if s.AccountsSchema == "" {
		s.AccountsSchema = s.Schema
	}
	s.DefaultProfile = strings.TrimSpace(s.DefaultProfile)
	if s.DefaultProfile == "" {
		s.DefaultProfile = defaultProfileID
	}

	for i := range s.Profiles {
		if s.Profiles[i].Superuser || s.Profiles[i].All {
			s.Profiles[i].Superuser = true
		}
	}

	if len(s.Profiles) == 0 {
		s.Profiles = []ProfileConfig{
			{ID: defaultProfileID, Name: "Member"},
			{ID: "admin", Name: "Administrator", Superuser: true},
		}
	}
}

func (s Setup) Validate() error {
	if err := validateIdent(s.Schema); err != nil {
		return ErrInvalidSchema.With("schema", err.Error())
	}
	if err := validateIdent(s.AccountsSchema); err != nil {
		return ErrInvalidSchema.With("accounts_schema", err.Error())
	}
	if s.DefaultProfile == "" {
		return ErrDefaultProfileRequired
	}

	groupIDs := make(map[string]struct{}, len(s.Groups))
	for _, g := range s.Groups {
		id := strings.TrimSpace(g.ID)
		if id == "" {
			return ErrInvalidSchema.With("group", "id required")
		}
		if _, dup := groupIDs[id]; dup {
			return ErrInvalidSchema.With("group", fmt.Sprintf("duplicate id %q", id))
		}
		groupIDs[id] = struct{}{}
	}

	categoryIDs := make(map[string]struct{}, len(s.Categories))
	for _, cat := range s.Categories {
		id := strings.TrimSpace(cat.ID)
		if id == "" {
			return ErrInvalidSchema.With("category", "id required")
		}
		if _, dup := categoryIDs[id]; dup {
			return ErrInvalidSchema.With("category", fmt.Sprintf("duplicate id %q", id))
		}
		if _, ok := groupIDs[strings.TrimSpace(cat.Group)]; !ok && len(s.Groups) > 0 {
			return ErrInvalidSchema.With("category", fmt.Sprintf("unknown group %q for category %q", cat.Group, id))
		}
		categoryIDs[id] = struct{}{}
	}

	permissionIDs := make(map[string]struct{}, len(s.Permissions))
	for _, p := range s.Permissions {
		id := strings.TrimSpace(p.ID)
		if err := p.Validate(); err != nil {
			return ErrInvalidSchema.With("permission", err.Error())
		}
		if _, dup := permissionIDs[id]; dup {
			return ErrInvalidSchema.With("permission", fmt.Sprintf("duplicate id %q", id))
		}
		if _, ok := categoryIDs[strings.TrimSpace(p.Category)]; !ok && len(s.Categories) > 0 {
			return ErrInvalidSchema.With("permission", fmt.Sprintf("unknown category %q for permission %q", p.Category, id))
		}
		permissionIDs[id] = struct{}{}
	}

	for _, p := range s.Permissions {
		parent := strings.TrimSpace(p.Parent)
		if parent == "" {
			continue
		}
		if _, ok := permissionIDs[parent]; !ok {
			return ErrInvalidSchema.With("permission", fmt.Sprintf("unknown parent %q for permission %q", parent, p.ID))
		}
	}

	profileIDs := make(map[string]struct{}, len(s.Profiles))
	hasDefault := false
	for _, pr := range s.Profiles {
		id := strings.TrimSpace(pr.ID)
		if id == "" {
			return ErrInvalidSchema.With("profile", "id required")
		}
		if _, dup := profileIDs[id]; dup {
			return ErrInvalidSchema.With("profile", fmt.Sprintf("duplicate id %q", id))
		}
		profileIDs[id] = struct{}{}
		if id == s.DefaultProfile {
			hasDefault = true
		}
	}
	if len(s.Profiles) > 0 && !hasDefault {
		return ErrInvalidSchema.With("default_profile", fmt.Sprintf("profile %q not defined", s.DefaultProfile))
	}

	for _, pp := range s.ProfilePerms {
		if err := pp.Validate(); err != nil {
			return ErrInvalidSchema.With("profile_permissions", err.Error())
		}
		if _, ok := profileIDs[strings.TrimSpace(pp.Profile)]; !ok {
			return ErrInvalidSchema.With("profile_permissions", fmt.Sprintf("unknown profile %q", pp.Profile))
		}
		if _, ok := permissionIDs[strings.TrimSpace(pp.Permission)]; !ok {
			return ErrInvalidSchema.With("profile_permissions", fmt.Sprintf("unknown permission %q", pp.Permission))
		}
	}

	return nil
}

func (s Setup) AdminProfileID() string {
	for _, pr := range s.Profiles {
		if pr.Superuser {
			return strings.TrimSpace(pr.ID)
		}
	}
	return "admin"
}
