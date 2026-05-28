package stores

type CreateStoreRequest struct {
	Code string `json:"str_code" validate:"required"`
	Name string `json:"str_name" validate:"required"`
	Desc string `json:"str_desc"`
	Org  string `json:"str_org" validate:"required"`
}

type UpdateStoreRequest struct {
	Name string `json:"str_name"`
	Desc string `json:"str_desc"`
}

type CreateBinRequest struct {
	Code string `json:"bin_code" validate:"required"`
	Desc string `json:"bin_desc"`
	Org  string `json:"bin_org" validate:"required"`
}

type UpdateBinRequest struct {
	Desc string `json:"bin_desc"`
}
