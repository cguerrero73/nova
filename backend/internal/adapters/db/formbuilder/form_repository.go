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

type pgFormRepository struct {
	pool *pgxpool.Pool
}

func NewPgFormRepository(pool *pgxpool.Pool) formbuilder.FormRepository {
	return &pgFormRepository{pool: pool}
}

func (r *pgFormRepository) Create(ctx context.Context, form *formbuilder.Form) error {
	return infraDB.RunInTenantTx(ctx, r.pool, func(tx pgx.Tx) error {
		err := tx.QueryRow(ctx, `
			INSERT INTO eamform_definitions (frm_key, frm_name, frm_description, frm_status, frm_created_by, frm_created_at, frm_updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			RETURNING frm_id
		`, form.Key, form.Name, form.Description, form.Status, form.CreatedBy, form.CreatedAt, form.UpdatedAt).Scan(&form.ID)
		if err != nil {
			return fmt.Errorf("inserting form: %w", err)
		}
		return nil
	})
}

func (r *pgFormRepository) FindByKey(ctx context.Context, key string) (*formbuilder.Form, error) {
	var result *formbuilder.Form
	err := infraDB.RunInTenantTx(ctx, r.pool, func(tx pgx.Tx) error {
		var f formbuilder.Form
		err := tx.QueryRow(ctx, `
			SELECT frm_id, frm_key, frm_name, COALESCE(frm_description, ''), frm_status,
			       COALESCE(frm_created_by, ''), frm_created_at, frm_updated_at
			FROM eamform_definitions
			WHERE frm_key = $1
		`, key).Scan(&f.ID, &f.Key, &f.Name, &f.Description, &f.Status, &f.CreatedBy, &f.CreatedAt, &f.UpdatedAt)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		if err != nil {
			return fmt.Errorf("scanning form: %w", err)
		}
		result = &f
		return nil
	})
	return result, err
}

func (r *pgFormRepository) List(ctx context.Context, status string) ([]*formbuilder.Form, error) {
	var result []*formbuilder.Form
	err := infraDB.RunInTenantTx(ctx, r.pool, func(tx pgx.Tx) error {
		query := `
			SELECT frm_id, frm_key, frm_name, COALESCE(frm_description, ''), frm_status,
			       COALESCE(frm_created_by, ''), frm_created_at, frm_updated_at
			FROM eamform_definitions`
		var args []interface{}
		if status != "" {
			query += ` WHERE frm_status = $1`
			args = append(args, status)
		}
		query += ` ORDER BY frm_created_at DESC`

		rows, err := tx.Query(ctx, query, args...)
		if err != nil {
			return fmt.Errorf("querying forms: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var f formbuilder.Form
			if err := rows.Scan(&f.ID, &f.Key, &f.Name, &f.Description, &f.Status, &f.CreatedBy, &f.CreatedAt, &f.UpdatedAt); err != nil {
				return fmt.Errorf("scanning form row: %w", err)
			}
			result = append(result, &f)
		}
		return rows.Err()
	})
	return result, err
}

func (r *pgFormRepository) Archive(ctx context.Context, formID int64) error {
	return infraDB.RunInTenantTx(ctx, r.pool, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `
			UPDATE eamform_definitions
			SET frm_status = 'archived', frm_updated_at = now()
			WHERE frm_id = $1
		`, formID)
		return err
	})
}
