package organizations

// CreateOrganizationRequest represents an organization creation request
type CreateOrganizationRequest struct {
	Code   string `json:"org_code" validate:"required"`
	Name   string `json:"org_name" validate:"required"`
	Common string `json:"org_common"`
}

// UpdateOrganizationRequest represents an organization update request
type UpdateOrganizationRequest struct {
	Name   string `json:"org_name"`
	Common string `json:"org_common"`
}
