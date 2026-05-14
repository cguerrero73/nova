package stocks

import "time"

type Stock struct {
	ID         string    `json:"stc_id"`
	PartCode   string    `json:"stc_part_code"`
	PartOrg    string    `json:"stc_part_org"`
	StoreCode  string    `json:"stc_store_code"`
	StoreOrg   string    `json:"stc_store_org"`
	MinStock   float64   `json:"stc_min_stock"`
	ReorderQty float64   `json:"stc_reorder_qty"`
	ActualQty  float64   `json:"stc_actual_qty"`
	NotUsed    *string   `json:"stc_notused,omitempty"`
	TenantID   string    `json:"stc_tenant_id"`
	CreatedAt  time.Time `json:"stc_created_at"`
	UpdatedAt  time.Time `json:"stc_updated_at"`
	CreatedBy  string    `json:"stc_created_by,omitempty"`
	UpdatedBy  string    `json:"stc_updated_by,omitempty"`
}

func (s *Stock) NeedsReorder() bool {
	return s.ActualQty <= s.MinStock
}

type BinStock struct {
	ID         string    `json:"bis_id"`
	PartCode   string    `json:"bis_part_code"`
	PartOrg    string    `json:"bis_part_org"`
	StoreCode  string    `json:"bis_store_code"`
	StoreOrg   string    `json:"bis_store_org"`
	BinCode    string    `json:"bis_bin_code"`
	BinOrg     string    `json:"bis_bin_org"`
	Quantity   float64   `json:"bis_quantity"`
	TenantID   string    `json:"bis_tenant_id"`
	CreatedAt  time.Time `json:"bis_created_at"`
	UpdatedAt  time.Time `json:"bis_updated_at"`
}