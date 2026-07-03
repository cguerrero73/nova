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

type pgLayoutVersionRepository struct {
	pool *pgxpool.Pool
}

func NewPgLayoutVersionRepository(pool *pgxpool.Pool) formbuilder.LayoutVersionRepository {
	return &pgLayoutVersionRepository{pool: pool}
}

func (r *pgLayoutVersionRepository) Create(ctx context.Context, version *formbuilder.LayoutVersion) error {
	return infraDB.RunInTenantTx(ctx, r.pool, func(tx pgx.Tx) error {
		err := tx.QueryRow(ctx, `
			INSERT INTO eamform_layout_versions (flv_layout_id, flv_version_number, flv_kind, flv_description, flv_definition, flv_created_by, flv_created_at, flv_published_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			RETURNING flv_id
		`, version.LayoutID, version.VersionNumber, version.Kind, version.Description, version.Definition, version.CreatedBy, version.CreatedAt, version.PublishedAt).Scan(&version.ID)
		if err != nil {
			return fmt.Errorf("inserting layout version: %w", err)
		}
		return nil
	})
}

func (r *pgLayoutVersionRepository) FindByID(ctx context.Context, id int64) (*formbuilder.LayoutVersion, error) {
	var result *formbuilder.LayoutVersion
	err := infraDB.RunInTenantTx(ctx, r.pool, func(tx pgx.Tx) error {
		var v formbuilder.LayoutVersion
		err := tx.QueryRow(ctx, `
			SELECT flv_id, flv_layout_id, flv_version_number, flv_kind,
			       COALESCE(flv_description, ''), flv_definition,
			       COALESCE(flv_created_by, ''), flv_created_at, flv_published_at
			FROM eamform_layout_versions
			WHERE flv_id = $1
		`, id).Scan(&v.ID, &v.LayoutID, &v.VersionNumber, &v.Kind, &v.Description, &v.Definition, &v.CreatedBy, &v.CreatedAt, &v.PublishedAt)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		if err != nil {
			return fmt.Errorf("scanning layout version: %w", err)
		}
		result = &v
		return nil
	})
	return result, err
}

func (r *pgLayoutVersionRepository) FindMaxVersionNumber(ctx context.Context, layoutID int64) (int, error) {
	var maxVer int
	err := infraDB.RunInTenantTx(ctx, r.pool, func(tx pgx.Tx) error {
		err := tx.QueryRow(ctx, `
			SELECT COALESCE(MAX(flv_version_number), 0)
			FROM eamform_layout_versions
			WHERE flv_layout_id = $1
		`, layoutID).Scan(&maxVer)
		if errors.Is(err, pgx.ErrNoRows) {
			maxVer = 0
			return nil
		}
		return err
	})
	return maxVer, err
}

func (r *pgLayoutVersionRepository) FindDraft(ctx context.Context, layoutID int64) (*formbuilder.LayoutVersion, error) {
	var result *formbuilder.LayoutVersion
	err := infraDB.RunInTenantTx(ctx, r.pool, func(tx pgx.Tx) error {
		var v formbuilder.LayoutVersion
		err := tx.QueryRow(ctx, `
			SELECT flv_id, flv_layout_id, flv_version_number, flv_kind,
			       COALESCE(flv_description, ''), flv_definition,
			       COALESCE(flv_created_by, ''), flv_created_at, flv_published_at
			FROM eamform_layout_versions
			WHERE flv_layout_id = $1 AND flv_kind = 'draft'
		`, layoutID).Scan(&v.ID, &v.LayoutID, &v.VersionNumber, &v.Kind, &v.Description, &v.Definition, &v.CreatedBy, &v.CreatedAt, &v.PublishedAt)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		if err != nil {
			return fmt.Errorf("scanning draft version: %w", err)
		}
		result = &v
		return nil
	})
	return result, err
}

func (r *pgLayoutVersionRepository) UpdateDraftDefinition(ctx context.Context, versionID int64, definition []byte) error {
	return infraDB.RunInTenantTx(ctx, r.pool, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `
			UPDATE eamform_layout_versions
			SET flv_definition = $2
			WHERE flv_id = $1 AND flv_kind = 'draft'
		`, versionID, definition)
		return err
	})
}

func (r *pgLayoutVersionRepository) ListByLayoutID(ctx context.Context, layoutID int64) ([]*formbuilder.LayoutVersion, error) {
	var result []*formbuilder.LayoutVersion
	err := infraDB.RunInTenantTx(ctx, r.pool, func(tx pgx.Tx) error {
		rows, err := tx.Query(ctx, `
			SELECT flv_id, flv_layout_id, flv_version_number, flv_kind,
			       COALESCE(flv_description, ''), flv_definition,
			       COALESCE(flv_created_by, ''), flv_created_at, flv_published_at
			FROM eamform_layout_versions
			WHERE flv_layout_id = $1
			ORDER BY flv_version_number DESC
		`, layoutID)
		if err != nil {
			return fmt.Errorf("querying versions: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var v formbuilder.LayoutVersion
			if err := rows.Scan(&v.ID, &v.LayoutID, &v.VersionNumber, &v.Kind, &v.Description, &v.Definition, &v.CreatedBy, &v.CreatedAt, &v.PublishedAt); err != nil {
				return fmt.Errorf("scanning version row: %w", err)
			}
			result = append(result, &v)
		}
		return rows.Err()
	})
	return result, err
}
