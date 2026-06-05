package db

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	infraDB "github.com/nova/backend/internal/infrastructure/db"

	"github.com/nova/backend/internal/domain/organizations"
)

type PgOrganizationRepository struct {
	pool *pgxpool.Pool
}

func NewPgOrganizationRepository(pool *pgxpool.Pool) *PgOrganizationRepository {
	return &PgOrganizationRepository{pool: pool}
}

func (r *PgOrganizationRepository) FindByID(ctx context.Context, id string) (*organizations.Organization, error) {
	var org *organizations.Organization
	query := `
		SELECT org_id, org_code, org_name, org_common, org_notused, 
		       org_tenant_id, org_created_at, org_updated_at
		FROM eamorganizations WHERE org_id = $1`

	err := infraDB.RunInTenantTx(ctx, r.pool, func(tx pgx.Tx) error {
		var o organizations.Organization
		err := tx.QueryRow(ctx, query, id).Scan(
			&o.ID, &o.Code, &o.Name, &o.Common, &o.NotUsed,
			&o.TenantID, &o.CreatedAt, &o.UpdatedAt,
		)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		if err != nil {
			return err
		}
		org = &o
		return nil
	})

	return org, err
}

func (r *PgOrganizationRepository) FindByCode(ctx context.Context, code string) (*organizations.Organization, error) {
	var org *organizations.Organization
	query := `
		SELECT org_id, org_code, org_name, org_common, org_notused, 
		       org_tenant_id, org_created_at, org_updated_at
		FROM eamorganizations WHERE org_code = $1`

	err := infraDB.RunInTenantTx(ctx, r.pool, func(tx pgx.Tx) error {
		var o organizations.Organization
		err := tx.QueryRow(ctx, query, code).Scan(
			&o.ID, &o.Code, &o.Name, &o.Common, &o.NotUsed,
			&o.TenantID, &o.CreatedAt, &o.UpdatedAt,
		)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		if err != nil {
			return err
		}
		org = &o
		return nil
	})

	return org, err
}

func (r *PgOrganizationRepository) FindAll(ctx context.Context, tenantID string) ([]*organizations.Organization, error) {
	var result []*organizations.Organization
	query := `
		SELECT org_id, org_code, org_name, org_common, org_notused, 
		       org_tenant_id, org_created_at, org_updated_at
		FROM eamorganizations 
		WHERE org_tenant_id = $1
		ORDER BY org_created_at ASC`

	err := infraDB.RunInTenantTx(ctx, r.pool, func(tx pgx.Tx) error {
		rows, err := tx.Query(ctx, query, tenantID)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var o organizations.Organization
			err := rows.Scan(
				&o.ID, &o.Code, &o.Name, &o.Common, &o.NotUsed,
				&o.TenantID, &o.CreatedAt, &o.UpdatedAt,
			)
			if err != nil {
				return err
			}
			result = append(result, &o)
		}
		return rows.Err()
	})

	return result, err
}

func (r *PgOrganizationRepository) FindCommon(ctx context.Context, tenantID string) (*organizations.Organization, error) {
	var org *organizations.Organization
	query := `
		SELECT org_id, org_code, org_name, org_common, org_notused, 
		       org_tenant_id, org_created_at, org_updated_at
		FROM eamorganizations 
		WHERE org_tenant_id = $1 AND org_code = '*'`

	err := infraDB.RunInTenantTx(ctx, r.pool, func(tx pgx.Tx) error {
		var o organizations.Organization
		err := tx.QueryRow(ctx, query, tenantID).Scan(
			&o.ID, &o.Code, &o.Name, &o.Common, &o.NotUsed,
			&o.TenantID, &o.CreatedAt, &o.UpdatedAt,
		)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		if err != nil {
			return err
		}
		org = &o
		return nil
	})

	return org, err
}

func (r *PgOrganizationRepository) Create(ctx context.Context, org *organizations.Organization) error {
	query := `
		INSERT INTO eamorganizations (org_id, org_code, org_name, org_common, 
		                              org_notused, org_tenant_id, org_created_at, org_updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	return infraDB.RunInTenantTx(ctx, r.pool, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, query,
			org.ID, org.Code, org.Name, org.Common, org.NotUsed,
			org.TenantID, org.CreatedAt, org.UpdatedAt,
		)
		return err
	})
}

func (r *PgOrganizationRepository) Update(ctx context.Context, org *organizations.Organization) error {
	query := `
		UPDATE eamorganizations 
		SET org_name = $2, org_common = $3, org_notused = $4, org_updated_at = $5
		WHERE org_id = $1`

	return infraDB.RunInTenantTx(ctx, r.pool, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, query,
			org.ID, org.Name, org.Common, org.NotUsed, org.UpdatedAt,
		)
		return err
	})
}

func (r *PgOrganizationRepository) Delete(ctx context.Context, id string) error {
	query := `UPDATE eamorganizations SET org_notused = '+' WHERE org_id = $1`
	return infraDB.RunInTenantTx(ctx, r.pool, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, query, id)
		return err
	})
}
