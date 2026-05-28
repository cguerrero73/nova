package syscodes

import "context"

type SysCodeRepository interface {
	FindByID(ctx context.Context, id string) (*SysCode, error)
	FindByTypeAndCode(ctx context.Context, codeType, code string) (*SysCode, error)
	FindByType(ctx context.Context, codeType string) ([]*SysCode, error)
	FindByUCode(ctx context.Context, ucode string) (*SysCode, error)
	FindAll(ctx context.Context) ([]*SysCode, error)
	Create(ctx context.Context, s *SysCode) error
	Update(ctx context.Context, s *SysCode) error
	Delete(ctx context.Context, id string) error
}
