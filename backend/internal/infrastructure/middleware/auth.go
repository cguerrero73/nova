package middleware

import (
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"

	"github.com/nova/backend/internal/config"
	"github.com/nova/backend/internal/domain/auth"
)

type AuthMiddleware struct {
	jwtConfig config.JWTConfig
}

func NewAuthMiddleware(jwtConfig config.JWTConfig) *AuthMiddleware {
	return &AuthMiddleware{jwtConfig: jwtConfig}
}

func (m *AuthMiddleware) Authenticate() fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(401).JSON(fiber.Map{
				"success": false,
				"error": fiber.Map{
					"code":    "UNAUTHORIZED",
					"message": "Authorization header is required",
				},
			})
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return c.Status(401).JSON(fiber.Map{
				"success": false,
				"error": fiber.Map{
					"code":    "UNAUTHORIZED",
					"message": "Invalid authorization header format",
				},
			})
		}

		tokenString := parts[1]
		claims, err := m.ValidateToken(tokenString)
		if err != nil {
			return c.Status(401).JSON(fiber.Map{
				"success": false,
				"error": fiber.Map{
					"code":    "INVALID_TOKEN",
					"message": "Invalid or expired token",
				},
			})
		}

		// Store claims in context
		c.Locals("user", claims)
		c.Locals("tenant", claims.Tenant)

		return c.Next()
	}
}

func (m *AuthMiddleware) ValidateToken(tokenString string) (*auth.TokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &auth.TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(m.jwtConfig.Secret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*auth.TokenClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fiber.NewError(fiber.StatusUnauthorized, "Invalid token")
}

func GetUserClaims(c *fiber.Ctx) *auth.TokenClaims {
	if claims, ok := c.Locals("user").(*auth.TokenClaims); ok {
		return claims
	}
	return nil
}

func GetTenantFromContext(c *fiber.Ctx) string {
	if tenant, ok := c.Locals("tenant").(string); ok {
		return tenant
	}
	return ""
}

func GenerateToken(claims *auth.TokenClaims, secret string, expiry time.Duration) (string, error) {
	claims.RegisteredClaims = jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiry)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		NotBefore: jwt.NewNumericDate(time.Now()),
		Issuer:    "nova-eam",
		Subject:   claims.UserCode,
		Audience:  jwt.ClaimStrings{claims.Tenant},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}
