package organizations

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/nova/backend/pkg/errors"
)

type OrganizationService struct {
	repo OrganizationRepository
}

func NewOrganizationService(repo OrganizationRepository) *OrganizationService {
	return &OrganizationService{repo: repo}
}

func (s *OrganizationService) FindByID(ctx context.Context, id string) (*Organization, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *OrganizationService) FindByCode(ctx context.Context, code string) (*Organization, error) {
	return s.repo.FindByCode(ctx, code)
}

func (s *OrganizationService) FindAll(ctx context.Context, tenantID string) ([]*Organization, error) {
	return s.repo.FindAll(ctx, tenantID)
}

func (s *OrganizationService) Create(ctx context.Context, tenantID string, req *CreateOrganizationRequest) (*Organization, error) {
	// Check if code already exists
	existing, _ := s.repo.FindByCode(ctx, req.Code)
	if existing != nil {
		return nil, errors.New("CODE_EXISTS", "Organization code already exists", 409)
	}

	org := &Organization{
		ID:        uuid.New().String(),
		Code:      req.Code,
		Name:      req.Name,
		TenantID:  tenantID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if req.Common == "+" {
		org.Common = &req.Common
	}

	if err := s.repo.Create(ctx, org); err != nil {
		return nil, err
	}

	return org, nil
}

func (s *OrganizationService) Update(ctx context.Context, id string, req *UpdateOrganizationRequest) (*Organization, error) {
	org, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if org == nil {
		return nil, errors.New("NOT_FOUND", "Organization not found", 404)
	}

	if req.Name != "" {
		org.Name = req.Name
	}
	if req.Common != "" {
		if req.Common == "+" {
			org.Common = &req.Common
		} else {
			org.Common = nil
		}
	}
	org.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, org); err != nil {
		return nil, err
	}

	return org, nil
}

func (s *OrganizationService) Delete(ctx context.Context, id string) error {
	org, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if org == nil {
		return errors.New("NOT_FOUND", "Organization not found", 404)
	}

	return s.repo.Delete(ctx, id)
}