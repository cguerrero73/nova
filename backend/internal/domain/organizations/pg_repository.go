package organizations

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type pgOrganizationRepository struct {
	pool *pgxpool.Pool
}

func NewPgOrganizationRepository(pool *pgxpool.Pool) *pgOrganizationRepository {
	return &pgOrganizationRepository{pool: pool}
}

func (r *pgOrganizationRepository) FindByID(ctx context.Context, id string) (*Organization, error) {
	query := `
		SELECT org_id, org_code, org_name, org_common, org_notused, 
		       org_tenant_id, org_created_at, org_updated_at
		FROM eamorganizations WHERE org_id = $1`

	var org Organization
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&org.ID, &org.Code, &org.Name, &org.Common, &org.NotUsed,
		&org.TenantID, &org.CreatedAt, &org.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return &org, err
}

func (r *pgOrganizationRepository) FindByCode(ctx context.Context, code string) (*Organization, error) {
	query := `
		SELECT org_id, org_code, org_name, org_common, org_notused, 
		       org_tenant_id, org_created_at, org_updated_at
		FROM eamorganizations WHERE org_code = $1`

	var org Organization
	err := r.pool.QueryRow(ctx, query, code).Scan(
		&org.ID, &org.Code, &org.Name, &org.Common, &org.NotUsed,
		&org.TenantID, &org.CreatedAt, &org.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return &org, err
}

func (r *pgOrganizationRepository) FindAll(ctx context.Context, tenantID string) ([]*Organization, error) {
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

	var orgs []*Organization
	for rows.Next() {
		var org Organization
		err := rows.Scan(
			&org.ID, &org.Code, &org.Name, &org.Common, &org.NotUsed,
			&org.TenantID, &org.CreatedAt, &org.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		orgs = append(orgs, &org)
	}

	return orgs, nil
}

func (r *pgOrganizationRepository) FindCommon(ctx context.Context, tenantID string) (*Organization, error) {
	query := `
		SELECT org_id, org_code, org_name, org_common, org_notused, 
		       org_tenant_id, org_created_at, org_updated_at
		FROM eamorganizations 
		WHERE org_tenant_id = $1 AND org_code = '*'`

	var org Organization
	err := r.pool.QueryRow(ctx, query, tenantID).Scan(
		&org.ID, &org.Code, &org.Name, &org.Common, &org.NotUsed,
		&org.TenantID, &org.CreatedAt, &org.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return &org, err
}

func (r *pgOrganizationRepository) Create(ctx context.Context, org *Organization) error {
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

func (r *pgOrganizationRepository) Update(ctx context.Context, org *Organization) error {
	query := `
		UPDATE eamorganizations 
		SET org_name = $2, org_common = $3, org_notused = $4, org_updated_at = $5
		WHERE org_id = $1`

	_, err := r.pool.Exec(ctx, query,
		org.ID, org.Name, org.Common, org.NotUsed, org.UpdatedAt,
	)
	return err
}

func (r *pgOrganizationRepository) Delete(ctx context.Context, id string) error {
	query := `UPDATE eamorganizations SET org_notused = '+' WHERE org_id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	return err
}
