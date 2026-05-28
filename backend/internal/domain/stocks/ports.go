package stocks

import "context"

type StockRepository interface {
	FindByID(ctx context.Context, id string) (*Stock, error)
	FindByPartAndStore(ctx context.Context, partCode, partOrg, storeCode, storeOrg string) (*Stock, error)
	FindByPart(ctx context.Context, partCode, partOrg string) ([]*Stock, error)
	FindAll(ctx context.Context, tenantID string, storeCode, storeOrg string) ([]*Stock, error)
	FindLowStock(ctx context.Context, tenantID string) ([]*Stock, error)
	Create(ctx context.Context, s *Stock) error
	Update(ctx context.Context, s *Stock) error
	UpdateQuantity(ctx context.Context, id string, qty float64) error
	Delete(ctx context.Context, id string) error
}

type BinStockRepository interface {
	FindByID(ctx context.Context, id string) (*BinStock, error)
	FindByPartStoreBin(ctx context.Context, partCode, partOrg, storeCode, storeOrg, binCode, binOrg string) (*BinStock, error)
	FindByPartAndStore(ctx context.Context, partCode, partOrg, storeCode, storeOrg string) ([]*BinStock, error)
	Create(ctx context.Context, b *BinStock) error
	Update(ctx context.Context, b *BinStock) error
	Delete(ctx context.Context, id string) error
}
