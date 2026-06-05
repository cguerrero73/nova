package db

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	infraDB "github.com/nova/backend/internal/infrastructure/db"

	"github.com/nova/backend/internal/domain/structure"
)

type PgStructureRepository struct {
	pool *pgxpool.Pool
}

func NewPgStructureRepository(pool *pgxpool.Pool) *PgStructureRepository {
	return &PgStructureRepository{pool: pool}
}

func (r *PgStructureRepository) FindByID(ctx context.Context, id string) (*structure.Structure, error) {
	var s *structure.Structure
	query := `
		SELECT sct_id, sct_parent_code, sct_parent_org, sct_child_code, sct_child_org,
		       sct_cost, sct_meter, sct_tenant_id, sct_created_at, sct_updated_at
		FROM eamstructure WHERE sct_id = $1`

	err := infraDB.RunInTenantTx(ctx, r.pool, func(tx pgx.Tx) error {
		var str structure.Structure
		err := tx.QueryRow(ctx, query, id).Scan(
			&str.ID, &str.ParentCode, &str.ParentOrg, &str.ChildCode, &str.ChildOrg,
			&str.Cost, &str.Meter, &str.TenantID, &str.CreatedAt, &str.UpdatedAt,
		)
		if err != nil {
			return err
		}
		s = &str
		return nil
	})

	return s, err
}

func (r *PgStructureRepository) FindByParent(ctx context.Context, parentCode, parentOrg string) ([]*structure.Structure, error) {
	var result []*structure.Structure
	query := `
		SELECT sct_id, sct_parent_code, sct_parent_org, sct_child_code, sct_child_org,
		       sct_cost, sct_meter, sct_tenant_id, sct_created_at, sct_updated_at
		FROM eamstructure 
		WHERE sct_parent_code = $1 AND sct_parent_org = $2`

	err := infraDB.RunInTenantTx(ctx, r.pool, func(tx pgx.Tx) error {
		rows, err := tx.Query(ctx, query, parentCode, parentOrg)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var str structure.Structure
			err := rows.Scan(
				&str.ID, &str.ParentCode, &str.ParentOrg, &str.ChildCode, &str.ChildOrg,
				&str.Cost, &str.Meter, &str.TenantID, &str.CreatedAt, &str.UpdatedAt,
			)
			if err != nil {
				return err
			}
			result = append(result, &str)
		}
		return rows.Err()
	})

	return result, err
}

func (r *PgStructureRepository) FindByChild(ctx context.Context, childCode, childOrg string) ([]*structure.Structure, error) {
	var result []*structure.Structure
	query := `
		SELECT sct_id, sct_parent_code, sct_parent_org, sct_child_code, sct_child_org,
		       sct_cost, sct_meter, sct_tenant_id, sct_created_at, sct_updated_at
		FROM eamstructure 
		WHERE sct_child_code = $1 AND sct_child_org = $2`

	err := infraDB.RunInTenantTx(ctx, r.pool, func(tx pgx.Tx) error {
		rows, err := tx.Query(ctx, query, childCode, childOrg)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var str structure.Structure
			err := rows.Scan(
				&str.ID, &str.ParentCode, &str.ParentOrg, &str.ChildCode, &str.ChildOrg,
				&str.Cost, &str.Meter, &str.TenantID, &str.CreatedAt, &str.UpdatedAt,
			)
			if err != nil {
				return err
			}
			result = append(result, &str)
		}
		return rows.Err()
	})

	return result, err
}

func (r *PgStructureRepository) FindAll(ctx context.Context, tenantID string) ([]*structure.Structure, error) {
	var result []*structure.Structure
	query := `
		SELECT sct_id, sct_parent_code, sct_parent_org, sct_child_code, sct_child_org,
		       sct_cost, sct_meter, sct_tenant_id, sct_created_at, sct_updated_at
		FROM eamstructure WHERE sct_tenant_id = $1`

	err := infraDB.RunInTenantTx(ctx, r.pool, func(tx pgx.Tx) error {
		rows, err := tx.Query(ctx, query, tenantID)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var str structure.Structure
			err := rows.Scan(
				&str.ID, &str.ParentCode, &str.ParentOrg, &str.ChildCode, &str.ChildOrg,
				&str.Cost, &str.Meter, &str.TenantID, &str.CreatedAt, &str.UpdatedAt,
			)
			if err != nil {
				return err
			}
			result = append(result, &str)
		}
		return rows.Err()
	})

	return result, err
}

func (r *PgStructureRepository) Create(ctx context.Context, s *structure.Structure) error {
	query := `
		INSERT INTO eamstructure (sct_id, sct_parent_code, sct_parent_org, sct_child_code,
		                          sct_child_org, sct_cost, sct_meter, sct_tenant_id, 
		                          sct_created_at, sct_updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`

	return infraDB.RunInTenantTx(ctx, r.pool, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, query,
			s.ID, s.ParentCode, s.ParentOrg, s.ChildCode, s.ChildOrg,
			s.Cost, s.Meter, s.TenantID, s.CreatedAt, s.UpdatedAt,
		)
		return err
	})
}

func (r *PgStructureRepository) Update(ctx context.Context, s *structure.Structure) error {
	query := `
		UPDATE eamstructure 
		SET sct_cost = $2, sct_meter = $3, sct_updated_at = $4
		WHERE sct_id = $1`

	return infraDB.RunInTenantTx(ctx, r.pool, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, query, s.ID, s.Cost, s.Meter, s.UpdatedAt)
		return err
	})
}

func (r *PgStructureRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM eamstructure WHERE sct_id = $1`
	return infraDB.RunInTenantTx(ctx, r.pool, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, query, id)
		return err
	})
}

func (r *PgStructureRepository) DeleteByParentChild(ctx context.Context, parentCode, parentOrg, childCode, childOrg string) error {
	query := `
		DELETE FROM eamstructure 
		WHERE sct_parent_code = $1 AND sct_parent_org = $2 
		  AND sct_child_code = $3 AND sct_child_org = $4`
	return infraDB.RunInTenantTx(ctx, r.pool, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, query, parentCode, parentOrg, childCode, childOrg)
		return err
	})
}
