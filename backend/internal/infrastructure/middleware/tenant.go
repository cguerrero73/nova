package middleware

import (
	"github.com/gofiber/fiber/v2"
)

const TenantContextKey = "tenant"

type TenantMiddleware struct{}

func NewTenantMiddleware() *TenantMiddleware {
	return &TenantMiddleware{}
}

// ExtractTenant extracts tenant code from query param or header
func (m *TenantMiddleware) ExtractTenant() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Try query param first
		tenant := c.Query("tenant")

		// Fallback to header
		if tenant == "" {
			tenant = c.Get("X-Tenant-Code")
		}

		if tenant != "" {
			c.Locals(TenantContextKey, tenant)
		}

		return c.Next()
	}
}

// GetTenant retrieves the tenant code from context
func GetTenant(c *fiber.Ctx) string {
	if tenant, ok := c.Locals(TenantContextKey).(string); ok {
		return tenant
	}
	return ""
}
