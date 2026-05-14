package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/nova/backend/internal/domain/users"
)

// User type alias - User is defined in users package
type User = users.User

type Session struct {
	ID           string     `json:"ses_id"`
	UserCode     string     `json:"ses_user_code"`
	RefreshToken string     `json:"ses_refresh_token"`
	ExpiresAt    time.Time  `json:"ses_expires_at"`
	IPAddress    string     `json:"ses_ip_address"`
	UserAgent    string     `json:"ses_user_agent"`
	CreatedAt    time.Time  `json:"ses_created_at"`
	RevokedAt    *time.Time `json:"ses_revoked_at,omitempty"`
}

type LoginRequest struct {
	Tenant   string `json:"tenant" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type RegisterRequest struct {
	Tenant   string `json:"tenant" validate:"required"`
	Name     string `json:"name" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type AuthResponse struct {
	User         *User  `json:"user"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
}

type TokenClaims struct {
	UserCode string   `json:"user_code"`
	Email    string   `json:"email"`
	Name     string   `json:"name"`
	Tenant   string   `json:"tenant"`
	Roles    []string `json:"roles"`
	jwt.RegisteredClaims
}
