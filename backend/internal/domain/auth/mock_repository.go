package auth

import (
	"context"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type MockUserRepository struct {
	users map[string]*User
}

func NewMockUserRepository() *MockUserRepository {
	// Create a default admin user for testing
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	
	repo := &MockUserRepository{
		users: map[string]*User{
			"admin@nova.com": {
				ID:        "usr-001",
				Code:      "USR-ADMIN001",
				Name:      "Admin User",
				Email:     "admin@nova.com",
				Password:  string(hashedPassword),
				Phone:     "+1234567890",
				Status:    "active",
				DefaultOrg: "*",
				TenantID:  "default",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
		},
	}
	return repo
}

func (r *MockUserRepository) FindByEmail(ctx context.Context, email string) (*User, error) {
	if user, ok := r.users[email]; ok {
		return user, nil
	}
	return nil, context.Canceled
}

func (r *MockUserRepository) FindByCode(ctx context.Context, code string) (*User, error) {
	for _, user := range r.users {
		if user.Code == code {
			return user, nil
		}
	}
	return nil, context.Canceled
}

func (r *MockUserRepository) Create(ctx context.Context, user *User) error {
	r.users[user.Email] = user
	return nil
}

func (r *MockUserRepository) Update(ctx context.Context, user *User) error {
	r.users[user.Email] = user
	return nil
}

type MockSessionRepository struct {
	sessions map[string]*Session
}

func NewMockSessionRepository() *MockSessionRepository {
	return &MockSessionRepository{
		sessions: make(map[string]*Session),
	}
}

func (r *MockSessionRepository) Create(ctx context.Context, session *Session) error {
	r.sessions[session.RefreshToken] = session
	return nil
}

func (r *MockSessionRepository) FindByRefreshToken(ctx context.Context, token string) (*Session, error) {
	if session, ok := r.sessions[token]; ok {
		return session, nil
	}
	return nil, context.Canceled
}

func (r *MockSessionRepository) Revoke(ctx context.Context, userCode string) error {
	for token, session := range r.sessions {
		if session.UserCode == userCode {
			delete(r.sessions, token)
		}
	}
	return nil
}

func (r *MockSessionRepository) RevokeAll(ctx context.Context, userCode string) error {
	return r.Revoke(ctx, userCode)
}
