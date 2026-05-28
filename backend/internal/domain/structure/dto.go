package structure

// CreateStructureRequest represents a structure creation request
type CreateStructureRequest struct {
	ParentCode string `json:"sct_parent_code" validate:"required"`
	ParentOrg  string `json:"sct_parent_org" validate:"required"`
	ChildCode  string `json:"sct_child_code" validate:"required"`
	ChildOrg   string `json:"sct_child_org" validate:"required"`
	Cost       string `json:"sct_cost"`
	Meter      string `json:"sct_meter"`
}

// UpdateStructureRequest represents a structure update request
type UpdateStructureRequest struct {
	Cost  string `json:"sct_cost"`
	Meter string `json:"sct_meter"`
}
