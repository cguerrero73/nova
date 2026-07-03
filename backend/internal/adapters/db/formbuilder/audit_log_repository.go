package formbuilder

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/nova/backend/internal/domain/formbuilder"
	infraDB "github.com/nova/backend/internal/infrastructure/db"
)

type pgAuditLogRepository struct {
	pool *pgxpool.Pool
}

func NewPgAuditLogRepository(pool *pgxpool.Pool) formbuilder.AuditLogRepository {
	return &pgAuditLogRepository{pool: pool}
}

func (r *pgAuditLogRepository) Create(ctx context.Context, entry *formbuilder.AuditEntry) error {
	return infraDB.RunInTenantTx(ctx, r.pool, func(tx pgx.Tx) error {
		err := tx.QueryRow(ctx, `
			INSERT INTO eamform_audit_log (fal_actor_user_id, fal_action, fal_entity_type, fal_entity_id, fal_metadata, fal_note, fal_created_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			RETURNING fal_id
		`, entry.ActorUserID, entry.Action, entry.EntityType, entry.EntityID, entry.Metadata, entry.Note, entry.CreatedAt).Scan(&entry.ID)
		if err != nil {
			return fmt.Errorf("inserting audit entry: %w", err)
		}
		return nil
	})
}

func (r *pgAuditLogRepository) ListByForm(ctx context.Context, formID int64, filter formbuilder.AuditFilter, limit, offset int) ([]*formbuilder.AuditEntry, int, error) {
	var total int
	var result []*formbuilder.AuditEntry

	err := infraDB.RunInTenantTx(ctx, r.pool, func(tx pgx.Tx) error {
		// Build WHERE clause for form-related audit entries.
		// Audit entries reference entities (form, layout) by entity_id.
		// We need to find all audit entries related to a form:
		// - entity_type = 'form' AND entity_id = formID
		// - entity_type = 'layout' AND entity_id IN (layouts for this form)
		// - entity_type = 'version' AND entity_id IN (versions for layouts of this form)
		// - entity_type = 'assignment' AND entity_id IN (assignments for this form)
		baseWhere := `
			(fal_entity_type = 'form' AND fal_entity_id = $1)
			OR (fal_entity_type = 'layout' AND fal_entity_id IN (SELECT fl_id FROM eamform_layouts WHERE fl_form_id = $1))
			OR (fal_entity_type = 'version' AND fal_entity_id IN (SELECT flv_id FROM eamform_layout_versions WHERE flv_layout_id IN (SELECT fl_id FROM eamform_layouts WHERE fl_form_id = $1)))
			OR (fal_entity_type = 'assignment' AND fal_entity_id IN (SELECT fra_id FROM eamform_role_assignments WHERE fra_form_id = $1))
		`

		// Build dynamic filter clause
		filterClause := ""
		var args []interface{}
		args = append(args, formID)
		paramIdx := 2

		if filter.Action != "" {
			filterClause += fmt.Sprintf(` AND fal_action = $%d`, paramIdx)
			args = append(args, filter.Action)
			paramIdx++
		}
		if filter.EntityType != "" {
			filterClause += fmt.Sprintf(` AND fal_entity_type = $%d`, paramIdx)
			args = append(args, filter.EntityType)
			paramIdx++
		}
		if filter.Actor != "" {
			filterClause += fmt.Sprintf(` AND fal_actor_user_id = $%d`, paramIdx)
			args = append(args, filter.Actor)
			paramIdx++
		}
		if filter.From != nil {
			filterClause += fmt.Sprintf(` AND fal_created_at >= $%d`, paramIdx)
			args = append(args, *filter.From)
			paramIdx++
		}
		if filter.To != nil {
			filterClause += fmt.Sprintf(` AND fal_created_at <= $%d`, paramIdx)
			args = append(args, *filter.To)
			paramIdx++
		}

		countQuery := `SELECT COUNT(*) FROM eamform_audit_log WHERE ` + baseWhere + filterClause
		if err := tx.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
			return fmt.Errorf("counting audit entries: %w", err)
		}

		dataQuery := `
			SELECT fal_id, fal_actor_user_id, fal_action, fal_entity_type, fal_entity_id,
			       COALESCE(fal_metadata, '{}'::jsonb), COALESCE(fal_note, ''), fal_created_at
			FROM eamform_audit_log
			WHERE ` + baseWhere + filterClause
		dataQuery += ` ORDER BY fal_created_at DESC`

		dataQuery += fmt.Sprintf(` LIMIT $%d OFFSET $%d`, paramIdx, paramIdx+1)
		args = append(args, limit, offset)

		rows, err := tx.Query(ctx, dataQuery, args...)
		if err != nil {
			return fmt.Errorf("querying audit entries: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var e formbuilder.AuditEntry
			if err := rows.Scan(&e.ID, &e.ActorUserID, &e.Action, &e.EntityType, &e.EntityID, &e.Metadata, &e.Note, &e.CreatedAt); err != nil {
				return fmt.Errorf("scanning audit entry: %w", err)
			}
			result = append(result, &e)
		}
		return rows.Err()
	})

	return result, total, err
}
