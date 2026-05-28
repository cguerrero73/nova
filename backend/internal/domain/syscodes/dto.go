package syscodes

// CreateSysCodeRequest represents a syscode creation request
type CreateSysCodeRequest struct {
	Type   string `json:"sys_type" validate:"required"`
	Code   string `json:"sys_code" validate:"required"`
	UCode  string `json:"sys_ucode" validate:"required"`
	Desc   string `json:"sys_desc"`
	System string `json:"sys_system"`
}

// UpdateSysCodeRequest represents a syscode update request
type UpdateSysCodeRequest struct {
	UCode string `json:"sys_ucode"`
	Desc  string `json:"sys_desc"`
}
