package stores

import "time"

type Store struct {
	ID        string    `json:"str_id"`
	Code      string    `json:"str_code"`
	Name      string    `json:"str_name"`
	Desc      string    `json:"str_desc"`
	Org       string    `json:"str_org"`
	NotUsed   *string   `json:"str_notused,omitempty"`
	TenantID  string    `json:"str_tenant_id"`
	CreatedAt time.Time `json:"str_created_at"`
	UpdatedAt time.Time `json:"str_updated_at"`
}

func (s *Store) IsActive() bool {
	return s.NotUsed == nil || *s.NotUsed != "+"
}

type Bin struct {
	ID        string    `json:"bin_id"`
	Code      string    `json:"bin_code"`
	Desc      string    `json:"bin_desc"`
	Org       string    `json:"bin_org"`
	NotUsed   *string   `json:"bin_notused,omitempty"`
	TenantID  string    `json:"bin_tenant_id"`
	CreatedAt time.Time `json:"bin_created_at"`
	UpdatedAt time.Time `json:"bin_updated_at"`
}

func (b *Bin) IsActive() bool {
	return b.NotUsed == nil || *b.NotUsed != "+"
}