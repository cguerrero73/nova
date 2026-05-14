package stocks

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type pgStockRepository struct {
	pool *pgxpool.Pool
}

func NewPgStockRepository(pool *pgxpool.Pool) *pgStockRepository {
	return &pgStockRepository{pool: pool}
}

func (r *pgStockRepository) FindByID(ctx context.Context, id string) (*Stock, error) {
	query := `
		SELECT stc_id, stc_part_code, stc_part_org, stc_store_code, stc_store_org,
		       stc_min_stock, stc_reorder_qty, stc_actual_qty, stc_notused,
		       stc_tenant_id, stc_created_at, stc_updated_at, stc_created_by, stc_updated_by
		FROM eamstocks WHERE stc_id = $1`

	var s Stock
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&s.ID, &s.PartCode, &s.PartOrg, &s.StoreCode, &s.StoreOrg,
		&s.MinStock, &s.ReorderQty, &s.ActualQty, &s.NotUsed,
		&s.TenantID, &s.CreatedAt, &s.UpdatedAt, &s.CreatedBy, &s.UpdatedBy,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return &s, err
}

func (r *pgStockRepository) FindByPartAndStore(ctx context.Context, partCode, partOrg, storeCode, storeOrg string) (*Stock, error) {
	query := `
		SELECT stc_id, stc_part_code, stc_part_org, stc_store_code, stc_store_org,
		       stc_min_stock, stc_reorder_qty, stc_actual_qty, stc_notused,
		       stc_tenant_id, stc_created_at, stc_updated_at, stc_created_by, stc_updated_by
		FROM eamstocks 
		WHERE stc_part_code = $1 AND stc_part_org = $2 AND stc_store_code = $3 AND stc_store_org = $4`

	var s Stock
	err := r.pool.QueryRow(ctx, query, partCode, partOrg, storeCode, storeOrg).Scan(
		&s.ID, &s.PartCode, &s.PartOrg, &s.StoreCode, &s.StoreOrg,
		&s.MinStock, &s.ReorderQty, &s.ActualQty, &s.NotUsed,
		&s.TenantID, &s.CreatedAt, &s.UpdatedAt, &s.CreatedBy, &s.UpdatedBy,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return &s, err
}

func (r *pgStockRepository) FindByPart(ctx context.Context, partCode, partOrg string) ([]*Stock, error) {
	query := `
		SELECT stc_id, stc_part_code, stc_part_org, stc_store_code, stc_store_org,
		       stc_min_stock, stc_reorder_qty, stc_actual_qty, stc_notused,
		       stc_tenant_id, stc_created_at, stc_updated_at, stc_created_by, stc_updated_by
		FROM eamstocks 
		WHERE stc_part_code = $1 AND stc_part_org = $2`

	rows, err := r.pool.Query(ctx, query, partCode, partOrg)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stocks []*Stock
	for rows.Next() {
		var s Stock
		err := rows.Scan(
			&s.ID, &s.PartCode, &s.PartOrg, &s.StoreCode, &s.StoreOrg,
			&s.MinStock, &s.ReorderQty, &s.ActualQty, &s.NotUsed,
			&s.TenantID, &s.CreatedAt, &s.UpdatedAt, &s.CreatedBy, &s.UpdatedBy,
		)
		if err != nil {
			return nil, err
		}
		stocks = append(stocks, &s)
	}

	return stocks, nil
}

func (r *pgStockRepository) FindAll(ctx context.Context, tenantID string, storeCode, storeOrg string) ([]*Stock, error) {
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

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stocks []*Stock
	for rows.Next() {
		var s Stock
		err := rows.Scan(
			&s.ID, &s.PartCode, &s.PartOrg, &s.StoreCode, &s.StoreOrg,
			&s.MinStock, &s.ReorderQty, &s.ActualQty, &s.NotUsed,
			&s.TenantID, &s.CreatedAt, &s.UpdatedAt, &s.CreatedBy, &s.UpdatedBy,
		)
		if err != nil {
			return nil, err
		}
		stocks = append(stocks, &s)
	}

	return stocks, nil
}

func (r *pgStockRepository) FindLowStock(ctx context.Context, tenantID string) ([]*Stock, error) {
	query := `
		SELECT stc_id, stc_part_code, stc_part_org, stc_store_code, stc_store_org,
		       stc_min_stock, stc_reorder_qty, stc_actual_qty, stc_notused,
		       stc_tenant_id, stc_created_at, stc_updated_at, stc_created_by, stc_updated_by
		FROM eamstocks 
		WHERE stc_tenant_id = $1 AND stc_actual_qty <= stc_min_stock`

	rows, err := r.pool.Query(ctx, query, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stocks []*Stock
	for rows.Next() {
		var s Stock
		err := rows.Scan(
			&s.ID, &s.PartCode, &s.PartOrg, &s.StoreCode, &s.StoreOrg,
			&s.MinStock, &s.ReorderQty, &s.ActualQty, &s.NotUsed,
			&s.TenantID, &s.CreatedAt, &s.UpdatedAt, &s.CreatedBy, &s.UpdatedBy,
		)
		if err != nil {
			return nil, err
		}
		stocks = append(stocks, &s)
	}

	return stocks, nil
}

func (r *pgStockRepository) Create(ctx context.Context, s *Stock) error {
	query := `
		INSERT INTO eamstocks (stc_id, stc_part_code, stc_part_org, stc_store_code, stc_store_org,
		                       stc_min_stock, stc_reorder_qty, stc_actual_qty, stc_notused,
		                       stc_tenant_id, stc_created_at, stc_updated_at, stc_created_by, stc_updated_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)`

	_, err := r.pool.Exec(ctx, query,
		s.ID, s.PartCode, s.PartOrg, s.StoreCode, s.StoreOrg,
		s.MinStock, s.ReorderQty, s.ActualQty, s.NotUsed,
		s.TenantID, s.CreatedAt, s.UpdatedAt, s.CreatedBy, s.UpdatedBy,
	)
	return err
}

func (r *pgStockRepository) Update(ctx context.Context, s *Stock) error {
	query := `
		UPDATE eamstocks 
		SET stc_min_stock = $2, stc_reorder_qty = $3, stc_actual_qty = $4, 
		    stc_notused = $5, stc_updated_at = $6, stc_updated_by = $7
		WHERE stc_id = $1`

	_, err := r.pool.Exec(ctx, query,
		s.ID, s.MinStock, s.ReorderQty, s.ActualQty, s.NotUsed, s.UpdatedAt, s.UpdatedBy,
	)
	return err
}

func (r *pgStockRepository) Delete(ctx context.Context, id string) error {
	query := `UPDATE eamstocks SET stc_notused = '+' WHERE stc_id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	return err
}

func (r *pgStockRepository) UpdateQuantity(ctx context.Context, id string, qty float64) error {
	query := `UPDATE eamstocks SET stc_actual_qty = $2, stc_updated_at = now() WHERE stc_id = $1`
	_, err := r.pool.Exec(ctx, query, id, qty)
	return err
}

type pgBinStockRepository struct {
	pool *pgxpool.Pool
}

func NewPgBinStockRepository(pool *pgxpool.Pool) *pgBinStockRepository {
	return &pgBinStockRepository{pool: pool}
}

func (r *pgBinStockRepository) FindByID(ctx context.Context, id string) (*BinStock, error) {
	query := `
		SELECT bis_id, bis_part_code, bis_part_org, bis_store_code, bis_store_org,
		       bis_bin_code, bis_bin_org, bis_quantity, bis_tenant_id,
		       bis_created_at, bis_updated_at
		FROM eambin_stocks WHERE bis_id = $1`

	var b BinStock
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&b.ID, &b.PartCode, &b.PartOrg, &b.StoreCode, &b.StoreOrg,
		&b.BinCode, &b.BinOrg, &b.Quantity, &b.TenantID,
		&b.CreatedAt, &b.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return &b, err
}

func (r *pgBinStockRepository) FindByPartStoreBin(ctx context.Context, partCode, partOrg, storeCode, storeOrg, binCode, binOrg string) (*BinStock, error) {
	query := `
		SELECT bis_id, bis_part_code, bis_part_org, bis_store_code, bis_store_org,
		       bis_bin_code, bis_bin_org, bis_quantity, bis_tenant_id,
		       bis_created_at, bis_updated_at
		FROM eambin_stocks 
		WHERE bis_part_code = $1 AND bis_part_org = $2 AND bis_store_code = $3 
		  AND bis_store_org = $4 AND bis_bin_code = $5 AND bis_bin_org = $6`

	var b BinStock
	err := r.pool.QueryRow(ctx, query, partCode, partOrg, storeCode, storeOrg, binCode, binOrg).Scan(
		&b.ID, &b.PartCode, &b.PartOrg, &b.StoreCode, &b.StoreOrg,
		&b.BinCode, &b.BinOrg, &b.Quantity, &b.TenantID,
		&b.CreatedAt, &b.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return &b, err
}

func (r *pgBinStockRepository) FindByPartAndStore(ctx context.Context, partCode, partOrg, storeCode, storeOrg string) ([]*BinStock, error) {
	query := `
		SELECT bis_id, bis_part_code, bis_part_org, bis_store_code, bis_store_org,
		       bis_bin_code, bis_bin_org, bis_quantity, bis_tenant_id,
		       bis_created_at, bis_updated_at
		FROM eambin_stocks 
		WHERE bis_part_code = $1 AND bis_part_org = $2 AND bis_store_code = $3 AND bis_store_org = $4`

	rows, err := r.pool.Query(ctx, query, partCode, partOrg, storeCode, storeOrg)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var binStocks []*BinStock
	for rows.Next() {
		var b BinStock
		err := rows.Scan(
			&b.ID, &b.PartCode, &b.PartOrg, &b.StoreCode, &b.StoreOrg,
			&b.BinCode, &b.BinOrg, &b.Quantity, &b.TenantID,
			&b.CreatedAt, &b.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		binStocks = append(binStocks, &b)
	}

	return binStocks, nil
}

func (r *pgBinStockRepository) Create(ctx context.Context, b *BinStock) error {
	query := `
		INSERT INTO eambin_stocks (bis_id, bis_part_code, bis_part_org, bis_store_code, bis_store_org,
		                           bis_bin_code, bis_bin_org, bis_quantity, bis_tenant_id, 
		                           bis_created_at, bis_updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`

	_, err := r.pool.Exec(ctx, query,
		b.ID, b.PartCode, b.PartOrg, b.StoreCode, b.StoreOrg,
		b.BinCode, b.BinOrg, b.Quantity, b.TenantID, b.CreatedAt, b.UpdatedAt,
	)
	return err
}

func (r *pgBinStockRepository) Update(ctx context.Context, b *BinStock) error {
	query := `
		UPDATE eambin_stocks 
		SET bis_quantity = $2, bis_updated_at = $3
		WHERE bis_id = $1`

	_, err := r.pool.Exec(ctx, query, b.ID, b.Quantity, b.UpdatedAt)
	return err
}

func (r *pgBinStockRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM eambin_stocks WHERE bis_id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	return err
}

func (r *pgBinStockRepository) UpdateQuantity(ctx context.Context, id string, qty float64) error {
	query := `UPDATE eambin_stocks SET bis_quantity = $2, bis_updated_at = now() WHERE bis_id = $1`
	_, err := r.pool.Exec(ctx, query, id, qty)
	return err
}
