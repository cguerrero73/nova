package objects

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type pgObjectRepository struct {
	pool *pgxpool.Pool
}

func NewPgObjectRepository(pool *pgxpool.Pool) *pgObjectRepository {
	return &pgObjectRepository{pool: pool}
}

func (r *pgObjectRepository) FindByID(ctx context.Context, id string) (*Object, error) {
	query := `
		SELECT obj_id, obj_code, obj_type, obj_desc, obj_serial, obj_status, 
		       obj_org, obj_parent_code, obj_parent_org, obj_install_date,
		       obj_notused, obj_tenant_id, obj_created_at, obj_updated_at,
		       obj_created_by, obj_updated_by
		FROM eamobjects WHERE obj_id = $1`

	var obj Object
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&obj.ID, &obj.Code, &obj.Type, &obj.Desc, &obj.Serial, &obj.Status,
		&obj.Org, &obj.ParentCode, &obj.ParentOrg, &obj.InstallDate,
		&obj.NotUsed, &obj.TenantID, &obj.CreatedAt, &obj.UpdatedAt,
		&obj.CreatedBy, &obj.UpdatedBy,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return &obj, err
}

func (r *pgObjectRepository) FindByCode(ctx context.Context, code string) (*Object, error) {
	query := `
		SELECT obj_id, obj_code, obj_type, obj_desc, obj_serial, obj_status, 
		       obj_org, obj_parent_code, obj_parent_org, obj_install_date,
		       obj_notused, obj_tenant_id, obj_created_at, obj_updated_at,
		       obj_created_by, obj_updated_by
		FROM eamobjects WHERE obj_code = $1`

	var obj Object
	err := r.pool.QueryRow(ctx, query, code).Scan(
		&obj.ID, &obj.Code, &obj.Type, &obj.Desc, &obj.Serial, &obj.Status,
		&obj.Org, &obj.ParentCode, &obj.ParentOrg, &obj.InstallDate,
		&obj.NotUsed, &obj.TenantID, &obj.CreatedAt, &obj.UpdatedAt,
		&obj.CreatedBy, &obj.UpdatedBy,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return &obj, err
}

func (r *pgObjectRepository) FindAll(ctx context.Context, tenantID string, org string, limit, offset int) ([]*Object, int, error) {
	countQuery := `SELECT COUNT(*) FROM eamobjects WHERE obj_tenant_id = $1`
	var total int
	r.pool.QueryRow(ctx, countQuery, tenantID).Scan(&total)

	query := `
		SELECT obj_id, obj_code, obj_type, obj_desc, obj_serial, obj_status, 
		       obj_org, obj_parent_code, obj_parent_org, obj_install_date,
		       obj_notused, obj_tenant_id, obj_created_at, obj_updated_at,
		       obj_created_by, obj_updated_by
		FROM eamobjects WHERE obj_tenant_id = $1`

	args := []interface{}{tenantID}
	argIndex := 2

	if org != "" {
		query += ` AND obj_org = $` + string(rune('0'+argIndex))
		args = append(args, org)
		argIndex++
	}

	query += ` ORDER BY obj_created_at DESC LIMIT $` + string(rune('0'+argIndex)) + ` OFFSET $` + string(rune('0'+argIndex+1))
	args = append(args, limit, offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var objects []*Object
	for rows.Next() {
		var obj Object
		err := rows.Scan(
			&obj.ID, &obj.Code, &obj.Type, &obj.Desc, &obj.Serial, &obj.Status,
			&obj.Org, &obj.ParentCode, &obj.ParentOrg, &obj.InstallDate,
			&obj.NotUsed, &obj.TenantID, &obj.CreatedAt, &obj.UpdatedAt,
			&obj.CreatedBy, &obj.UpdatedBy,
		)
		if err != nil {
			return nil, 0, err
		}
		objects = append(objects, &obj)
	}
	return objects, total, nil
}

func (r *pgObjectRepository) FindByOrg(ctx context.Context, org string) ([]*Object, error) {
	query := `
		SELECT obj_id, obj_code, obj_type, obj_desc, obj_serial, obj_status, 
		       obj_org, obj_parent_code, obj_parent_org, obj_install_date,
		       obj_notused, obj_tenant_id, obj_created_at, obj_updated_at,
		       obj_created_by, obj_updated_by
		FROM eamobjects WHERE obj_org = $1`

	rows, err := r.pool.Query(ctx, query, org)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var objects []*Object
	for rows.Next() {
		var obj Object
		err := rows.Scan(
			&obj.ID, &obj.Code, &obj.Type, &obj.Desc, &obj.Serial, &obj.Status,
			&obj.Org, &obj.ParentCode, &obj.ParentOrg, &obj.InstallDate,
			&obj.NotUsed, &obj.TenantID, &obj.CreatedAt, &obj.UpdatedAt,
			&obj.CreatedBy, &obj.UpdatedBy,
		)
		if err != nil {
			return nil, err
		}
		objects = append(objects, &obj)
	}
	return objects, nil
}

func (r *pgObjectRepository) FindChildren(ctx context.Context, parentCode, parentOrg string) ([]*Object, error) {
	query := `
		SELECT obj_id, obj_code, obj_type, obj_desc, obj_serial, obj_status, 
		       obj_org, obj_parent_code, obj_parent_org, obj_install_date,
		       obj_notused, obj_tenant_id, obj_created_at, obj_updated_at,
		       obj_created_by, obj_updated_by
		FROM eamobjects WHERE obj_parent_code = $1 AND obj_parent_org = $2`

	rows, err := r.pool.Query(ctx, query, parentCode, parentOrg)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var objects []*Object
	for rows.Next() {
		var obj Object
		err := rows.Scan(
			&obj.ID, &obj.Code, &obj.Type, &obj.Desc, &obj.Serial, &obj.Status,
			&obj.Org, &obj.ParentCode, &obj.ParentOrg, &obj.InstallDate,
			&obj.NotUsed, &obj.TenantID, &obj.CreatedAt, &obj.UpdatedAt,
			&obj.CreatedBy, &obj.UpdatedBy,
		)
		if err != nil {
			return nil, err
		}
		objects = append(objects, &obj)
	}
	return objects, nil
}

func (r *pgObjectRepository) Create(ctx context.Context, obj *Object) error {
	query := `
		INSERT INTO eamobjects (obj_id, obj_code, obj_type, obj_desc, obj_serial, obj_status,
		                         obj_org, obj_parent_code, obj_parent_org, obj_install_date,
		                         obj_notused, obj_tenant_id, obj_created_at, obj_updated_at,
		                         obj_created_by, obj_updated_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)`

	_, err := r.pool.Exec(ctx, query,
		obj.ID, obj.Code, obj.Type, obj.Desc, obj.Serial, obj.Status,
		obj.Org, obj.ParentCode, obj.ParentOrg, obj.InstallDate,
		obj.NotUsed, obj.TenantID, obj.CreatedAt, obj.UpdatedAt,
		obj.CreatedBy, obj.UpdatedBy,
	)
	return err
}

func (r *pgObjectRepository) Update(ctx context.Context, obj *Object) error {
	query := `
		UPDATE eamobjects SET obj_type = $1, obj_desc = $2, obj_serial = $3, obj_status = $4,
		                      obj_org = $5, obj_parent_code = $6, obj_parent_org = $7,
		                      obj_install_date = $8, obj_notused = $9, obj_updated_at = $10,
		                      obj_updated_by = $11
		WHERE obj_id = $12`

	_, err := r.pool.Exec(ctx, query,
		obj.Type, obj.Desc, obj.Serial, obj.Status,
		obj.Org, obj.ParentCode, obj.ParentOrg, obj.InstallDate,
		obj.NotUsed, obj.UpdatedAt, obj.UpdatedBy, obj.ID,
	)
	return err
}

func (r *pgObjectRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM eamobjects WHERE obj_id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	return err
}
