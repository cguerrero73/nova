package db

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/nova/backend/internal/domain/organizations"
)

type PgOrganizationRepository struct {
	pool *pgxpool.Pool
}

func NewPgOrganizationRepository(pool *pgxpool.Pool) *PgOrganizationRepository {
	return &PgOrganizationRepository{pool: pool}
}

func (r *PgOrganizationRepository) FindByID(ctx context.Context, id string) (*organizations.Organization, error) {
	query := `
		SELECT org_id, org_code, org_name, org_common, org_notused, 
		       org_tenant_id, org_created_at, org_updated_at
		FROM eamorganizations WHERE org_id = $1`

	var org organizations.Organization
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&org.ID, &org.Code, &org.Name, &org.Common, &org.NotUsed,
		&org.TenantID, &org.CreatedAt, &org.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return &org, err
}

func (r *PgOrganizationRepository) FindByCode(ctx context.Context, code string) (*organizations.Organization, error) {
	query := `
		SELECT org_id, org_code, org_name, org_common, org_notused, 
		       org_tenant_id, org_created_at, org_updated_at
		FROM eamorganizations WHERE org_code = $1`

	var org organizations.Organization
	err := r.pool.QueryRow(ctx, query, code).Scan(
		&org.ID, &org.Code, &org.Name, &org.Common, &org.NotUsed,
		&org.TenantID, &org.CreatedAt, &org.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return &org, err
}

func (r *PgOrganizationRepository) FindAll(ctx context.Context, tenantID string) ([]*organizations.Organization, error) {
	query := `
		SELECT org_id, org_code, org_name, org_common, org_notused, 
		       org_tenant_id, org_created_at, org_updated_at
		FROM eamorganizations 
		WHERE org_tenant_id = $1
		ORDER BY org_created_at ASC`

	rows, err := r.pool.Query(ctx, query, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*organizations.Organization
	for rows.Next() {
		var org organizations.Organization
		err := rows.Scan(
			&org.ID, &org.Code, &org.Name, &org.Common, &org.NotUsed,
			&org.TenantID, &org.CreatedAt, &org.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		result = append(result, &org)
	}

	return result, nil
}

func (r *PgOrganizationRepository) FindCommon(ctx context.Context, tenantID string) (*organizations.Organization, error) {
	query := `
		SELECT org_id, org_code, org_name, org_common, org_notused, 
		       org_tenant_id, org_created_at, org_updated_at
		FROM eamorganizations 
		WHERE org_tenant_id = $1 AND org_code = '*'`

	var org organizations.Organization
	err := r.pool.QueryRow(ctx, query, tenantID).Scan(
		&org.ID, &org.Code, &org.Name, &org.Common, &org.NotUsed,
		&org.TenantID, &org.CreatedAt, &org.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return &org, err
}

func (r *PgOrganizationRepository) Create(ctx context.Context, org *organizations.Organization) error {
	query := `
		INSERT INTO eamorganizations (org_id, org_code, org_name, org_common, 
		                              org_notused, org_tenant_id, org_created_at, org_updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err := r.pool.Exec(ctx, query,
		org.ID, org.Code, org.Name, org.Common, org.NotUsed,
		org.TenantID, org.CreatedAt, org.UpdatedAt,
	)
	return err
}

func (r *PgOrganizationRepository) Update(ctx context.Context, org *organizations.Organization) error {
	query := `
		UPDATE eamorganizations 
		SET org_name = $2, org_common = $3, org_notused = $4, org_updated_at = $5
		WHERE org_id = $1`

	_, err := r.pool.Exec(ctx, query,
		org.ID, org.Name, org.Common, org.NotUsed, org.UpdatedAt,
	)
	return err
}

func (r *PgOrganizationRepository) Delete(ctx context.Context, id string) error {
	query := `UPDATE eamorganizations SET org_notused = '+' WHERE org_id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	return err
}
