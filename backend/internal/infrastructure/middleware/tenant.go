package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/nova/backend/internal/infrastructure/db"
)

const TenantContextKey = "tenant"

// TenantMiddleware extracts tenant from request and sets up tenant context
type TenantMiddleware struct {
	pool *pgxpool.Pool
}

func NewTenantMiddleware(pool *pgxpool.Pool) *TenantMiddleware {
	return &TenantMiddleware{pool: pool}
}

// ExtractTenant extracts tenant code from query param, header, or request body,
// then sets the search_path on the database connection for the duration of the request.
func (m *TenantMiddleware) ExtractTenant() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Try query param first
		tenant := c.Query("tenant")

		// Fallback to header
		if tenant == "" {
			tenant = c.Get("X-Tenant-Code")
		}

		// Fallback to request body (peek without consuming)
		if tenant == "" && len(c.Request().Body()) > 0 {
			var body struct {
				Tenant string `json:"tenant"`
			}
			if err := json.Unmarshal(c.Request().Body(), &body); err == nil && body.Tenant != "" {
				tenant = body.Tenant
			}
		}

		if tenant != "" {
			// Store in Fiber locals (accessible via c.Locals)
			c.Locals(TenantContextKey, tenant)
			log.Printf("[TenantMiddleware] Stored tenant='%s' in c.Locals", tenant)

			// Store tenant in Go context for propagation to repositories
			// Use the SAME type as in db package to ensure key matching
			ctx := context.WithValue(c.Context(), db.TenantContextKey{}, tenant)
			c.SetUserContext(ctx)
			log.Printf("[TenantMiddleware] Set c.SetUserContext with tenant='%s'", tenant)

			// Acquire a connection and set search_path for this request.
			if err := db.SetTenantSchema(c.Context(), m.pool, tenant); err != nil {
				return c.Status(500).JSON(fiber.Map{
					"success": false,
					"error": fiber.Map{
						"code":    "DB_ERROR",
						"message": fmt.Sprintf("Failed to set tenant schema: %v", err),
					},
				})
			}
		}

		// Continue to the next handler, then release the connection
		defer db.ReleaseTenantConn(c.Context())

		return c.Next()
	}
}

// GetTenant retrieves the tenant code from Fiber context
func GetTenant(c *fiber.Ctx) string {
	if tenant, ok := c.Locals(TenantContextKey).(string); ok {
		return tenant
	}
	return ""
}
