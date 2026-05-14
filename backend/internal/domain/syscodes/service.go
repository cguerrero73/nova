package syscodes

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/nova/backend/pkg/errors"
)

type SysCodeService struct {
	repo SysCodeRepository
}

func NewSysCodeService(repo SysCodeRepository) *SysCodeService {
	return &SysCodeService{repo: repo}
}

func (s *SysCodeService) FindByID(ctx context.Context, id string) (*SysCode, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *SysCodeService) FindByTypeAndCode(ctx context.Context, codeType, code string) (*SysCode, error) {
	return s.repo.FindByTypeAndCode(ctx, codeType, code)
}

func (s *SysCodeService) FindByType(ctx context.Context, codeType string) ([]*SysCode, error) {
	return s.repo.FindByType(ctx, codeType)
}

func (s *SysCodeService) FindAll(ctx context.Context) ([]*SysCode, error) {
	return s.repo.FindAll(ctx)
}

func (s *SysCodeService) Create(ctx context.Context, req *CreateSysCodeRequest) (*SysCode, error) {
	sysCode := &SysCode{
		ID:        uuid.New().String(),
		Type:      req.Type,
		Code:      req.Code,
		UCode:     req.UCode,
		Desc:      req.Desc,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if req.System == "+" {
		sysCode.System = &req.System
	}

	if err := s.repo.Create(ctx, sysCode); err != nil {
		return nil, err
	}

	return sysCode, nil
}

func (s *SysCodeService) Update(ctx context.Context, id string, req *UpdateSysCodeRequest) (*SysCode, error) {
	sysCode, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if sysCode == nil {
		return nil, errors.ErrNotFound
	}

	// System codes cannot be modified
	if sysCode.IsSystem() {
		return nil, errors.New("SYSTEM_CODE", "System codes cannot be modified", 400)
	}

	if req.UCode != "" {
		sysCode.UCode = req.UCode
	}
	if req.Desc != "" {
		sysCode.Desc = req.Desc
	}
	sysCode.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, sysCode); err != nil {
		return nil, err
	}

	return sysCode, nil
}

func (s *SysCodeService) Delete(ctx context.Context, id string) error {
	sysCode, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if sysCode == nil {
		return errors.ErrNotFound
	}

	// System codes cannot be deleted
	if sysCode.IsSystem() {
		return errors.New("SYSTEM_CODE", "System codes cannot be deleted", 400)
	}

	return s.repo.Delete(ctx, id)
}