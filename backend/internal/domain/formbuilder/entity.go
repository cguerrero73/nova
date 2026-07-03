package formbuilder

import (
	"encoding/json"
	"time"
)

// Form represents a logical form definition within a tenant schema.
type Form struct {
	ID          int64     `json:"frm_id"`
	Key         string    `json:"frm_key"`
	Name        string    `json:"frm_name"`
	Description string    `json:"frm_description"`
	Status      string    `json:"frm_status"`
	CreatedBy   string    `json:"frm_created_by"`
	CreatedAt   time.Time `json:"frm_created_at"`
	UpdatedAt   time.Time `json:"frm_updated_at"`
}

// Layout represents a named layout variant within a form.
type Layout struct {
	ID                 int64     `json:"fl_id"`
	FormID             int64     `json:"fl_form_id"`
	Name               string    `json:"fl_name"`
	DisplayName        string    `json:"fl_display_name"`
	Description        string    `json:"fl_description"`
	Status             string    `json:"fl_status"`
	DraftVersionID     *int64    `json:"fl_draft_version_id"`
	PublishedVersionID *int64    `json:"fl_published_version_id"`
	CreatedBy          string    `json:"fl_created_by"`
	CreatedAt          time.Time `json:"fl_created_at"`
	UpdatedAt          time.Time `json:"fl_updated_at"`
}

// LayoutVersion represents an immutable snapshot of a layout's definition.
type LayoutVersion struct {
	ID            int64           `json:"flv_id"`
	LayoutID      int64           `json:"flv_layout_id"`
	VersionNumber int             `json:"flv_version_number"`
	Kind          string          `json:"flv_kind"`
	Description   string          `json:"flv_description"`
	Definition    json.RawMessage `json:"flv_definition"`
	CreatedBy     string          `json:"flv_created_by"`
	CreatedAt     time.Time       `json:"flv_created_at"`
	PublishedAt   *time.Time      `json:"flv_published_at"`
}

// RoleAssignment maps a tenant role to a specific layout for a form.
type RoleAssignment struct {
	ID         int64      `json:"fra_id"`
	FormID     int64      `json:"fra_form_id"`
	LayoutID   int64      `json:"fra_layout_id"`
	RoleName   string     `json:"fra_role_name"`
	AssignedAt time.Time  `json:"fra_assigned_at"`
	RevokedAt  *time.Time `json:"fra_revoked_at"`
	AssignedBy string     `json:"fra_assigned_by"`
}

// AuditEntry records an immutable mutation event.
type AuditEntry struct {
	ID          int64           `json:"fal_id"`
	ActorUserID string          `json:"fal_actor_user_id"`
	Action      string          `json:"fal_action"`
	EntityType  string          `json:"fal_entity_type"`
	EntityID    int64           `json:"fal_entity_id"`
	Metadata    json.RawMessage `json:"fal_metadata"`
	Note        string          `json:"fal_note"`
	CreatedAt   time.Time       `json:"fal_created_at"`
}
