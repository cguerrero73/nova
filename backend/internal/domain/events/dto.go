package events

// CreateEventRequest represents an event creation request
type CreateEventRequest struct {
	Code      string `json:"evt_code" validate:"required"`
	Org       string `json:"evt_org" validate:"required"`
	Desc      string `json:"evt_desc"`
	Type      string `json:"evt_type"`
	RType     string `json:"evt_rtype"`
	Status    string `json:"evt_status"`
	RStatus   string `json:"evt_rstatus"`
	Object    string `json:"evt_object"`
	ObjectOrg string `json:"evt_object_org"`
}

// UpdateEventRequest represents an event update request
type UpdateEventRequest struct {
	Org       string `json:"evt_org"`
	Desc      string `json:"evt_desc"`
	Type      string `json:"evt_type"`
	RType     string `json:"evt_rtype"`
	Status    string `json:"evt_status"`
	RStatus   string `json:"evt_rstatus"`
	Object    string `json:"evt_object"`
	ObjectOrg string `json:"evt_object_org"`
}
