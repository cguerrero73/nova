package parts

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type pgPartRepository struct {
	pool *pgxpool.Pool
}

func NewPgPartRepository(pool *pgxpool.Pool) *pgPartRepository {
	return &pgPartRepository{pool: pool}
}

func (r *pgPartRepository) FindByID(ctx context.Context, id string) (*Part, error) {
	query := `
		SELECT par_id, par_code, par_desc, par_notused, par_org, par_tenant_id,
		       par_created_at, par_updated_at, par_created_by, par_updated_by
		FROM eamparts WHERE par_id = $1`

	var p Part
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&p.ID, &p.Code, &p.Desc, &p.NotUsed, &p.Org, &p.TenantID,
		&p.CreatedAt, &p.UpdatedAt, &p.CreatedBy, &p.UpdatedBy,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return &p, err
}

func (r *pgPartRepository) FindByCode(ctx context.Context, code string) (*Part, error) {
	query := `
		SELECT par_id, par_code, par_desc, par_notused, par_org, par_tenant_id,
		       par_created_at, par_updated_at, par_created_by, par_updated_by
		FROM eamparts WHERE par_code = $1`

	var p Part
	err := r.pool.QueryRow(ctx, query, code).Scan(
		&p.ID, &p.Code, &p.Desc, &p.NotUsed, &p.Org, &p.TenantID,
		&p.CreatedAt, &p.UpdatedAt, &p.CreatedBy, &p.UpdatedBy,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return &p, err
}

func (r *pgPartRepository) FindAll(ctx context.Context, tenantID string, org string, limit, offset int) ([]*Part, int, error) {
	countQuery := `SELECT COUNT(*) FROM eamparts WHERE par_tenant_id = $1`
	args := []interface{}{tenantID}

	if org != "" {
		countQuery += ` AND par_org = $2`
		args = append(args, org)
	}

	var total int
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
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

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var parts []*Part
	for rows.Next() {
		var p Part
		err := rows.Scan(
			&p.ID, &p.Code, &p.Desc, &p.NotUsed, &p.Org, &p.TenantID,
			&p.CreatedAt, &p.UpdatedAt, &p.CreatedBy, &p.UpdatedBy,
		)
		if err != nil {
			return nil, 0, err
		}
		parts = append(parts, &p)
	}

	return parts, total, nil
}

func (r *pgPartRepository) FindByOrg(ctx context.Context, org string) ([]*Part, error) {
	query := `
		SELECT par_id, par_code, par_desc, par_notused, par_org, par_tenant_id,
		       par_created_at, par_updated_at, par_created_by, par_updated_by
		FROM eamparts WHERE par_org = $1 AND par_notused IS NULL
		ORDER BY par_code ASC`

	rows, err := r.pool.Query(ctx, query, org)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var parts []*Part
	for rows.Next() {
		var p Part
		err := rows.Scan(
			&p.ID, &p.Code, &p.Desc, &p.NotUsed, &p.Org, &p.TenantID,
			&p.CreatedAt, &p.UpdatedAt, &p.CreatedBy, &p.UpdatedBy,
		)
		if err != nil {
			return nil, err
		}
		parts = append(parts, &p)
	}

	return parts, nil
}

func (r *pgPartRepository) Create(ctx context.Context, p *Part) error {
	query := `
		INSERT INTO eamparts (par_id, par_code, par_desc, par_notused, par_org, 
		                      par_tenant_id, par_created_at, par_updated_at, 
		                      par_created_by, par_updated_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`

	_, err := r.pool.Exec(ctx, query,
		p.ID, p.Code, p.Desc, p.NotUsed, p.Org, p.TenantID,
		p.CreatedAt, p.UpdatedAt, p.CreatedBy, p.UpdatedBy,
	)
	return err
}

func (r *pgPartRepository) Update(ctx context.Context, p *Part) error {
	query := `
		UPDATE eamparts 
		SET par_desc = $2, par_notused = $3, par_updated_at = $4, par_updated_by = $5
		WHERE par_id = $1`

	_, err := r.pool.Exec(ctx, query, p.ID, p.Desc, p.NotUsed, p.UpdatedAt, p.UpdatedBy)
	return err
}

func (r *pgPartRepository) Delete(ctx context.Context, id string) error {
	query := `UPDATE eamparts SET par_notused = '+' WHERE par_id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	return err
}
