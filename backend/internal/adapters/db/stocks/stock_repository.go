package db

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	infraDB "github.com/nova/backend/internal/infrastructure/db"

	"github.com/nova/backend/internal/domain/stocks"
)

type PgStockRepository struct {
	pool *pgxpool.Pool
}

func NewPgStockRepository(pool *pgxpool.Pool) *PgStockRepository {
	return &PgStockRepository{pool: pool}
}

func (r *PgStockRepository) FindByID(ctx context.Context, id string) (*stocks.Stock, error) {
	var s *stocks.Stock
	query := `
		SELECT stc_id, stc_part_code, stc_part_org, stc_store_code, stc_store_org,
		       stc_min_stock, stc_reorder_qty, stc_actual_qty, stc_notused,
		       stc_tenant_id, stc_created_at, stc_updated_at, stc_created_by, stc_updated_by
		FROM eamstocks WHERE stc_id = $1`

	err := infraDB.RunInTenantTx(ctx, r.pool, func(tx pgx.Tx) error {
		var stock stocks.Stock
		err := tx.QueryRow(ctx, query, id).Scan(
			&stock.ID, &stock.PartCode, &stock.PartOrg, &stock.StoreCode, &stock.StoreOrg,
			&stock.MinStock, &stock.ReorderQty, &stock.ActualQty, &stock.NotUsed,
			&stock.TenantID, &stock.CreatedAt, &stock.UpdatedAt, &stock.CreatedBy, &stock.UpdatedBy,
		)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		if err != nil {
			return err
		}
		s = &stock
		return nil
	})

	return s, err
}

func (r *PgStockRepository) FindByPartAndStore(ctx context.Context, partCode, partOrg, storeCode, storeOrg string) (*stocks.Stock, error) {
	var s *stocks.Stock
	query := `
		SELECT stc_id, stc_part_code, stc_part_org, stc_store_code, stc_store_org,
		       stc_min_stock, stc_reorder_qty, stc_actual_qty, stc_notused,
		       stc_tenant_id, stc_created_at, stc_updated_at, stc_created_by, stc_updated_by
		FROM eamstocks 
		WHERE stc_part_code = $1 AND stc_part_org = $2 AND stc_store_code = $3 AND stc_store_org = $4`

	err := infraDB.RunInTenantTx(ctx, r.pool, func(tx pgx.Tx) error {
		var stock stocks.Stock
		err := tx.QueryRow(ctx, query, partCode, partOrg, storeCode, storeOrg).Scan(
			&stock.ID, &stock.PartCode, &stock.PartOrg, &stock.StoreCode, &stock.StoreOrg,
			&stock.MinStock, &stock.ReorderQty, &stock.ActualQty, &stock.NotUsed,
			&stock.TenantID, &stock.CreatedAt, &stock.UpdatedAt, &stock.CreatedBy, &stock.UpdatedBy,
		)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		if err != nil {
			return err
		}
		s = &stock
		return nil
	})

	return s, err
}

func (r *PgStockRepository) FindByPart(ctx context.Context, partCode, partOrg string) ([]*stocks.Stock, error) {
	var result []*stocks.Stock
	query := `
		SELECT stc_id, stc_part_code, stc_part_org, stc_store_code, stc_store_org,
		       stc_min_stock, stc_reorder_qty, stc_actual_qty, stc_notused,
		       stc_tenant_id, stc_created_at, stc_updated_at, stc_created_by, stc_updated_by
		FROM eamstocks 
		WHERE stc_part_code = $1 AND stc_part_org = $2`

	err := infraDB.RunInTenantTx(ctx, r.pool, func(tx pgx.Tx) error {
		rows, err := tx.Query(ctx, query, partCode, partOrg)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var stock stocks.Stock
			err := rows.Scan(
				&stock.ID, &stock.PartCode, &stock.PartOrg, &stock.StoreCode, &stock.StoreOrg,
				&stock.MinStock, &stock.ReorderQty, &stock.ActualQty, &stock.NotUsed,
				&stock.TenantID, &stock.CreatedAt, &stock.UpdatedAt, &stock.CreatedBy, &stock.UpdatedBy,
			)
			if err != nil {
				return err
			}
			result = append(result, &stock)
		}
		return rows.Err()
	})

	return result, err
}

func (r *PgStockRepository) FindAll(ctx context.Context, tenantID string, storeCode, storeOrg string) ([]*stocks.Stock, error) {
	var result []*stocks.Stock

	err := infraDB.RunInTenantTx(ctx, r.pool, func(tx pgx.Tx) error {
		query := `
			SELECT stc_id, stc_part_code, stc_part_org, stc_store_code, stc_store_org,
			       stc_min_stock, stc_reorder_qty, stc_actual_qty, stc_notused,
			       stc_tenant_id, stc_created_at, stc_updated_at, stc_created_by, stc_updated_by
			FROM eamstocks WHERE stc_tenant_id = $1`
		args := []interface{}{tenantID}

		if storeCode != "" {
			query += ` AND stc_store_code = $2 AND stc_store_org = $3`
			args = append(args, storeCode, storeOrg)
		}

		rows, err := tx.Query(ctx, query, args...)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var stock stocks.Stock
			err := rows.Scan(
				&stock.ID, &stock.PartCode, &stock.PartOrg, &stock.StoreCode, &stock.StoreOrg,
				&stock.MinStock, &stock.ReorderQty, &stock.ActualQty, &stock.NotUsed,
				&stock.TenantID, &stock.CreatedAt, &stock.UpdatedAt, &stock.CreatedBy, &stock.UpdatedBy,
			)
			if err != nil {
				return err
			}
			result = append(result, &stock)
		}
		return rows.Err()
	})

	return result, err
}

func (r *PgStockRepository) FindLowStock(ctx context.Context, tenantID string) ([]*stocks.Stock, error) {
	var result []*stocks.Stock
	query := `
		SELECT stc_id, stc_part_code, stc_part_org, stc_store_code, stc_store_org,
		       stc_min_stock, stc_reorder_qty, stc_actual_qty, stc_notused,
		       stc_tenant_id, stc_created_at, stc_updated_at, stc_created_by, stc_updated_by
		FROM eamstocks 
		WHERE stc_tenant_id = $1 AND stc_actual_qty <= stc_min_stock`

	err := infraDB.RunInTenantTx(ctx, r.pool, func(tx pgx.Tx) error {
		rows, err := tx.Query(ctx, query, tenantID)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var stock stocks.Stock
			err := rows.Scan(
				&stock.ID, &stock.PartCode, &stock.PartOrg, &stock.StoreCode, &stock.StoreOrg,
				&stock.MinStock, &stock.ReorderQty, &stock.ActualQty, &stock.NotUsed,
				&stock.TenantID, &stock.CreatedAt, &stock.UpdatedAt, &stock.CreatedBy, &stock.UpdatedBy,
			)
			if err != nil {
				return err
			}
			result = append(result, &stock)
		}
		return rows.Err()
	})

	return result, err
}

func (r *PgStockRepository) Create(ctx context.Context, s *stocks.Stock) error {
	query := `
		INSERT INTO eamstocks (stc_id, stc_part_code, stc_part_org, stc_store_code, stc_store_org,
		                       stc_min_stock, stc_reorder_qty, stc_actual_qty, stc_notused,
		                       stc_tenant_id, stc_created_at, stc_updated_at, stc_created_by, stc_updated_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)`

	return infraDB.RunInTenantTx(ctx, r.pool, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, query,
			s.ID, s.PartCode, s.PartOrg, s.StoreCode, s.StoreOrg,
			s.MinStock, s.ReorderQty, s.ActualQty, s.NotUsed,
			s.TenantID, s.CreatedAt, s.UpdatedAt, s.CreatedBy, s.UpdatedBy,
		)
		return err
	})
}

func (r *PgStockRepository) Update(ctx context.Context, s *stocks.Stock) error {
	query := `
		UPDATE eamstocks 
		SET stc_min_stock = $2, stc_reorder_qty = $3, stc_actual_qty = $4, 
		    stc_notused = $5, stc_updated_at = $6, stc_updated_by = $7
		WHERE stc_id = $1`

	return infraDB.RunInTenantTx(ctx, r.pool, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, query,
			s.ID, s.MinStock, s.ReorderQty, s.ActualQty, s.NotUsed, s.UpdatedAt, s.UpdatedBy,
		)
		return err
	})
}

func (r *PgStockRepository) Delete(ctx context.Context, id string) error {
	query := `UPDATE eamstocks SET stc_notused = '+' WHERE stc_id = $1`
	return infraDB.RunInTenantTx(ctx, r.pool, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, query, id)
		return err
	})
}

func (r *PgStockRepository) UpdateQuantity(ctx context.Context, id string, qty float64) error {
	query := `UPDATE eamstocks SET stc_actual_qty = $2, stc_updated_at = $3 WHERE stc_id = $1`
	return infraDB.RunInTenantTx(ctx, r.pool, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, query, id, qty, time.Now())
		return err
	})
}

// PgBinStockRepository handles bin stock database operations
type PgBinStockRepository struct {
	pool *pgxpool.Pool
}

func NewPgBinStockRepository(pool *pgxpool.Pool) *PgBinStockRepository {
	return &PgBinStockRepository{pool: pool}
}

func (r *PgBinStockRepository) FindByID(ctx context.Context, id string) (*stocks.BinStock, error) {
	var b *stocks.BinStock
	query := `
		SELECT bis_id, bis_part_code, bis_part_org, bis_store_code, bis_store_org,
		       bis_bin_code, bis_bin_org, bis_quantity, bis_tenant_id,
		       bis_created_at, bis_updated_at
		FROM eambin_stocks WHERE bis_id = $1`

	err := infraDB.RunInTenantTx(ctx, r.pool, func(tx pgx.Tx) error {
		var binStock stocks.BinStock
		err := tx.QueryRow(ctx, query, id).Scan(
			&binStock.ID, &binStock.PartCode, &binStock.PartOrg, &binStock.StoreCode, &binStock.StoreOrg,
			&binStock.BinCode, &binStock.BinOrg, &binStock.Quantity, &binStock.TenantID,
			&binStock.CreatedAt, &binStock.UpdatedAt,
		)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		if err != nil {
			return err
		}
		b = &binStock
		return nil
	})

	return b, err
}

func (r *PgBinStockRepository) FindByPartStoreBin(ctx context.Context, partCode, partOrg, storeCode, storeOrg, binCode, binOrg string) (*stocks.BinStock, error) {
	var b *stocks.BinStock
	query := `
		SELECT bis_id, bis_part_code, bis_part_org, bis_store_code, bis_store_org,
		       bis_bin_code, bis_bin_org, bis_quantity, bis_tenant_id,
		       bis_created_at, bis_updated_at
		FROM eambin_stocks 
		WHERE bis_part_code = $1 AND bis_part_org = $2 AND bis_store_code = $3 
		  AND bis_store_org = $4 AND bis_bin_code = $5 AND bis_bin_org = $6`

	err := infraDB.RunInTenantTx(ctx, r.pool, func(tx pgx.Tx) error {
		var binStock stocks.BinStock
		err := tx.QueryRow(ctx, query, partCode, partOrg, storeCode, storeOrg, binCode, binOrg).Scan(
			&binStock.ID, &binStock.PartCode, &binStock.PartOrg, &binStock.StoreCode, &binStock.StoreOrg,
			&binStock.BinCode, &binStock.BinOrg, &binStock.Quantity, &binStock.TenantID,
			&binStock.CreatedAt, &binStock.UpdatedAt,
		)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		if err != nil {
			return err
		}
		b = &binStock
		return nil
	})

	return b, err
}

func (r *PgBinStockRepository) FindByPartAndStore(ctx context.Context, partCode, partOrg, storeCode, storeOrg string) ([]*stocks.BinStock, error) {
	var result []*stocks.BinStock
	query := `
		SELECT bis_id, bis_part_code, bis_part_org, bis_store_code, bis_store_org,
		       bis_bin_code, bis_bin_org, bis_quantity, bis_tenant_id,
		       bis_created_at, bis_updated_at
		FROM eambin_stocks 
		WHERE bis_part_code = $1 AND bis_part_org = $2 AND bis_store_code = $3 AND bis_store_org = $4`

	err := infraDB.RunInTenantTx(ctx, r.pool, func(tx pgx.Tx) error {
		rows, err := tx.Query(ctx, query, partCode, partOrg, storeCode, storeOrg)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var binStock stocks.BinStock
			err := rows.Scan(
				&binStock.ID, &binStock.PartCode, &binStock.PartOrg, &binStock.StoreCode, &binStock.StoreOrg,
				&binStock.BinCode, &binStock.BinOrg, &binStock.Quantity, &binStock.TenantID,
				&binStock.CreatedAt, &binStock.UpdatedAt,
			)
			if err != nil {
				return err
			}
			result = append(result, &binStock)
		}
		return rows.Err()
	})

	return result, err
}

func (r *PgBinStockRepository) Create(ctx context.Context, b *stocks.BinStock) error {
	query := `
		INSERT INTO eambin_stocks (bis_id, bis_part_code, bis_part_org, bis_store_code, bis_store_org,
		                           bis_bin_code, bis_bin_org, bis_quantity, bis_tenant_id, 
		                           bis_created_at, bis_updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`

	return infraDB.RunInTenantTx(ctx, r.pool, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, query,
			b.ID, b.PartCode, b.PartOrg, b.StoreCode, b.StoreOrg,
			b.BinCode, b.BinOrg, b.Quantity, b.TenantID, b.CreatedAt, b.UpdatedAt,
		)
		return err
	})
}

func (r *PgBinStockRepository) Update(ctx context.Context, b *stocks.BinStock) error {
	query := `
		UPDATE eambin_stocks 
		SET bis_quantity = $2, bis_updated_at = $3
		WHERE bis_id = $1`

	return infraDB.RunInTenantTx(ctx, r.pool, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, query, b.ID, b.Quantity, b.UpdatedAt)
		return err
	})
}

func (r *PgBinStockRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM eambin_stocks WHERE bis_id = $1`
	return infraDB.RunInTenantTx(ctx, r.pool, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, query, id)
		return err
	})
}

func (r *PgBinStockRepository) UpdateQuantity(ctx context.Context, id string, qty float64) error {
	query := `UPDATE eambin_stocks SET bis_quantity = $2, bis_updated_at = $3 WHERE bis_id = $1`
	return infraDB.RunInTenantTx(ctx, r.pool, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, query, id, qty, time.Now())
		return err
	})
}
