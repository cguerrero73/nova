package users

import "context"

type UserRepository interface {
	FindByID(ctx context.Context, id string) (*User, error)
	FindByCode(ctx context.Context, code string) (*User, error)
	FindByEmail(ctx context.Context, email string) (*User, error)
	FindAll(ctx context.Context, tenantID string, limit, offset int) ([]*User, int, error)
	Create(ctx context.Context, user *User) error
	Update(ctx context.Context, user *User) error
	Delete(ctx context.Context, id string) error
}
