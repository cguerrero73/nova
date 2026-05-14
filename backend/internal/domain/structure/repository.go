package structure

import "context"

type StructureRepository interface {
	FindByID(ctx context.Context, id string) (*Structure, error)
	FindByParent(ctx context.Context, parentCode, parentOrg string) ([]*Structure, error)
	FindByChild(ctx context.Context, childCode, childOrg string) ([]*Structure, error)
	FindAll(ctx context.Context, tenantID string) ([]*Structure, error)
	Create(ctx context.Context, s *Structure) error
	Update(ctx context.Context, s *Structure) error
	Delete(ctx context.Context, id string) error
	DeleteByParentChild(ctx context.Context, parentCode, parentOrg, childCode, childOrg string) error
}
