package structure

import "time"

type Structure struct {
	ID         string    `json:"sct_id"`
	ParentCode string    `json:"sct_parent_code"`
	ParentOrg  string    `json:"sct_parent_org"`
	ChildCode  string    `json:"sct_child_code"`
	ChildOrg   string    `json:"sct_child_org"`
	Cost       *string   `json:"sct_cost,omitempty"`
	Meter      *string   `json:"sct_meter,omitempty"`
	TenantID   string    `json:"sct_tenant_id"`
	CreatedAt  time.Time `json:"sct_created_at"`
	UpdatedAt  time.Time `json:"sct_updated_at"`
}

func (s *Structure) InheritsCost() bool {
	return s.Cost != nil && *s.Cost == "+"
}

func (s *Structure) InheritsMeter() bool {
	return s.Meter != nil && *s.Meter == "+"
}