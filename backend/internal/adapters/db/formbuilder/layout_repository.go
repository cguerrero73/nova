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

type pgLayoutRepository struct {
	pool *pgxpool.Pool
}

func NewPgLayoutRepository(pool *pgxpool.Pool) formbuilder.LayoutRepository {
	return &pgLayoutRepository{pool: pool}
}

func (r *pgLayoutRepository) Create(ctx context.Context, layout *formbuilder.Layout) error {
	return infraDB.RunInTenantTx(ctx, r.pool, func(tx pgx.Tx) error {
		err := tx.QueryRow(ctx, `
			INSERT INTO eamform_layouts (fl_form_id, fl_name, fl_display_name, fl_description, fl_status, fl_created_by, fl_created_at, fl_updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			RETURNING fl_id
		`, layout.FormID, layout.Name, layout.DisplayName, layout.Description, layout.Status, layout.CreatedBy, layout.CreatedAt, layout.UpdatedAt).Scan(&layout.ID)
		if err != nil {
			return fmt.Errorf("inserting layout: %w", err)
		}
		return nil
	})
}

func (r *pgLayoutRepository) FindByID(ctx context.Context, id int64) (*formbuilder.Layout, error) {
	var result *formbuilder.Layout
	err := infraDB.RunInTenantTx(ctx, r.pool, func(tx pgx.Tx) error {
		var l formbuilder.Layout
		err := tx.QueryRow(ctx, `
			SELECT fl_id, fl_form_id, fl_name, fl_display_name, COALESCE(fl_description, ''),
			       fl_status, fl_draft_version_id, fl_published_version_id,
			       COALESCE(fl_created_by, ''), fl_created_at, fl_updated_at
			FROM eamform_layouts
			WHERE fl_id = $1
		`, id).Scan(&l.ID, &l.FormID, &l.Name, &l.DisplayName, &l.Description,
			&l.Status, &l.DraftVersionID, &l.PublishedVersionID, &l.CreatedBy, &l.CreatedAt, &l.UpdatedAt)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		if err != nil {
			return fmt.Errorf("scanning layout: %w", err)
		}
		result = &l
		return nil
	})
	return result, err
}

func (r *pgLayoutRepository) FindByFormAndName(ctx context.Context, formID int64, name string) (*formbuilder.Layout, error) {
	var result *formbuilder.Layout
	err := infraDB.RunInTenantTx(ctx, r.pool, func(tx pgx.Tx) error {
		var l formbuilder.Layout
		err := tx.QueryRow(ctx, `
			SELECT fl_id, fl_form_id, fl_name, fl_display_name, COALESCE(fl_description, ''),
			       fl_status, fl_draft_version_id, fl_published_version_id,
			       COALESCE(fl_created_by, ''), fl_created_at, fl_updated_at
			FROM eamform_layouts
			WHERE fl_form_id = $1 AND fl_name = $2
		`, formID, name).Scan(&l.ID, &l.FormID, &l.Name, &l.DisplayName, &l.Description,
			&l.Status, &l.DraftVersionID, &l.PublishedVersionID, &l.CreatedBy, &l.CreatedAt, &l.UpdatedAt)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		if err != nil {
			return fmt.Errorf("scanning layout: %w", err)
		}
		result = &l
		return nil
	})
	return result, err
}

func (r *pgLayoutRepository) FindByName(ctx context.Context, name string) (*formbuilder.Layout, error) {
	var result *formbuilder.Layout
	err := infraDB.RunInTenantTx(ctx, r.pool, func(tx pgx.Tx) error {
		var l formbuilder.Layout
		err := tx.QueryRow(ctx, `
			SELECT fl_id, fl_form_id, fl_name, fl_display_name, COALESCE(fl_description, ''),
			       fl_status, fl_draft_version_id, fl_published_version_id,
			       COALESCE(fl_created_by, ''), fl_created_at, fl_updated_at
			FROM eamform_layouts
			WHERE fl_name = $1
			LIMIT 1
		`, name).Scan(&l.ID, &l.FormID, &l.Name, &l.DisplayName, &l.Description,
			&l.Status, &l.DraftVersionID, &l.PublishedVersionID, &l.CreatedBy, &l.CreatedAt, &l.UpdatedAt)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		if err != nil {
			return fmt.Errorf("scanning layout: %w", err)
		}
		result = &l
		return nil
	})
	return result, err
}

func (r *pgLayoutRepository) ListByFormID(ctx context.Context, formID int64) ([]*formbuilder.Layout, error) {
	var result []*formbuilder.Layout
	err := infraDB.RunInTenantTx(ctx, r.pool, func(tx pgx.Tx) error {
		rows, err := tx.Query(ctx, `
			SELECT fl_id, fl_form_id, fl_name, fl_display_name, COALESCE(fl_description, ''),
			       fl_status, fl_draft_version_id, fl_published_version_id,
			       COALESCE(fl_created_by, ''), fl_created_at, fl_updated_at
			FROM eamform_layouts
			WHERE fl_form_id = $1
			ORDER BY fl_name ASC
		`, formID)
		if err != nil {
			return fmt.Errorf("querying layouts: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var l formbuilder.Layout
			if err := rows.Scan(&l.ID, &l.FormID, &l.Name, &l.DisplayName, &l.Description,
				&l.Status, &l.DraftVersionID, &l.PublishedVersionID, &l.CreatedBy, &l.CreatedAt, &l.UpdatedAt); err != nil {
				return fmt.Errorf("scanning layout row: %w", err)
			}
			result = append(result, &l)
		}
		return rows.Err()
	})
	return result, err
}

func (r *pgLayoutRepository) UpdatePointers(ctx context.Context, layoutID int64, draftVersionID, publishedVersionID *int64) error {
	return infraDB.RunInTenantTx(ctx, r.pool, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `
			UPDATE eamform_layouts
			SET fl_draft_version_id = $2, fl_published_version_id = $3, fl_updated_at = now()
			WHERE fl_id = $1
		`, layoutID, draftVersionID, publishedVersionID)
		return err
	})
}

func (r *pgLayoutRepository) Archive(ctx context.Context, layoutID int64) error {
	return infraDB.RunInTenantTx(ctx, r.pool, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `
			UPDATE eamform_layouts
			SET fl_status = 'archived', fl_updated_at = now()
			WHERE fl_id = $1
		`, layoutID)
		return err
	})
}

func (r *pgLayoutRepository) ArchiveByFormID(ctx context.Context, formID int64) error {
	return infraDB.RunInTenantTx(ctx, r.pool, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `
			UPDATE eamform_layouts
			SET fl_status = 'archived', fl_updated_at = now()
			WHERE fl_form_id = $1 AND fl_status = 'active'
		`, formID)
		return err
	})
}
