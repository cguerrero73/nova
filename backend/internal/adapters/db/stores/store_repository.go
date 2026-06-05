package db

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	infraDB "github.com/nova/backend/internal/infrastructure/db"

	"github.com/nova/backend/internal/domain/stores"
)

type pgStoreRepository struct {
	pool *pgxpool.Pool
}

func NewPgStoreRepository(pool *pgxpool.Pool) *pgStoreRepository {
	return &pgStoreRepository{pool: pool}
}

func (r *pgStoreRepository) FindByID(ctx context.Context, id string) (*stores.Store, error) {
	var s *stores.Store
	query := `
		SELECT str_id, str_code, str_name, str_desc, str_org, str_notused,
		       str_tenant_id, str_created_at, str_updated_at
		FROM eamstores WHERE str_id = $1`

	err := infraDB.RunInTenantTx(ctx, r.pool, func(tx pgx.Tx) error {
		var store stores.Store
		err := tx.QueryRow(ctx, query, id).Scan(
			&store.ID, &store.Code, &store.Name, &store.Desc, &store.Org, &store.NotUsed,
			&store.TenantID, &store.CreatedAt, &store.UpdatedAt,
		)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		if err != nil {
			return err
		}
		s = &store
		return nil
	})

	return s, err
}

func (r *pgStoreRepository) FindByCode(ctx context.Context, code string) (*stores.Store, error) {
	var s *stores.Store
	query := `
		SELECT str_id, str_code, str_name, str_desc, str_org, str_notused,
		       str_tenant_id, str_created_at, str_updated_at
		FROM eamstores WHERE str_code = $1`

	err := infraDB.RunInTenantTx(ctx, r.pool, func(tx pgx.Tx) error {
		var store stores.Store
		err := tx.QueryRow(ctx, query, code).Scan(
			&store.ID, &store.Code, &store.Name, &store.Desc, &store.Org, &store.NotUsed,
			&store.TenantID, &store.CreatedAt, &store.UpdatedAt,
		)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		if err != nil {
			return err
		}
		s = &store
		return nil
	})

	return s, err
}

func (r *pgStoreRepository) FindAll(ctx context.Context, tenantID string, org string) ([]*stores.Store, error) {
	var result []*stores.Store

	err := infraDB.RunInTenantTx(ctx, r.pool, func(tx pgx.Tx) error {
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

		rows, err := tx.Query(ctx, query, args...)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var store stores.Store
			err := rows.Scan(
				&store.ID, &store.Code, &store.Name, &store.Desc, &store.Org, &store.NotUsed,
				&store.TenantID, &store.CreatedAt, &store.UpdatedAt,
			)
			if err != nil {
				return err
			}
			result = append(result, &store)
		}
		return rows.Err()
	})

	return result, err
}

func (r *pgStoreRepository) FindByOrg(ctx context.Context, org string) ([]*stores.Store, error) {
	return r.FindAll(ctx, "", org)
}

func (r *pgStoreRepository) Create(ctx context.Context, s *stores.Store) error {
	query := `
		INSERT INTO eamstores (str_id, str_code, str_name, str_desc, str_org,
		                        str_notused, str_tenant_id, str_created_at, str_updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	return infraDB.RunInTenantTx(ctx, r.pool, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, query,
			s.ID, s.Code, s.Name, s.Desc, s.Org, s.NotUsed,
			s.TenantID, s.CreatedAt, s.UpdatedAt,
		)
		return err
	})
}

func (r *pgStoreRepository) Update(ctx context.Context, s *stores.Store) error {
	query := `
		UPDATE eamstores 
		SET str_name = $2, str_desc = $3, str_notused = $4, str_updated_at = $5
		WHERE str_id = $1`

	return infraDB.RunInTenantTx(ctx, r.pool, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, query, s.ID, s.Name, s.Desc, s.NotUsed, s.UpdatedAt)
		return err
	})
}

func (r *pgStoreRepository) Delete(ctx context.Context, id string) error {
	query := `UPDATE eamstores SET str_notused = '+' WHERE str_id = $1`
	return infraDB.RunInTenantTx(ctx, r.pool, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, query, id)
		return err
	})
}

type pgBinRepository struct {
	pool *pgxpool.Pool
}

func NewPgBinRepository(pool *pgxpool.Pool) *pgBinRepository {
	return &pgBinRepository{pool: pool}
}

func (r *pgBinRepository) FindByID(ctx context.Context, id string) (*stores.Bin, error) {
	var b *stores.Bin
	query := `
		SELECT bin_id, bin_code, bin_desc, bin_org, bin_notused,
		       bin_tenant_id, bin_created_at, bin_updated_at
		FROM eambins WHERE bin_id = $1`

	err := infraDB.RunInTenantTx(ctx, r.pool, func(tx pgx.Tx) error {
		var bin stores.Bin
		err := tx.QueryRow(ctx, query, id).Scan(
			&bin.ID, &bin.Code, &bin.Desc, &bin.Org, &bin.NotUsed,
			&bin.TenantID, &bin.CreatedAt, &bin.UpdatedAt,
		)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		if err != nil {
			return err
		}
		b = &bin
		return nil
	})

	return b, err
}

func (r *pgBinRepository) FindByCode(ctx context.Context, code, org string) (*stores.Bin, error) {
	var b *stores.Bin
	query := `
		SELECT bin_id, bin_code, bin_desc, bin_org, bin_notused,
		       bin_tenant_id, bin_created_at, bin_updated_at
		FROM eambins WHERE bin_code = $1 AND bin_org = $2`

	err := infraDB.RunInTenantTx(ctx, r.pool, func(tx pgx.Tx) error {
		var bin stores.Bin
		err := tx.QueryRow(ctx, query, code, org).Scan(
			&bin.ID, &bin.Code, &bin.Desc, &bin.Org, &bin.NotUsed,
			&bin.TenantID, &bin.CreatedAt, &bin.UpdatedAt,
		)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		if err != nil {
			return err
		}
		b = &bin
		return nil
	})

	return b, err
}

func (r *pgBinRepository) FindAll(ctx context.Context, tenantID string, org string) ([]*stores.Bin, error) {
	var result []*stores.Bin

	err := infraDB.RunInTenantTx(ctx, r.pool, func(tx pgx.Tx) error {
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

		rows, err := tx.Query(ctx, query, args...)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var bin stores.Bin
			err := rows.Scan(
				&bin.ID, &bin.Code, &bin.Desc, &bin.Org, &bin.NotUsed,
				&bin.TenantID, &bin.CreatedAt, &bin.UpdatedAt,
			)
			if err != nil {
				return err
			}
			result = append(result, &bin)
		}
		return rows.Err()
	})

	return result, err
}

func (r *pgBinRepository) FindByOrg(ctx context.Context, org string) ([]*stores.Bin, error) {
	return r.FindAll(ctx, "", org)
}

func (r *pgBinRepository) Create(ctx context.Context, b *stores.Bin) error {
	query := `
		INSERT INTO eambins (bin_id, bin_code, bin_desc, bin_org,
		                      bin_notused, bin_tenant_id, bin_created_at, bin_updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	return infraDB.RunInTenantTx(ctx, r.pool, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, query,
			b.ID, b.Code, b.Desc, b.Org, b.NotUsed,
			b.TenantID, b.CreatedAt, b.UpdatedAt,
		)
		return err
	})
}

func (r *pgBinRepository) Update(ctx context.Context, b *stores.Bin) error {
	query := `
		UPDATE eambins 
		SET bin_desc = $2, bin_notused = $3, bin_updated_at = $4
		WHERE bin_id = $1`

	return infraDB.RunInTenantTx(ctx, r.pool, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, query, b.ID, b.Desc, b.NotUsed, b.UpdatedAt)
		return err
	})
}

func (r *pgBinRepository) Delete(ctx context.Context, id string) error {
	query := `UPDATE eambins SET bin_notused = '+' WHERE bin_id = $1`
	return infraDB.RunInTenantTx(ctx, r.pool, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, query, id)
		return err
	})
}
