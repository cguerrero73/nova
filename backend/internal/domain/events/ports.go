package events

import "context"

type EventRepository interface {
	FindByID(ctx context.Context, id string) (*Event, error)
	FindByCode(ctx context.Context, code string) (*Event, error)
	FindAll(ctx context.Context, tenantID string, org string, limit, offset int) ([]*Event, int, error)
	FindByOrg(ctx context.Context, org string) ([]*Event, error)
	FindByObject(ctx context.Context, objectCode, objectOrg string) ([]*Event, error)
	FindByType(ctx context.Context, typeCode string) ([]*Event, error)
	FindByStatus(ctx context.Context, status string) ([]*Event, error)
	Create(ctx context.Context, e *Event) error
	Update(ctx context.Context, e *Event) error
	Delete(ctx context.Context, id string) error
}
