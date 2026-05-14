package auth

import (
	"context"
)

// UserRepository defines auth-specific user operations (minimal set for authentication)
type UserRepository interface {
	FindByEmail(ctx context.Context, email string) (*User, error)
	FindByCode(ctx context.Context, code string) (*User, error)
	Create(ctx context.Context, user *User) error
}

// SessionRepository defines session operations
type SessionRepository interface {
	Create(ctx context.Context, session *Session) error
	FindByRefreshToken(ctx context.Context, token string) (*Session, error)
	Revoke(ctx context.Context, userCode string) error
	RevokeAll(ctx context.Context, userCode string) error
}
