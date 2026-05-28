package parts

// CreatePartRequest represents a part creation request
type CreatePartRequest struct {
	Code string `json:"par_code" validate:"required"`
	Desc string `json:"par_desc"`
	Org  string `json:"par_org" validate:"required"`
}

// UpdatePartRequest represents a part update request
type UpdatePartRequest struct {
	Desc string `json:"par_desc"`
}
