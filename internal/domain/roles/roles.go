package roles

import "time"

type Role string

// hierarchical roles
const (
	RoleAdmin   Role = "admin"
	RoleManager Role = "manager"
	RoleUser    Role = "user"
)

// global roles
const (
	RolePlatformModerator Role = "platform_moderator"
	RoleReporter          Role = "reporter"
)

// resource‑specific roles
const (
	RoleProjectEditor Role = "project_editor"
	RoleProjectViewer Role = "project_viewer"
)

var RoleHierarchy = map[Role]int{
	RoleAdmin:   3,
	RoleManager: 2,
	RoleUser:    1,
}

// UserRole represents a role assignment, optionally resource-specific and JIT-bound.
type UserRole struct {
	UserID     string     `json:"user_id"`
	Role       Role       `json:"role"`
	ResourceID *string    `json:"resource_id,omitempty"` // if nil then global role
	ExpiresAt  *time.Time `json:"expires_at,omitempty"`  // if nil then permanent role
}

// JITRequest represents a user's request to temporarily activate a role.
type JITRequestDB struct {
	ID              string    `db:"id" json:"id"`
	UserID          string    `db:"user_id" json:"user_id"`
	Role            string    `db:"role" json:"role"`
	ResourceID      *string   `db:"resource_id" json:"resource_id,omitempty"` // nullable
	DurationMinutes int       `db:"duration_minutes" json:"duration_minutes"`
	Reason          *string   `db:"reason" json:"reason,omitempty"`
	Status          string    `db:"status" json:"status"` // pending, approved, rejected
	ApprovedBy      *string   `db:"approved_by" json:"approved_by,omitempty"`
	CreatedAt       time.Time `db:"created_at" json:"created_at"`
	UpdatedAt       time.Time `db:"updated_at" json:"updated_at"`
}
