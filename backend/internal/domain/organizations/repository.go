package organizations

import "context"

type OrganizationRepository interface {
	FindByID(ctx context.Context, id string) (*Organization, error)
	FindByCode(ctx context.Context, code string) (*Organization, error)
	FindAll(ctx context.Context, tenantID string) ([]*Organization, error)
	FindCommon(ctx context.Context, tenantID string) (*Organization, error)
	Create(ctx context.Context, org *Organization) error
	Update(ctx context.Context, org *Organization) error
	Delete(ctx context.Context, id string) error
}
