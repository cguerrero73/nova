package roles

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/nova/backend/internal/domain/roles"
	infraDB "github.com/nova/backend/internal/infrastructure/db"
)

// PgRoleRepository implements roles.Repository using pgx.
type PgRoleRepository struct {
	pool *pgxpool.Pool
}

// NewPgRoleRepository creates a new PgRoleRepository.
func NewPgRoleRepository(pool *pgxpool.Pool) *PgRoleRepository {
	return &PgRoleRepository{pool: pool}
}

// FindByCode loads a role by its code with permissions populated.
func (r *PgRoleRepository) FindByCode(ctx context.Context, code string) (*roles.Role, error) {
	var role *roles.Role

	err := infraDB.RunInTenantTx(ctx, r.pool, func(tx pgx.Tx) error {
		// Load role basic info
		var ro roles.Role
		var rolID int64
		var systemFlag, notUsedFlag string
		err := tx.QueryRow(ctx, `
			SELECT rol_id, rol_code, rol_desc, COALESCE(rol_system, '-'), COALESCE(rol_notused, '-'),
			       COALESCE(rol_created_at, now()), rol_updated_at
			FROM eamroles
			WHERE rol_code = $1
		`, code).Scan(
			&rolID, &ro.ID, &ro.Desc, &systemFlag, &notUsedFlag,
			&ro.CreatedAt, &ro.UpdatedAt,
		)
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("role not found: %w", err)
		}
		if err != nil {
			return fmt.Errorf("scanning role: %w", err)
		}
		ro.IsSystem = systemFlag == "+"
		if notUsedFlag != "+" {
			ro.NotUsed = &notUsedFlag
		}

		// Load permissions
		rows, err := tx.Query(ctx, `
			SELECT rpe_screen, rpe_action, rpe_allowed
			FROM eamrole_permissions
			WHERE rpe_role = $1
		`, code)
		if err != nil {
			return fmt.Errorf("querying permissions: %w", err)
		}
		defer rows.Close()

		perms := make(map[string]map[string]bool)
		for rows.Next() {
			var screen, action string
			var allowed bool
			if err := rows.Scan(&screen, &action, &allowed); err != nil {
				return fmt.Errorf("scanning permission: %w", err)
			}
			if perms[screen] == nil {
				perms[screen] = make(map[string]bool)
			}
			perms[screen][action] = allowed
		}
		if err := rows.Err(); err != nil {
			return fmt.Errorf("iterating permissions: %w", err)
		}

		ro.Permissions = perms
		role = &ro
		return nil
	})

	return role, err
}

// FindActiveRoleForUser returns the active role code for a user session.
// Falls back to the user's default role if ses_active_role is not set.
func (r *PgRoleRepository) FindActiveRoleForUser(ctx context.Context, userCode string) (string, error) {
	var roleCode string

	err := infraDB.RunInTenantTx(ctx, r.pool, func(tx pgx.Tx) error {
		// Try ses_active_role first
		err := tx.QueryRow(ctx, `
			SELECT COALESCE(ses_active_role, '')
			FROM eamsessions
			WHERE ses_user_code = $1
			  AND ses_revoked_at IS NULL
			ORDER BY ses_created_at DESC
			LIMIT 1
		`, userCode).Scan(&roleCode)

		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("querying active role: %w", err)
		}

		// Fallback to default role from user_organizations
		if roleCode == "" || errors.Is(err, pgx.ErrNoRows) {
			err = tx.QueryRow(ctx, `
				SELECT uog_role
				FROM eamuser_organizations
				WHERE uog_user = $1
				  AND COALESCE(uog_default, '-') = '+'
				LIMIT 1
			`, userCode).Scan(&roleCode)
			if errors.Is(err, pgx.ErrNoRows) {
				return fmt.Errorf("no active role found for user %s", userCode)
			}
			if err != nil {
				return fmt.Errorf("querying default role: %w", err)
			}
		}

		return nil
	})

	return roleCode, err
}

// UpdateActiveRole updates the active role in the user's most recent session.
func (r *PgRoleRepository) UpdateActiveRole(ctx context.Context, userCode string, roleCode string) error {
	return infraDB.RunInTenantTx(ctx, r.pool, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `
			UPDATE eamsessions
			SET ses_active_role = $2, ses_updated_at = now()
			WHERE ses_user_code = $1
			  AND ses_revoked_at IS NULL
			  AND ses_id = (
				  SELECT ses_id
				  FROM eamsessions
				  WHERE ses_user_code = $1
				    AND ses_revoked_at IS NULL
				  ORDER BY ses_created_at DESC
				  LIMIT 1
			  )
		`, userCode, roleCode)
		return err
	})
}
