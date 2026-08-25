package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/nova/backend/internal/domain/users"
)

// User type alias - User is defined in users package
type User = users.User

type Session struct {
	ID           string     `json:"id"`
	UserCode     string     `json:"userCode"`
	ActiveRole   string     `json:"activeRole"`
	RefreshToken string     `json:"refreshToken"`
	ExpiresAt    time.Time  `json:"expiresAt"`
	IPAddress    string     `json:"ipAddress"`
	UserAgent    string     `json:"userAgent"`
	CreatedAt    time.Time  `json:"createdAt"`
	RevokedAt    *time.Time `json:"revokedAt,omitempty"`
}

type LoginRequest struct {
	Tenant   string `json:"tenant" validate:"required"`
	Code     string `json:"code" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type RegisterRequest struct {
	Tenant   string `json:"tenant" validate:"required"`
	Name     string `json:"name" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refreshToken" validate:"required"`
}

type AuthResponse struct {
	User         *User  `json:"user"`
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	ExpiresIn    int    `json:"expiresIn"`
}

type TokenClaims struct {
	UserCode string   `json:"userCode"`
	Email    string   `json:"email"`
	Name     string   `json:"name"`
	Tenant   string   `json:"tenant"`
	Roles    []string `json:"roles"`
	jwt.RegisteredClaims
}
