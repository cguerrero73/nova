package stocks

// CreateStockRequest represents a stock creation request
type CreateStockRequest struct {
	PartCode   string  `json:"stc_part_code" validate:"required"`
	PartOrg    string  `json:"stc_part_org" validate:"required"`
	StoreCode  string  `json:"stc_store_code" validate:"required"`
	StoreOrg   string  `json:"stc_store_org" validate:"required"`
	MinStock   float64 `json:"stc_min_stock"`
	ReorderQty float64 `json:"stc_reorder_qty"`
	ActualQty  float64 `json:"stc_actual_qty"`
}

// UpdateStockRequest represents a stock update request
type UpdateStockRequest struct {
	MinStock   float64 `json:"stc_min_stock"`
	ReorderQty float64 `json:"stc_reorder_qty"`
	ActualQty  float64 `json:"stc_actual_qty"`
}

// CreateBinStockRequest represents a bin stock creation request
type CreateBinStockRequest struct {
	PartCode  string  `json:"bis_part_code" validate:"required"`
	PartOrg   string  `json:"bis_part_org" validate:"required"`
	StoreCode string  `json:"bis_store_code" validate:"required"`
	StoreOrg  string  `json:"bis_store_org" validate:"required"`
	BinCode   string  `json:"bis_bin_code" validate:"required"`
	BinOrg    string  `json:"bis_bin_org" validate:"required"`
	Quantity  float64 `json:"bis_quantity"`
}

// UpdateBinStockRequest represents a bin stock update request
type UpdateBinStockRequest struct {
	Quantity float64 `json:"bis_quantity"`
}
