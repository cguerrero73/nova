package objects

import "context"

type ObjectRepository interface {
	FindByID(ctx context.Context, id string) (*Object, error)
	FindByCode(ctx context.Context, code string) (*Object, error)
	FindAll(ctx context.Context, tenantID string, org string, limit, offset int) ([]*Object, int, error)
	FindByOrg(ctx context.Context, org string) ([]*Object, error)
	FindChildren(ctx context.Context, parentCode, parentOrg string) ([]*Object, error)
	Create(ctx context.Context, o *Object) error
	Update(ctx context.Context, o *Object) error
	Delete(ctx context.Context, id string) error
}
