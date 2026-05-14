package auth

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/nova/backend/internal/config"
	apperrors "github.com/nova/backend/pkg/errors"
)

type AuthService struct {
	userRepo    UserRepository
	sessionRepo SessionRepository
	jwtConfig   config.JWTConfig
}

func NewAuthService(userRepo UserRepository, sessionRepo SessionRepository, jwtConfig config.JWTConfig) *AuthService {
	return &AuthService{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
		jwtConfig:   jwtConfig,
	}
}

func (s *AuthService) Login(ctx context.Context, req *LoginRequest) (*AuthResponse, error) {
	user, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return nil, apperrors.ErrInternal
		}
		return nil, apperrors.ErrInvalidCredentials()
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, apperrors.ErrInvalidCredentials()
	}

	return s.generateAuthResponse(ctx, user)
}

func (s *AuthService) Register(ctx context.Context, req *RegisterRequest) (*AuthResponse, error) {
	existing, _ := s.userRepo.FindByEmail(ctx, req.Email)
	if existing != nil {
		return nil, apperrors.ErrUserExists()
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, apperrors.ErrInternal
	}

	user := &User{
		Code:      generateCode(),
		Name:      req.Name,
		Email:     req.Email,
		Password:  string(hashedPassword),
		Status:    "active",
		TenantID:  req.Tenant,
		CreatedAt: time.Now(),
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, apperrors.ErrInternal
	}

	return s.generateAuthResponse(ctx, user)
}

func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string) (*AuthResponse, error) {
	session, err := s.sessionRepo.FindByRefreshToken(ctx, refreshToken)
	if err != nil {
		return nil, apperrors.ErrInvalidToken()
	}

	if session.ExpiresAt.Before(time.Now()) {
		return nil, apperrors.ErrInvalidToken()
	}

	user, err := s.userRepo.FindByCode(ctx, session.UserCode)
	if err != nil {
		return nil, apperrors.ErrUserNotFound()
	}

	// Revoke old session
	if err := s.sessionRepo.Revoke(ctx, session.UserCode); err != nil {
		// Log but continue
	}

	return s.generateAuthResponse(ctx, user)
}

func (s *AuthService) Logout(ctx context.Context, userCode string) error {
	return s.sessionRepo.RevokeAll(ctx, userCode)
}

func (s *AuthService) GetUserByCode(ctx context.Context, code string) (*User, error) {
	return s.userRepo.FindByCode(ctx, code)
}

func (s *AuthService) generateAuthResponse(ctx context.Context, user *User) (*AuthResponse, error) {
	accessToken, err := s.generateAccessToken(user)
	if err != nil {
		return nil, apperrors.ErrInternal
	}

	refreshToken := generateRefreshToken()
	expiresAt := time.Now().Add(7 * 24 * time.Hour) // 7 days

	session := &Session{
		UserCode:     user.Code,
		RefreshToken: refreshToken,
		ExpiresAt:    expiresAt,
		CreatedAt:    time.Now(),
	}

	if err := s.sessionRepo.Create(ctx, session); err != nil {
		return nil, apperrors.ErrInternal
	}

	return &AuthResponse{
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    s.jwtConfig.ExpiryMins * 60,
	}, nil
}

func (s *AuthService) generateAccessToken(user *User) (string, error) {
	claims := TokenClaims{
		UserCode: user.Code,
		Email:    user.Email,
		Name:     user.Name,
		Tenant:   user.TenantID,
		Roles:    []string{}, // TODO: fetch roles
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtConfig.Secret))
}

// ValidateToken validates a JWT token and returns the claims
func (s *AuthService) ValidateToken(tokenString string) (*TokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.jwtConfig.Secret), nil
	})

	if err != nil {
		return nil, apperrors.ErrInvalidToken()
	}

	if claims, ok := token.Claims.(*TokenClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, apperrors.ErrInvalidToken()
}

func generateCode() string {
	return "USR-" + generateRandomString(8)
}

func generateRefreshToken() string {
	return "rt-" + generateRandomString(32)
}

func generateRandomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[time.Now().UnixNano()%int64(len(letters))]
	}
	return string(b)
}
