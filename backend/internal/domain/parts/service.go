package parts

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/nova/backend/pkg/errors"
)

type PartService struct {
	repo PartRepository
}

func NewPartService(repo PartRepository) *PartService {
	return &PartService{repo: repo}
}

func (s *PartService) FindByID(ctx context.Context, id string) (*Part, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *PartService) FindByCode(ctx context.Context, code string) (*Part, error) {
	return s.repo.FindByCode(ctx, code)
}

func (s *PartService) FindAll(ctx context.Context, tenantID string, org string, limit, offset int) ([]*Part, int, error) {
	return s.repo.FindAll(ctx, tenantID, org, limit, offset)
}

func (s *PartService) FindByOrg(ctx context.Context, org string) ([]*Part, error) {
	return s.repo.FindByOrg(ctx, org)
}

func (s *PartService) Create(ctx context.Context, tenantID string, req *CreatePartRequest) (*Part, error) {
	part := &Part{
		ID:        uuid.New().String(),
		Code:      req.Code,
		Desc:      req.Desc,
		Org:       req.Org,
		TenantID:  tenantID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.repo.Create(ctx, part); err != nil {
		return nil, err
	}

	return part, nil
}

func (s *PartService) Update(ctx context.Context, id string, req *UpdatePartRequest) (*Part, error) {
	part, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if part == nil {
		return nil, errors.ErrNotFound
	}

	if req.Desc != "" {
		part.Desc = req.Desc
	}
	part.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, part); err != nil {
		return nil, err
	}

	return part, nil
}

func (s *PartService) Delete(ctx context.Context, id string) error {
	part, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if part == nil {
		return errors.ErrNotFound
	}

	return s.repo.Delete(ctx, id)
}