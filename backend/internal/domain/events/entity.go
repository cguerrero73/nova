package events

import "time"

type Event struct {
	ID        string    `json:"evt_id"`
	Code      string    `json:"evt_code"`
	Org       string    `json:"evt_org"`
	Desc      string    `json:"evt_desc"`
	Type      string    `json:"evt_type"`
	RType     string    `json:"evt_rtype"`
	Status    string    `json:"evt_status"`
	RStatus   string    `json:"evt_rstatus"`
	Object    string    `json:"evt_object"`
	ObjectOrg string    `json:"evt_object_org"`
	NotUsed   *string   `json:"evt_notused,omitempty"`
	TenantID  string    `json:"evt_tenant_id"`
	CreatedAt time.Time `json:"evt_created_at"`
	UpdatedAt time.Time `json:"evt_updated_at"`
	CreatedBy string    `json:"evt_created_by,omitempty"`
	UpdatedBy string    `json:"evt_updated_by,omitempty"`
}

func (e *Event) IsActive() bool {
	return e.NotUsed == nil || *e.NotUsed != "+"
}

func (e *Event) IsOpen() bool {
	// JBST = 'OP' means Open
	return e.Status == "OP"
}