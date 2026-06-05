package db

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	infraDB "github.com/nova/backend/internal/infrastructure/db"

	"github.com/nova/backend/internal/domain/objects"
)

type PgObjectRepository struct {
	pool *pgxpool.Pool
}

func NewPgObjectRepository(pool *pgxpool.Pool) *PgObjectRepository {
	return &PgObjectRepository{pool: pool}
}

func (r *PgObjectRepository) FindByID(ctx context.Context, id string) (*objects.Object, error) {
	var obj *objects.Object
	query := `
		SELECT obj_id, obj_code, obj_type, obj_desc, obj_serial, obj_status, 
		       obj_org, obj_parent_code, obj_parent_org, obj_install_date,
		       obj_notused, obj_tenant_id, obj_created_at, obj_updated_at,
		       obj_created_by, obj_updated_by
		FROM eamobjects WHERE obj_id = $1`

	err := infraDB.RunInTenantTx(ctx, r.pool, func(tx pgx.Tx) error {
		var o objects.Object
		err := tx.QueryRow(ctx, query, id).Scan(
			&o.ID, &o.Code, &o.Type, &o.Desc, &o.Serial, &o.Status,
			&o.Org, &o.ParentCode, &o.ParentOrg, &o.InstallDate,
			&o.NotUsed, &o.TenantID, &o.CreatedAt, &o.UpdatedAt,
			&o.CreatedBy, &o.UpdatedBy,
		)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		if err != nil {
			return err
		}
		obj = &o
		return nil
	})

	return obj, err
}

func (r *PgObjectRepository) FindByCode(ctx context.Context, code string) (*objects.Object, error) {
	var obj *objects.Object
	query := `
		SELECT obj_id, obj_code, obj_type, obj_desc, obj_serial, obj_status, 
		       obj_org, obj_parent_code, obj_parent_org, obj_install_date,
		       obj_notused, obj_tenant_id, obj_created_at, obj_updated_at,
		       obj_created_by, obj_updated_by
		FROM eamobjects WHERE obj_code = $1`

	err := infraDB.RunInTenantTx(ctx, r.pool, func(tx pgx.Tx) error {
		var o objects.Object
		err := tx.QueryRow(ctx, query, code).Scan(
			&o.ID, &o.Code, &o.Type, &o.Desc, &o.Serial, &o.Status,
			&o.Org, &o.ParentCode, &o.ParentOrg, &o.InstallDate,
			&o.NotUsed, &o.TenantID, &o.CreatedAt, &o.UpdatedAt,
			&o.CreatedBy, &o.UpdatedBy,
		)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		if err != nil {
			return err
		}
		obj = &o
		return nil
	})

	return obj, err
}

func (r *PgObjectRepository) FindAll(ctx context.Context, tenantID string, org string, limit, offset int) ([]*objects.Object, int, error) {
	var total int
	var result []*objects.Object

	err := infraDB.RunInTenantTx(ctx, r.pool, func(tx pgx.Tx) error {
		countQuery := `SELECT COUNT(*) FROM eamobjects WHERE obj_tenant_id = $1`
		if err := tx.QueryRow(ctx, countQuery, tenantID).Scan(&total); err != nil {
			return err
		}

		query := `
			SELECT obj_id, obj_code, obj_type, obj_desc, obj_serial, obj_status, 
			       obj_org, obj_parent_code, obj_parent_org, obj_install_date,
			       obj_notused, obj_tenant_id, obj_created_at, obj_updated_at,
			       obj_created_by, obj_updated_by
			FROM eamobjects 
			WHERE obj_tenant_id = $1
			ORDER BY obj_created_at DESC
			LIMIT $2 OFFSET $3`

		rows, err := tx.Query(ctx, query, tenantID, limit, offset)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var o objects.Object
			err := rows.Scan(
				&o.ID, &o.Code, &o.Type, &o.Desc, &o.Serial, &o.Status,
				&o.Org, &o.ParentCode, &o.ParentOrg, &o.InstallDate,
				&o.NotUsed, &o.TenantID, &o.CreatedAt, &o.UpdatedAt,
				&o.CreatedBy, &o.UpdatedBy,
			)
			if err != nil {
				return err
			}
			result = append(result, &o)
		}
		return rows.Err()
	})

	return result, total, err
}

func (r *PgObjectRepository) FindByOrg(ctx context.Context, org string) ([]*objects.Object, error) {
	var result []*objects.Object
	query := `
		SELECT obj_id, obj_code, obj_type, obj_desc, obj_serial, obj_status, 
		       obj_org, obj_parent_code, obj_parent_org, obj_install_date,
		       obj_notused, obj_tenant_id, obj_created_at, obj_updated_at,
		       obj_created_by, obj_updated_by
		FROM eamobjects WHERE obj_org = $1`

	err := infraDB.RunInTenantTx(ctx, r.pool, func(tx pgx.Tx) error {
		rows, err := tx.Query(ctx, query, org)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var o objects.Object
			err := rows.Scan(
				&o.ID, &o.Code, &o.Type, &o.Desc, &o.Serial, &o.Status,
				&o.Org, &o.ParentCode, &o.ParentOrg, &o.InstallDate,
				&o.NotUsed, &o.TenantID, &o.CreatedAt, &o.UpdatedAt,
				&o.CreatedBy, &o.UpdatedBy,
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

func (r *PgObjectRepository) FindChildren(ctx context.Context, parentCode, parentOrg string) ([]*objects.Object, error) {
	var result []*objects.Object
	query := `
		SELECT obj_id, obj_code, obj_type, obj_desc, obj_serial, obj_status, 
		       obj_org, obj_parent_code, obj_parent_org, obj_install_date,
		       obj_notused, obj_tenant_id, obj_created_at, obj_updated_at,
		       obj_created_by, obj_updated_by
		FROM eamobjects 
		WHERE obj_parent_code = $1 AND obj_parent_org = $2`

	err := infraDB.RunInTenantTx(ctx, r.pool, func(tx pgx.Tx) error {
		rows, err := tx.Query(ctx, query, parentCode, parentOrg)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var o objects.Object
			err := rows.Scan(
				&o.ID, &o.Code, &o.Type, &o.Desc, &o.Serial, &o.Status,
				&o.Org, &o.ParentCode, &o.ParentOrg, &o.InstallDate,
				&o.NotUsed, &o.TenantID, &o.CreatedAt, &o.UpdatedAt,
				&o.CreatedBy, &o.UpdatedBy,
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

func (r *PgObjectRepository) Create(ctx context.Context, obj *objects.Object) error {
	query := `
		INSERT INTO eamobjects (obj_id, obj_code, obj_type, obj_desc, obj_serial, 
		                        obj_status, obj_org, obj_parent_code, obj_parent_org,
		                        obj_install_date, obj_notused, obj_tenant_id, 
		                        obj_created_at, obj_updated_at, obj_created_by, obj_updated_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)`

	return infraDB.RunInTenantTx(ctx, r.pool, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, query,
			obj.ID, obj.Code, obj.Type, obj.Desc, obj.Serial, obj.Status,
			obj.Org, obj.ParentCode, obj.ParentOrg, obj.InstallDate,
			obj.NotUsed, obj.TenantID, obj.CreatedAt, obj.UpdatedAt,
			obj.CreatedBy, obj.UpdatedBy,
		)
		return err
	})
}

func (r *PgObjectRepository) Update(ctx context.Context, obj *objects.Object) error {
	query := `
		UPDATE eamobjects 
		SET obj_type = $2, obj_desc = $3, obj_serial = $4, obj_status = $5,
		    obj_org = $6, obj_parent_code = $7, obj_parent_org = $8,
		    obj_install_date = $9, obj_notused = $10, obj_updated_at = $11,
		    obj_updated_by = $12
		WHERE obj_id = $1`

	return infraDB.RunInTenantTx(ctx, r.pool, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, query,
			obj.ID, obj.Type, obj.Desc, obj.Serial, obj.Status,
			obj.Org, obj.ParentCode, obj.ParentOrg, obj.InstallDate,
			obj.NotUsed, time.Now(), obj.UpdatedBy,
		)
		return err
	})
}

func (r *PgObjectRepository) Delete(ctx context.Context, id string) error {
	query := `UPDATE eamobjects SET obj_notused = '+', obj_updated_at = $2 WHERE obj_id = $1`
	return infraDB.RunInTenantTx(ctx, r.pool, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, query, id, time.Now())
		return err
	})
}
