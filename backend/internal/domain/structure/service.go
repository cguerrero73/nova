package structure

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/nova/backend/pkg/errors"
)

type StructureService struct {
	repo StructureRepository
}

func NewStructureService(repo StructureRepository) *StructureService {
	return &StructureService{repo: repo}
}

func (s *StructureService) FindByID(ctx context.Context, id string) (*Structure, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *StructureService) FindByParent(ctx context.Context, parentCode, parentOrg string) ([]*Structure, error) {
	return s.repo.FindByParent(ctx, parentCode, parentOrg)
}

func (s *StructureService) FindAll(ctx context.Context, tenantID string) ([]*Structure, error) {
	return s.repo.FindAll(ctx, tenantID)
}

func (s *StructureService) Create(ctx context.Context, tenantID string, req *CreateStructureRequest) (*Structure, error) {
	structure := &Structure{
		ID:         uuid.New().String(),
		ParentCode: req.ParentCode,
		ParentOrg:  req.ParentOrg,
		ChildCode:  req.ChildCode,
		ChildOrg:   req.ChildOrg,
		TenantID:   tenantID,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	if req.Cost == "+" {
		structure.Cost = &req.Cost
	}
	if req.Meter == "+" {
		structure.Meter = &req.Meter
	}

	if err := s.repo.Create(ctx, structure); err != nil {
		return nil, err
	}

	return structure, nil
}

func (s *StructureService) Update(ctx context.Context, id string, req *UpdateStructureRequest) (*Structure, error) {
	structure, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if structure == nil {
		return nil, errors.ErrNotFound
	}

	if req.Cost != "" {
		if req.Cost == "+" {
			structure.Cost = &req.Cost
		} else {
			structure.Cost = nil
		}
	}
	if req.Meter != "" {
		if req.Meter == "+" {
			structure.Meter = &req.Meter
		} else {
			structure.Meter = nil
		}
	}
	structure.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, structure); err != nil {
		return nil, err
	}

	return structure, nil
}

func (s *StructureService) Delete(ctx context.Context, id string) error {
	structure, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if structure == nil {
		return errors.ErrNotFound
	}

	return s.repo.Delete(ctx, id)
}