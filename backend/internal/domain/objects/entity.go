package objects

import "time"

type Object struct {
	ID           string    `json:"obj_id"`
	Code         string    `json:"obj_code"`
	Type         string    `json:"obj_type"`
	Desc         string    `json:"obj_desc"`
	Serial       string    `json:"obj_serial"`
	Status       string    `json:"obj_status"`
	Org          string    `json:"obj_org"`
	ParentCode   *string   `json:"obj_parent_code,omitempty"`
	ParentOrg    *string   `json:"obj_parent_org,omitempty"`
	InstallDate  *time.Time `json:"obj_install_date,omitempty"`
	NotUsed      *string   `json:"obj_notused,omitempty"`
	TenantID     string    `json:"obj_tenant_id"`
	CreatedAt    time.Time `json:"obj_created_at"`
	UpdatedAt    time.Time `json:"obj_updated_at"`
	CreatedBy    string    `json:"obj_created_by,omitempty"`
	UpdatedBy    string    `json:"obj_updated_by,omitempty"`
}

func (o *Object) IsActive() bool {
	return o.NotUsed == nil || *o.NotUsed != "+"
}