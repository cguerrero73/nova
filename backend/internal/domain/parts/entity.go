package parts

import "time"

type Part struct {
	ID        string    `json:"par_id"`
	Code      string    `json:"par_code"`
	Desc      string    `json:"par_desc"`
	NotUsed   *string   `json:"par_notused,omitempty"`
	Org       string    `json:"par_org"`
	TenantID  string    `json:"par_tenant_id"`
	CreatedAt time.Time `json:"par_created_at"`
	UpdatedAt time.Time `json:"par_updated_at"`
	CreatedBy string    `json:"par_created_by,omitempty"`
	UpdatedBy string    `json:"par_updated_by,omitempty"`
}

func (p *Part) IsActive() bool {
	return p.NotUsed == nil || *p.NotUsed != "+"
}