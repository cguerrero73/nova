package objects

// CreateObjectRequest represents an object creation request
type CreateObjectRequest struct {
	Code        string `json:"obj_code" validate:"required"`
	Type        string `json:"obj_type" validate:"required"`
	Desc        string `json:"obj_desc"`
	Serial      string `json:"obj_serial"`
	Status      string `json:"obj_status"`
	Org         string `json:"obj_org" validate:"required"`
	ParentCode  string `json:"obj_parent_code"`
	ParentOrg   string `json:"obj_parent_org"`
	InstallDate string `json:"obj_install_date"`
}

// UpdateObjectRequest represents an object update request
type UpdateObjectRequest struct {
	Type        string `json:"obj_type"`
	Desc        string `json:"obj_desc"`
	Serial      string `json:"obj_serial"`
	Status      string `json:"obj_status"`
	Org         string `json:"obj_org"`
	ParentCode  string `json:"obj_parent_code"`
	ParentOrg   string `json:"obj_parent_org"`
	InstallDate string `json:"obj_install_date"`
}
