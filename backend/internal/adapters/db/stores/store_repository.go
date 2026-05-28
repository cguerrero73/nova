package db

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/nova/backend/internal/domain/stores"
)

type pgStoreRepository struct {
	pool *pgxpool.Pool
}

func NewPgStoreRepository(pool *pgxpool.Pool) *pgStoreRepository {
	return &pgStoreRepository{pool: pool}
}

func (r *pgStoreRepository) FindByID(ctx context.Context, id string) (*stores.Store, error) {
	query := `
		SELECT str_id, str_code, str_name, str_desc, str_org, str_notused,
		       str_tenant_id, str_created_at, str_updated_at
		FROM eamstores WHERE str_id = $1`

	var s stores.Store
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&s.ID, &s.Code, &s.Name, &s.Desc, &s.Org, &s.NotUsed,
		&s.TenantID, &s.CreatedAt, &s.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return &s, err
}

func (r *pgStoreRepository) FindByCode(ctx context.Context, code string) (*stores.Store, error) {
	query := `
		SELECT str_id, str_code, str_name, str_desc, str_org, str_notused,
		       str_tenant_id, str_created_at, str_updated_at
		FROM eamstores WHERE str_code = $1`

	var s stores.Store
	err := r.pool.QueryRow(ctx, query, code).Scan(
		&s.ID, &s.Code, &s.Name, &s.Desc, &s.Org, &s.NotUsed,
		&s.TenantID, &s.CreatedAt, &s.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return &s, err
}

func (r *pgStoreRepository) FindAll(ctx context.Context, tenantID string, org string) ([]*stores.Store, error) {
	query := `
		SELECT str_id, str_code, str_name, str_desc, str_org, str_notused,
		       str_tenant_id, str_created_at, str_updated_at
		FROM eamstores WHERE str_tenant_id = $1`
	args := []interface{}{tenantID}

	if org != "" {
		query += ` AND str_org = $2`
		args = append(args, org)
	}
	query += ` ORDER BY str_code ASC`

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*stores.Store
	for rows.Next() {
		var s stores.Store
		err := rows.Scan(
			&s.ID, &s.Code, &s.Name, &s.Desc, &s.Org, &s.NotUsed,
			&s.TenantID, &s.CreatedAt, &s.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		result = append(result, &s)
	}

	return result, nil
}

func (r *pgStoreRepository) FindByOrg(ctx context.Context, org string) ([]*stores.Store, error) {
	return r.FindAll(ctx, "", org)
}

func (r *pgStoreRepository) Create(ctx context.Context, s *stores.Store) error {
	query := `
		INSERT INTO eamstores (str_id, str_code, str_name, str_desc, str_org,
		                        str_notused, str_tenant_id, str_created_at, str_updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	_, err := r.pool.Exec(ctx, query,
		s.ID, s.Code, s.Name, s.Desc, s.Org, s.NotUsed,
		s.TenantID, s.CreatedAt, s.UpdatedAt,
	)
	return err
}

func (r *pgStoreRepository) Update(ctx context.Context, s *stores.Store) error {
	query := `
		UPDATE eamstores 
		SET str_name = $2, str_desc = $3, str_notused = $4, str_updated_at = $5
		WHERE str_id = $1`

	_, err := r.pool.Exec(ctx, query, s.ID, s.Name, s.Desc, s.NotUsed, s.UpdatedAt)
	return err
}

func (r *pgStoreRepository) Delete(ctx context.Context, id string) error {
	query := `UPDATE eamstores SET str_notused = '+' WHERE str_id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	return err
}

type pgBinRepository struct {
	pool *pgxpool.Pool
}

func NewPgBinRepository(pool *pgxpool.Pool) *pgBinRepository {
	return &pgBinRepository{pool: pool}
}

func (r *pgBinRepository) FindByID(ctx context.Context, id string) (*stores.Bin, error) {
	query := `
		SELECT bin_id, bin_code, bin_desc, bin_org, bin_notused,
		       bin_tenant_id, bin_created_at, bin_updated_at
		FROM eambins WHERE bin_id = $1`

	var b stores.Bin
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&b.ID, &b.Code, &b.Desc, &b.Org, &b.NotUsed,
		&b.TenantID, &b.CreatedAt, &b.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return &b, err
}

func (r *pgBinRepository) FindByCode(ctx context.Context, code, org string) (*stores.Bin, error) {
	query := `
		SELECT bin_id, bin_code, bin_desc, bin_org, bin_notused,
		       bin_tenant_id, bin_created_at, bin_updated_at
		FROM eambins WHERE bin_code = $1 AND bin_org = $2`

	var b stores.Bin
	err := r.pool.QueryRow(ctx, query, code, org).Scan(
		&b.ID, &b.Code, &b.Desc, &b.Org, &b.NotUsed,
		&b.TenantID, &b.CreatedAt, &b.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return &b, err
}

func (r *pgBinRepository) FindAll(ctx context.Context, tenantID string, org string) ([]*stores.Bin, error) {
	query := `
		SELECT bin_id, bin_code, bin_desc, bin_org, bin_notused,
		       bin_tenant_id, bin_created_at, bin_updated_at
		FROM eambins WHERE bin_tenant_id = $1`
	args := []interface{}{tenantID}

	if org != "" {
		query += ` AND bin_org = $2`
		args = append(args, org)
	}
	query += ` ORDER BY bin_code ASC`

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*stores.Bin
	for rows.Next() {
		var b stores.Bin
		err := rows.Scan(
			&b.ID, &b.Code, &b.Desc, &b.Org, &b.NotUsed,
			&b.TenantID, &b.CreatedAt, &b.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		result = append(result, &b)
	}

	return result, nil
}

func (r *pgBinRepository) FindByOrg(ctx context.Context, org string) ([]*stores.Bin, error) {
	return r.FindAll(ctx, "", org)
}

func (r *pgBinRepository) Create(ctx context.Context, b *stores.Bin) error {
	query := `
		INSERT INTO eambins (bin_id, bin_code, bin_desc, bin_org,
		                      bin_notused, bin_tenant_id, bin_created_at, bin_updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err := r.pool.Exec(ctx, query,
		b.ID, b.Code, b.Desc, b.Org, b.NotUsed,
		b.TenantID, b.CreatedAt, b.UpdatedAt,
	)
	return err
}

func (r *pgBinRepository) Update(ctx context.Context, b *stores.Bin) error {
	query := `
		UPDATE eambins 
		SET bin_desc = $2, bin_notused = $3, bin_updated_at = $4
		WHERE bin_id = $1`

	_, err := r.pool.Exec(ctx, query, b.ID, b.Desc, b.NotUsed, b.UpdatedAt)
	return err
}

func (r *pgBinRepository) Delete(ctx context.Context, id string) error {
	query := `UPDATE eambins SET bin_notused = '+' WHERE bin_id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	return err
}
