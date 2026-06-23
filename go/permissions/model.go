package permissions

import "time"

type Group struct {
	IDPermissionGroup string
	Name              string
	SortOrder         int
	RoutePrefix       string
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

type Category struct {
	IDPermissionCategory string
	Name                 string
	IDPermissionGroup    string
	SortOrder            int
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

type Permission struct {
	IDPermission         string
	Name                 string
	IDPermissionCategory string
	IDParent             *string
	BeAction             int
	RouteName            string
	Icon                 string
	Enabled              bool
	SortOrder            int
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

type Profile struct {
	IDProfile    string
	Name         string
	IsSuperuser  bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type ProfilePermission struct {
	IDProfile       string
	IDPermission    string
	GrantedActions  *int
}

// FlatPermission is a permission row with effective granted actions for an account.
type FlatPermission struct {
	Permission
	GrantedActions int
}
