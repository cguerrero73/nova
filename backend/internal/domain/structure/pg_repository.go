package structure

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type pgStructureRepository struct {
	pool *pgxpool.Pool
}

func NewPgStructureRepository(pool *pgxpool.Pool) *pgStructureRepository {
	return &pgStructureRepository{pool: pool}
}

func (r *pgStructureRepository) FindByID(ctx context.Context, id string) (*Structure, error) {
	query := `
		SELECT sct_id, sct_parent_code, sct_parent_org, sct_child_code, sct_child_org,
		       sct_cost, sct_meter, sct_tenant_id, sct_created_at, sct_updated_at
		FROM eamstructure WHERE sct_id = $1`

	var s Structure
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&s.ID, &s.ParentCode, &s.ParentOrg, &s.ChildCode, &s.ChildOrg,
		&s.Cost, &s.Meter, &s.TenantID, &s.CreatedAt, &s.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return &s, err
}

func (r *pgStructureRepository) FindByParent(ctx context.Context, parentCode, parentOrg string) ([]*Structure, error) {
	query := `
		SELECT sct_id, sct_parent_code, sct_parent_org, sct_child_code, sct_child_org,
		       sct_cost, sct_meter, sct_tenant_id, sct_created_at, sct_updated_at
		FROM eamstructure 
		WHERE sct_parent_code = $1 AND sct_parent_org = $2`

	rows, err := r.pool.Query(ctx, query, parentCode, parentOrg)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var structures []*Structure
	for rows.Next() {
		var s Structure
		err := rows.Scan(
			&s.ID, &s.ParentCode, &s.ParentOrg, &s.ChildCode, &s.ChildOrg,
			&s.Cost, &s.Meter, &s.TenantID, &s.CreatedAt, &s.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		structures = append(structures, &s)
	}

	return structures, nil
}

func (r *pgStructureRepository) FindByChild(ctx context.Context, childCode, childOrg string) ([]*Structure, error) {
	query := `
		SELECT sct_id, sct_parent_code, sct_parent_org, sct_child_code, sct_child_org,
		       sct_cost, sct_meter, sct_tenant_id, sct_created_at, sct_updated_at
		FROM eamstructure 
		WHERE sct_child_code = $1 AND sct_child_org = $2`

	rows, err := r.pool.Query(ctx, query, childCode, childOrg)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var structures []*Structure
	for rows.Next() {
		var s Structure
		err := rows.Scan(
			&s.ID, &s.ParentCode, &s.ParentOrg, &s.ChildCode, &s.ChildOrg,
			&s.Cost, &s.Meter, &s.TenantID, &s.CreatedAt, &s.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		structures = append(structures, &s)
	}

	return structures, nil
}

func (r *pgStructureRepository) FindAll(ctx context.Context, tenantID string) ([]*Structure, error) {
	query := `
		SELECT sct_id, sct_parent_code, sct_parent_org, sct_child_code, sct_child_org,
		       sct_cost, sct_meter, sct_tenant_id, sct_created_at, sct_updated_at
		FROM eamstructure WHERE sct_tenant_id = $1`

	rows, err := r.pool.Query(ctx, query, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var structures []*Structure
	for rows.Next() {
		var s Structure
		err := rows.Scan(
			&s.ID, &s.ParentCode, &s.ParentOrg, &s.ChildCode, &s.ChildOrg,
			&s.Cost, &s.Meter, &s.TenantID, &s.CreatedAt, &s.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		structures = append(structures, &s)
	}

	return structures, nil
}

func (r *pgStructureRepository) Create(ctx context.Context, s *Structure) error {
	query := `
		INSERT INTO eamstructure (sct_id, sct_parent_code, sct_parent_org, sct_child_code,
		                          sct_child_org, sct_cost, sct_meter, sct_tenant_id, 
		                          sct_created_at, sct_updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`

	_, err := r.pool.Exec(ctx, query,
		s.ID, s.ParentCode, s.ParentOrg, s.ChildCode, s.ChildOrg,
		s.Cost, s.Meter, s.TenantID, s.CreatedAt, s.UpdatedAt,
	)
	return err
}

func (r *pgStructureRepository) Update(ctx context.Context, s *Structure) error {
	query := `
		UPDATE eamstructure 
		SET sct_cost = $2, sct_meter = $3, sct_updated_at = $4
		WHERE sct_id = $1`

	_, err := r.pool.Exec(ctx, query, s.ID, s.Cost, s.Meter, s.UpdatedAt)
	return err
}

func (r *pgStructureRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM eamstructure WHERE sct_id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	return err
}

func (r *pgStructureRepository) DeleteByParentChild(ctx context.Context, parentCode, parentOrg, childCode, childOrg string) error {
	query := `
		DELETE FROM eamstructure 
		WHERE sct_parent_code = $1 AND sct_parent_org = $2 
		  AND sct_child_code = $3 AND sct_child_org = $4`
	_, err := r.pool.Exec(ctx, query, parentCode, parentOrg, childCode, childOrg)
	return err
}
