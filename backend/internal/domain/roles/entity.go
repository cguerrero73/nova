package roles

import "time"

type Role struct {
	ID           string                 `json:"rol_id"`
	Name         string                 `json:"rol_name"`
	Desc         string                 `json:"rol_desc"`
	IsSystem     bool                   `json:"rol_is_system"`
	Permissions  map[string]map[string]bool `json:"rol_permissions"`
	NotUsed      *string                `json:"rol_notused,omitempty"`
	TenantID     string                 `json:"rol_tenant_id"`
	CreatedAt    time.Time              `json:"rol_created_at"`
	UpdatedAt    time.Time              `json:"rol_updated_at"`
}

func (r *Role) HasPermission(screen, action string) bool {
	if r.Permissions == nil {
		return false
	}

	if screenPerms, ok := r.Permissions["*"]; ok {
		if actionPerm, ok := screenPerms["*"]; ok && actionPerm {
			return true
		}
	}

	if screenPerms, ok := r.Permissions[screen]; ok {
		if actionPerm, ok := screenPerms[action]; ok && actionPerm {
			return true
		}
		if actionPerm, ok := screenPerms["*"]; ok && actionPerm {
			return true
		}
	}

	return false
}

func (r *Role) IsActive() bool {
	return r.NotUsed == nil || *r.NotUsed != "+"
}

type UserRole struct {
	ID         string    `json:"urr_id"`
	UserID     string    `json:"urr_user_id"`
	RoleID     string    `json:"urr_role_id"`
	TenantID   string    `json:"urr_tenant_id"`
	AssignedAt time.Time `json:"urr_assigned_at"`
	AssignedBy string    `json:"urr_assigned_by,omitempty"`
}

type RolePermission struct {
	ID       string `json:"rpe_id"`
	RoleID   string `json:"rpe_role_id"`
	Screen   string `json:"rpe_screen"`
	Action   string `json:"rpe_action"`
	Allowed  bool   `json:"rpe_allowed"`
	TenantID string `json:"rpe_tenant_id"`
}