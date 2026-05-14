package stores

import "context"

type StoreRepository interface {
	FindByID(ctx context.Context, id string) (*Store, error)
	FindByCode(ctx context.Context, code string) (*Store, error)
	FindAll(ctx context.Context, tenantID string, org string) ([]*Store, error)
	FindByOrg(ctx context.Context, org string) ([]*Store, error)
	Create(ctx context.Context, s *Store) error
	Update(ctx context.Context, s *Store) error
	Delete(ctx context.Context, id string) error
}

type BinRepository interface {
	FindByID(ctx context.Context, id string) (*Bin, error)
	FindByCode(ctx context.Context, code, org string) (*Bin, error)
	FindByOrg(ctx context.Context, org string) ([]*Bin, error)
	Create(ctx context.Context, b *Bin) error
	Update(ctx context.Context, b *Bin) error
	Delete(ctx context.Context, id string) error
}
