package stocks

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/nova/backend/pkg/errors"
)

type StockService struct {
	stockRepo    StockRepository
	binStockRepo BinStockRepository
}

func NewStockService(stockRepo StockRepository, binStockRepo BinStockRepository) *StockService {
	return &StockService{stockRepo: stockRepo, binStockRepo: binStockRepo}
}

func (s *StockService) FindByID(ctx context.Context, id string) (*Stock, error) {
	return s.stockRepo.FindByID(ctx, id)
}

func (s *StockService) FindByPartAndStore(ctx context.Context, partCode, partOrg, storeCode, storeOrg string) (*Stock, error) {
	return s.stockRepo.FindByPartAndStore(ctx, partCode, partOrg, storeCode, storeOrg)
}

func (s *StockService) FindAll(ctx context.Context, tenantID string, storeCode, storeOrg string) ([]*Stock, error) {
	return s.stockRepo.FindAll(ctx, tenantID, storeCode, storeOrg)
}

func (s *StockService) FindLowStock(ctx context.Context, tenantID string) ([]*Stock, error) {
	return s.stockRepo.FindLowStock(ctx, tenantID)
}

func (s *StockService) Create(ctx context.Context, tenantID string, req *CreateStockRequest) (*Stock, error) {
	stock := &Stock{
		ID:         uuid.New().String(),
		PartCode:   req.PartCode,
		PartOrg:    req.PartOrg,
		StoreCode:  req.StoreCode,
		StoreOrg:   req.StoreOrg,
		MinStock:   req.MinStock,
		ReorderQty: req.ReorderQty,
		ActualQty:  req.ActualQty,
		TenantID:   tenantID,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	if err := s.stockRepo.Create(ctx, stock); err != nil {
		return nil, err
	}

	return stock, nil
}

func (s *StockService) Update(ctx context.Context, id string, req *UpdateStockRequest) (*Stock, error) {
	stock, err := s.stockRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if stock == nil {
		return nil, errors.ErrNotFound
	}

	if req.MinStock > 0 {
		stock.MinStock = req.MinStock
	}
	if req.ReorderQty > 0 {
		stock.ReorderQty = req.ReorderQty
	}
	if req.ActualQty > 0 {
		stock.ActualQty = req.ActualQty
	}
	stock.UpdatedAt = time.Now()

	if err := s.stockRepo.Update(ctx, stock); err != nil {
		return nil, err
	}

	return stock, nil
}

func (s *StockService) Delete(ctx context.Context, id string) error {
	stock, err := s.stockRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if stock == nil {
		return errors.ErrNotFound
	}

	return s.stockRepo.Delete(ctx, id)
}

func (s *StockService) AdjustQuantity(ctx context.Context, id string, quantity float64, adjustment float64) (*Stock, error) {
	stock, err := s.stockRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if stock == nil {
		return nil, errors.ErrNotFound
	}

	stock.ActualQty = quantity
	stock.UpdatedAt = time.Now()

	if err := s.stockRepo.UpdateQuantity(ctx, id, quantity); err != nil {
		return nil, err
	}

	return stock, nil
}

// BinStock operations
func (s *StockService) FindBinStockByID(ctx context.Context, id string) (*BinStock, error) {
	return s.binStockRepo.FindByID(ctx, id)
}

func (s *StockService) FindBinStocksByPartAndStore(ctx context.Context, partCode, partOrg, storeCode, storeOrg string) ([]*BinStock, error) {
	return s.binStockRepo.FindByPartAndStore(ctx, partCode, partOrg, storeCode, storeOrg)
}

func (s *StockService) CreateBinStock(ctx context.Context, tenantID string, req *CreateBinStockRequest) (*BinStock, error) {
	binStock := &BinStock{
		ID:         uuid.New().String(),
		PartCode:   req.PartCode,
		PartOrg:    req.PartOrg,
		StoreCode:  req.StoreCode,
		StoreOrg:   req.StoreOrg,
		BinCode:    req.BinCode,
		BinOrg:     req.BinOrg,
		Quantity:   req.Quantity,
		TenantID:   tenantID,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	if err := s.binStockRepo.Create(ctx, binStock); err != nil {
		return nil, err
	}

	return binStock, nil
}

func (s *StockService) UpdateBinStock(ctx context.Context, id string, req *UpdateBinStockRequest) (*BinStock, error) {
	binStock, err := s.binStockRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if binStock == nil {
		return nil, errors.ErrNotFound
	}

	if req.Quantity > 0 {
		binStock.Quantity = req.Quantity
	}
	binStock.UpdatedAt = time.Now()

	if err := s.binStockRepo.Update(ctx, binStock); err != nil {
		return nil, err
	}

	return binStock, nil
}

func (s *StockService) DeleteBinStock(ctx context.Context, id string) error {
	binStock, err := s.binStockRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if binStock == nil {
		return errors.ErrNotFound
	}

	return s.binStockRepo.Delete(ctx, id)
}