package parts

import "context"

type PartRepository interface {
	FindByID(ctx context.Context, id string) (*Part, error)
	FindByCode(ctx context.Context, code string) (*Part, error)
	FindAll(ctx context.Context, tenantID string, org string, limit, offset int) ([]*Part, int, error)
	FindByOrg(ctx context.Context, org string) ([]*Part, error)
	Create(ctx context.Context, p *Part) error
	Update(ctx context.Context, p *Part) error
	Delete(ctx context.Context, id string) error
}
