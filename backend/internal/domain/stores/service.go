package stores

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/nova/backend/pkg/errors"
)

type StoreService struct {
	storeRepo StoreRepository
	binRepo   BinRepository
}

func NewStoreService(storeRepo StoreRepository, binRepo BinRepository) *StoreService {
	return &StoreService{storeRepo: storeRepo, binRepo: binRepo}
}

func (s *StoreService) FindByID(ctx context.Context, id string) (*Store, error) {
	return s.storeRepo.FindByID(ctx, id)
}

func (s *StoreService) FindByCode(ctx context.Context, code string) (*Store, error) {
	return s.storeRepo.FindByCode(ctx, code)
}

func (s *StoreService) FindAll(ctx context.Context, tenantID string, org string) ([]*Store, error) {
	return s.storeRepo.FindAll(ctx, tenantID, org)
}

func (s *StoreService) Create(ctx context.Context, tenantID string, req *CreateStoreRequest) (*Store, error) {
	store := &Store{
		ID:        uuid.New().String(),
		Code:      req.Code,
		Name:      req.Name,
		Desc:      req.Desc,
		Org:       req.Org,
		TenantID:  tenantID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.storeRepo.Create(ctx, store); err != nil {
		return nil, err
	}

	return store, nil
}

func (s *StoreService) Update(ctx context.Context, id string, req *UpdateStoreRequest) (*Store, error) {
	store, err := s.storeRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if store == nil {
		return nil, errors.ErrNotFound
	}

	if req.Name != "" {
		store.Name = req.Name
	}
	if req.Desc != "" {
		store.Desc = req.Desc
	}
	store.UpdatedAt = time.Now()

	if err := s.storeRepo.Update(ctx, store); err != nil {
		return nil, err
	}

	return store, nil
}

func (s *StoreService) Delete(ctx context.Context, id string) error {
	store, err := s.storeRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if store == nil {
		return errors.ErrNotFound
	}

	return s.storeRepo.Delete(ctx, id)
}

// Bin operations
func (s *StoreService) FindBinByID(ctx context.Context, id string) (*Bin, error) {
	return s.binRepo.FindByID(ctx, id)
}

func (s *StoreService) FindBinsByOrg(ctx context.Context, org string) ([]*Bin, error) {
	return s.binRepo.FindByOrg(ctx, org)
}

func (s *StoreService) CreateBin(ctx context.Context, tenantID string, req *CreateBinRequest) (*Bin, error) {
	bin := &Bin{
		ID:        uuid.New().String(),
		Code:      req.Code,
		Desc:      req.Desc,
		Org:       req.Org,
		TenantID:  tenantID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.binRepo.Create(ctx, bin); err != nil {
		return nil, err
	}

	return bin, nil
}

func (s *StoreService) UpdateBin(ctx context.Context, id string, req *UpdateBinRequest) (*Bin, error) {
	bin, err := s.binRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if bin == nil {
		return nil, errors.ErrNotFound
	}

	if req.Desc != "" {
		bin.Desc = req.Desc
	}
	bin.UpdatedAt = time.Now()

	if err := s.binRepo.Update(ctx, bin); err != nil {
		return nil, err
	}

	return bin, nil
}

func (s *StoreService) DeleteBin(ctx context.Context, id string) error {
	bin, err := s.binRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if bin == nil {
		return errors.ErrNotFound
	}

	return s.binRepo.Delete(ctx, id)
}