package db

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	infraDB "github.com/nova/backend/internal/infrastructure/db"

	"github.com/nova/backend/internal/domain/parts"
)

type pgPartRepository struct {
	pool *pgxpool.Pool
}

func NewPgPartRepository(pool *pgxpool.Pool) *pgPartRepository {
	return &pgPartRepository{pool: pool}
}

func (r *pgPartRepository) FindByID(ctx context.Context, id string) (*parts.Part, error) {
	var p *parts.Part
	query := `
		SELECT par_id, par_code, par_desc, par_notused, par_org, par_tenant_id,
		       par_created_at, par_updated_at, par_created_by, par_updated_by
		FROM eamparts WHERE par_id = $1`

	err := infraDB.RunInTenantTx(ctx, r.pool, func(tx pgx.Tx) error {
		var part parts.Part
		err := tx.QueryRow(ctx, query, id).Scan(
			&part.ID, &part.Code, &part.Desc, &part.NotUsed, &part.Org, &part.TenantID,
			&part.CreatedAt, &part.UpdatedAt, &part.CreatedBy, &part.UpdatedBy,
		)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		if err != nil {
			return err
		}
		p = &part
		return nil
	})

	return p, err
}

func (r *pgPartRepository) FindByCode(ctx context.Context, code string) (*parts.Part, error) {
	var p *parts.Part
	query := `
		SELECT par_id, par_code, par_desc, par_notused, par_org, par_tenant_id,
		       par_created_at, par_updated_at, par_created_by, par_updated_by
		FROM eamparts WHERE par_code = $1`

	err := infraDB.RunInTenantTx(ctx, r.pool, func(tx pgx.Tx) error {
		var part parts.Part
		err := tx.QueryRow(ctx, query, code).Scan(
			&part.ID, &part.Code, &part.Desc, &part.NotUsed, &part.Org, &part.TenantID,
			&part.CreatedAt, &part.UpdatedAt, &part.CreatedBy, &part.UpdatedBy,
		)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		if err != nil {
			return err
		}
		p = &part
		return nil
	})

	return p, err
}

func (r *pgPartRepository) FindAll(ctx context.Context, tenantID string, org string, limit, offset int) ([]*parts.Part, int, error) {
	var total int
	var result []*parts.Part

	err := infraDB.RunInTenantTx(ctx, r.pool, func(tx pgx.Tx) error {
		countQuery := `SELECT COUNT(*) FROM eamparts WHERE par_tenant_id = $1`
		args := []interface{}{tenantID}

		if org != "" {
			countQuery += ` AND par_org = $2`
			args = append(args, org)
		}

		if err := tx.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
			return err
		}

		query := `
			SELECT par_id, par_code, par_desc, par_notused, par_org, par_tenant_id,
			       par_created_at, par_updated_at, par_created_by, par_updated_by
			FROM eamparts WHERE par_tenant_id = $1`

		if org != "" {
			query += ` AND par_org = $2`
		}
		query += ` ORDER BY par_code ASC LIMIT $3 OFFSET $4`
		args = append(args, limit, offset)

		rows, err := tx.Query(ctx, query, args...)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var part parts.Part
			err := rows.Scan(
				&part.ID, &part.Code, &part.Desc, &part.NotUsed, &part.Org, &part.TenantID,
				&part.CreatedAt, &part.UpdatedAt, &part.CreatedBy, &part.UpdatedBy,
			)
			if err != nil {
				return err
			}
			result = append(result, &part)
		}
		return rows.Err()
	})

	return result, total, err
}

func (r *pgPartRepository) FindByOrg(ctx context.Context, org string) ([]*parts.Part, error) {
	var result []*parts.Part
	query := `
		SELECT par_id, par_code, par_desc, par_notused, par_org, par_tenant_id,
		       par_created_at, par_updated_at, par_created_by, par_updated_by
		FROM eamparts WHERE par_org = $1 AND par_notused IS NULL
		ORDER BY par_code ASC`

	err := infraDB.RunInTenantTx(ctx, r.pool, func(tx pgx.Tx) error {
		rows, err := tx.Query(ctx, query, org)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var part parts.Part
			err := rows.Scan(
				&part.ID, &part.Code, &part.Desc, &part.NotUsed, &part.Org, &part.TenantID,
				&part.CreatedAt, &part.UpdatedAt, &part.CreatedBy, &part.UpdatedBy,
			)
			if err != nil {
				return err
			}
			result = append(result, &part)
		}
		return rows.Err()
	})

	return result, err
}

func (r *pgPartRepository) Create(ctx context.Context, p *parts.Part) error {
	query := `
		INSERT INTO eamparts (par_id, par_code, par_desc, par_notused, par_org, 
		                      par_tenant_id, par_created_at, par_updated_at, 
		                      par_created_by, par_updated_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`

	return infraDB.RunInTenantTx(ctx, r.pool, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, query,
			p.ID, p.Code, p.Desc, p.NotUsed, p.Org, p.TenantID,
			p.CreatedAt, p.UpdatedAt, p.CreatedBy, p.UpdatedBy,
		)
		return err
	})
}

func (r *pgPartRepository) Update(ctx context.Context, p *parts.Part) error {
	query := `
		UPDATE eamparts 
		SET par_desc = $2, par_notused = $3, par_updated_at = $4, par_updated_by = $5
		WHERE par_id = $1`

	return infraDB.RunInTenantTx(ctx, r.pool, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, query, p.ID, p.Desc, p.NotUsed, p.UpdatedAt, p.UpdatedBy)
		return err
	})
}

func (r *pgPartRepository) Delete(ctx context.Context, id string) error {
	query := `UPDATE eamparts SET par_notused = '+' WHERE par_id = $1`
	return infraDB.RunInTenantTx(ctx, r.pool, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, query, id)
		return err
	})
}
