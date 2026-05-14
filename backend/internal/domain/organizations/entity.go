package organizations

import "time"

type Organization struct {
	ID        string    `json:"org_id"`
	Code      string    `json:"org_code"`
	Name      string    `json:"org_name"`
	Common    *string   `json:"org_common,omitempty"`
	NotUsed   *string   `json:"org_notused,omitempty"`
	TenantID  string    `json:"org_tenant_id"`
	CreatedAt time.Time `json:"org_created_at"`
	UpdatedAt time.Time `json:"org_updated_at"`
}

func (o *Organization) IsCommon() bool {
	return o.Common != nil && *o.Common == "+"
}

func (o *Organization) IsActive() bool {
	return o.NotUsed == nil || *o.NotUsed != "+"
}