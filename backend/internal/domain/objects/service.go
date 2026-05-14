package objects

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/nova/backend/pkg/errors"
)

type ObjectService struct {
	repo ObjectRepository
}

func NewObjectService(repo ObjectRepository) *ObjectService {
	return &ObjectService{repo: repo}
}

func (s *ObjectService) FindByID(ctx context.Context, id string) (*Object, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *ObjectService) FindByCode(ctx context.Context, code string) (*Object, error) {
	return s.repo.FindByCode(ctx, code)
}

func (s *ObjectService) FindAll(ctx context.Context, tenantID string, org string, limit, offset int) ([]*Object, int, error) {
	return s.repo.FindAll(ctx, tenantID, org, limit, offset)
}

func (s *ObjectService) FindByOrg(ctx context.Context, org string) ([]*Object, error) {
	return s.repo.FindByOrg(ctx, org)
}

func (s *ObjectService) FindChildren(ctx context.Context, parentCode, parentOrg string) ([]*Object, error) {
	return s.repo.FindChildren(ctx, parentCode, parentOrg)
}

func (s *ObjectService) Create(ctx context.Context, tenantID string, req *CreateObjectRequest) (*Object, error) {
	obj := &Object{
		ID:        uuid.New().String(),
		Code:      req.Code,
		Type:      req.Type,
		Desc:      req.Desc,
		Serial:    req.Serial,
		Status:    req.Status,
		Org:       req.Org,
		TenantID:  tenantID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if req.ParentCode != "" {
		obj.ParentCode = &req.ParentCode
	}
	if req.ParentOrg != "" {
		obj.ParentOrg = &req.ParentOrg
	}
	if req.InstallDate != "" {
		t, _ := time.Parse("2006-01-02", req.InstallDate)
		obj.InstallDate = &t
	}

	if err := s.repo.Create(ctx, obj); err != nil {
		return nil, err
	}

	return obj, nil
}

func (s *ObjectService) Update(ctx context.Context, id string, req *UpdateObjectRequest) (*Object, error) {
	obj, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if obj == nil {
		return nil, errors.ErrNotFound
	}

	if req.Type != "" {
		obj.Type = req.Type
	}
	if req.Desc != "" {
		obj.Desc = req.Desc
	}
	if req.Serial != "" {
		obj.Serial = req.Serial
	}
	if req.Status != "" {
		obj.Status = req.Status
	}
	if req.Org != "" {
		obj.Org = req.Org
	}
	if req.ParentCode != "" {
		obj.ParentCode = &req.ParentCode
	}
	if req.ParentOrg != "" {
		obj.ParentOrg = &req.ParentOrg
	}
	if req.InstallDate != "" {
		t, _ := time.Parse("2006-01-02", req.InstallDate)
		obj.InstallDate = &t
	}
	obj.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, obj); err != nil {
		return nil, err
	}

	return obj, nil
}

func (s *ObjectService) Delete(ctx context.Context, id string) error {
	obj, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if obj == nil {
		return errors.ErrNotFound
	}

	return s.repo.Delete(ctx, id)
}