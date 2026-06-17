package tenants

import "time"

const (
	RoleOwner  = "owner"
	RoleAdmin  = "admin"
	RoleStaff  = "staff"
	RoleViewer = "viewer"

	StatusTrial     = "trial"
	StatusActive    = "active"
	StatusSuspended = "suspended"
)

type Tenant struct {
	ID       string
	Slug     string
	Name     string
	Timezone string
	Status   string
}

type Membership struct {
	Tenant Tenant
	Role   string
}

type Invite struct {
	ID        string
	TenantID  string
	Email     string
	Role      string
	ExpiresAt time.Time
}
