package formbuilder

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/nova/backend/internal/domain/formbuilder"
	infraDB "github.com/nova/backend/internal/infrastructure/db"
)

type pgAssignmentRepository struct {
	pool *pgxpool.Pool
}

func NewPgAssignmentRepository(pool *pgxpool.Pool) formbuilder.AssignmentRepository {
	return &pgAssignmentRepository{pool: pool}
}

func (r *pgAssignmentRepository) Create(ctx context.Context, assignment *formbuilder.RoleAssignment) error {
	return infraDB.RunInTenantTx(ctx, r.pool, func(tx pgx.Tx) error {
		err := tx.QueryRow(ctx, `
			INSERT INTO eamform_role_assignments (fra_form_id, fra_layout_id, fra_role_name, fra_assigned_at, fra_assigned_by)
			VALUES ($1, $2, $3, $4, $5)
			RETURNING fra_id
		`, assignment.FormID, assignment.LayoutID, assignment.RoleName, assignment.AssignedAt, assignment.AssignedBy).Scan(&assignment.ID)
		if err != nil {
			return fmt.Errorf("inserting assignment: %w", err)
		}
		return nil
	})
}

func (r *pgAssignmentRepository) FindActiveByFormAndRole(ctx context.Context, formID int64, roleName string) (*formbuilder.RoleAssignment, error) {
	var result *formbuilder.RoleAssignment
	err := infraDB.RunInTenantTx(ctx, r.pool, func(tx pgx.Tx) error {
		var a formbuilder.RoleAssignment
		err := tx.QueryRow(ctx, `
			SELECT fra_id, fra_form_id, fra_layout_id, fra_role_name,
			       fra_assigned_at, fra_revoked_at, COALESCE(fra_assigned_by, '')
			FROM eamform_role_assignments
			WHERE fra_form_id = $1 AND fra_role_name = $2 AND fra_revoked_at IS NULL
		`, formID, roleName).Scan(&a.ID, &a.FormID, &a.LayoutID, &a.RoleName, &a.AssignedAt, &a.RevokedAt, &a.AssignedBy)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		if err != nil {
			return fmt.Errorf("scanning assignment: %w", err)
		}
		result = &a
		return nil
	})
	return result, err
}

func (r *pgAssignmentRepository) ListByFormID(ctx context.Context, formID int64) ([]*formbuilder.RoleAssignment, error) {
	var result []*formbuilder.RoleAssignment
	err := infraDB.RunInTenantTx(ctx, r.pool, func(tx pgx.Tx) error {
		rows, err := tx.Query(ctx, `
			SELECT fra_id, fra_form_id, fra_layout_id, fra_role_name,
			       fra_assigned_at, fra_revoked_at, COALESCE(fra_assigned_by, '')
			FROM eamform_role_assignments
			WHERE fra_form_id = $1 AND fra_revoked_at IS NULL
			ORDER BY fra_role_name ASC
		`, formID)
		if err != nil {
			return fmt.Errorf("querying assignments: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var a formbuilder.RoleAssignment
			if err := rows.Scan(&a.ID, &a.FormID, &a.LayoutID, &a.RoleName, &a.AssignedAt, &a.RevokedAt, &a.AssignedBy); err != nil {
				return fmt.Errorf("scanning assignment row: %w", err)
			}
			result = append(result, &a)
		}
		return rows.Err()
	})
	return result, err
}

func (r *pgAssignmentRepository) Revoke(ctx context.Context, assignmentID int64) error {
	return infraDB.RunInTenantTx(ctx, r.pool, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `
			UPDATE eamform_role_assignments
			SET fra_revoked_at = now()
			WHERE fra_id = $1
		`, assignmentID)
		return err
	})
}
