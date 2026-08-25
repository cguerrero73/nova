package roles

import "context"

// Repository defines the operations for loading roles and their permissions.
type Repository interface {
	// FindByCode loads a role by its code with permissions populated.
	FindByCode(ctx context.Context, code string) (*Role, error)

	// FindActiveRoleForUser returns the active role code for a user session.
	// If ses_active_role is set, it returns that. Otherwise, it falls back to
	// the user's default role from eamuser_organizations.
	FindActiveRoleForUser(ctx context.Context, userCode string) (string, error)

	// UpdateActiveRole updates the active role in the session.
	UpdateActiveRole(ctx context.Context, userCode string, roleCode string) error
}
