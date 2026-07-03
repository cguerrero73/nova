package middleware

import (
	"log"

	"github.com/gofiber/fiber/v2"

	"github.com/nova/backend/internal/domain/roles"
)

// ContextLoaderMiddleware loads the active role and its permissions into the
// request context after authentication. This enables permission checks via
// Role.HasPermission(screen, action) in downstream handlers.
type ContextLoaderMiddleware struct {
	roleRepo roles.Repository
}

// NewContextLoader creates a new ContextLoaderMiddleware.
func NewContextLoader(roleRepo roles.Repository) *ContextLoaderMiddleware {
	return &ContextLoaderMiddleware{roleRepo: roleRepo}
}

// LoadContext is the middleware handler. It:
// 1. Reads the authenticated user from c.Locals("user")
// 2. Resolves the active role code (from session or default)
// 3. Loads the full Role entity with permissions
// 4. Stores role code in c.Locals("activeRole") and Role entity in c.Locals("role")
func (m *ContextLoaderMiddleware) LoadContext() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get authenticated user claims
		claims := GetUserClaims(c)
		if claims == nil {
			// No auth — skip context loading (public route)
			return c.Next()
		}

		userCode := claims.UserCode
		if userCode == "" {
			log.Printf("[ContextLoader] WARNING: userCode is empty in JWT claims")
			return c.Next()
		}

		// Resolve active role code
		roleCode, err := m.roleRepo.FindActiveRoleForUser(c.UserContext(), userCode)
		if err != nil {
			log.Printf("[ContextLoader] WARNING: could not resolve active role for user %s: %v", userCode, err)
			// Continue without role — handlers that need permissions will fail gracefully
			return c.Next()
		}

		if roleCode == "" {
			log.Printf("[ContextLoader] WARNING: no active role found for user %s", userCode)
			return c.Next()
		}

		// Store role code in locals (conventions §3)
		c.Locals("activeRole", roleCode)

		// Load full role with permissions
		role, err := m.roleRepo.FindByCode(c.UserContext(), roleCode)
		if err != nil {
			log.Printf("[ContextLoader] WARNING: could not load role %s: %v", roleCode, err)
			return c.Next()
		}

		// Store full role entity for permission checks
		c.Locals("role", role)

		log.Printf("[ContextLoader] user=%s activeRole=%s permissions=%d screens",
			userCode, roleCode, len(role.Permissions))

		return c.Next()
	}
}

// GetActiveRole returns the active role code from the context.
func GetActiveRole(c *fiber.Ctx) string {
	if role, ok := c.Locals("activeRole").(string); ok {
		return role
	}
	return ""
}

// GetRole returns the full Role entity from the context.
func GetRole(c *fiber.Ctx) *roles.Role {
	if role, ok := c.Locals("role").(*roles.Role); ok {
		return role
	}
	return nil
}
