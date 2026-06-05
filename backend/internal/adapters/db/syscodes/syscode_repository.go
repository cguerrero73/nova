package db

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	infraDB "github.com/nova/backend/internal/infrastructure/db"

	"github.com/nova/backend/internal/domain/syscodes"
)

type PgSysCodeRepository struct {
	pool *pgxpool.Pool
}

func NewPgSysCodeRepository(pool *pgxpool.Pool) *PgSysCodeRepository {
	return &PgSysCodeRepository{pool: pool}
}

func (r *PgSysCodeRepository) FindByID(ctx context.Context, id string) (*syscodes.SysCode, error) {
	var s *syscodes.SysCode
	query := `
		SELECT sys_id, sys_type, sys_code, sys_ucode, sys_desc, sys_system,
		       sys_notused, sys_created_at, sys_updated_at
		FROM eamsyscodes WHERE sys_id = $1`

	err := infraDB.RunInTenantTx(ctx, r.pool, func(tx pgx.Tx) error {
		var code syscodes.SysCode
		err := tx.QueryRow(ctx, query, id).Scan(
			&code.ID, &code.Type, &code.Code, &code.UCode, &code.Desc, &code.System,
			&code.NotUsed, &code.CreatedAt, &code.UpdatedAt,
		)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		if err != nil {
			return err
		}
		s = &code
		return nil
	})

	return s, err
}

func (r *PgSysCodeRepository) FindByTypeAndCode(ctx context.Context, codeType, code string) (*syscodes.SysCode, error) {
	var s *syscodes.SysCode
	query := `
		SELECT sys_id, sys_type, sys_code, sys_ucode, sys_desc, sys_system,
		       sys_notused, sys_created_at, sys_updated_at
		FROM eamsyscodes WHERE sys_type = $1 AND sys_code = $2`

	err := infraDB.RunInTenantTx(ctx, r.pool, func(tx pgx.Tx) error {
		var code syscodes.SysCode
		err := tx.QueryRow(ctx, query, codeType, code).Scan(
			&code.ID, &code.Type, &code.Code, &code.UCode, &code.Desc, &code.System,
			&code.NotUsed, &code.CreatedAt, &code.UpdatedAt,
		)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		if err != nil {
			return err
		}
		s = &code
		return nil
	})

	return s, err
}

func (r *PgSysCodeRepository) FindByType(ctx context.Context, codeType string) ([]*syscodes.SysCode, error) {
	var result []*syscodes.SysCode
	query := `
		SELECT sys_id, sys_type, sys_code, sys_ucode, sys_desc, sys_system,
		       sys_notused, sys_created_at, sys_updated_at
		FROM eamsyscodes WHERE sys_type = $1
		ORDER BY sys_code ASC`

	err := infraDB.RunInTenantTx(ctx, r.pool, func(tx pgx.Tx) error {
		rows, err := tx.Query(ctx, query, codeType)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var code syscodes.SysCode
			err := rows.Scan(
				&code.ID, &code.Type, &code.Code, &code.UCode, &code.Desc, &code.System,
				&code.NotUsed, &code.CreatedAt, &code.UpdatedAt,
			)
			if err != nil {
				return err
			}
			result = append(result, &code)
		}
		return rows.Err()
	})

	return result, err
}

func (r *PgSysCodeRepository) FindByUCode(ctx context.Context, ucode string) (*syscodes.SysCode, error) {
	var s *syscodes.SysCode
	query := `
		SELECT sys_id, sys_type, sys_code, sys_ucode, sys_desc, sys_system,
		       sys_notused, sys_created_at, sys_updated_at
		FROM eamsyscodes WHERE sys_ucode = $1`

	err := infraDB.RunInTenantTx(ctx, r.pool, func(tx pgx.Tx) error {
		var code syscodes.SysCode
		err := tx.QueryRow(ctx, query, ucode).Scan(
			&code.ID, &code.Type, &code.Code, &code.UCode, &code.Desc, &code.System,
			&code.NotUsed, &code.CreatedAt, &code.UpdatedAt,
		)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		if err != nil {
			return err
		}
		s = &code
		return nil
	})

	return s, err
}

func (r *PgSysCodeRepository) FindAll(ctx context.Context) ([]*syscodes.SysCode, error) {
	var result []*syscodes.SysCode
	query := `
		SELECT sys_id, sys_type, sys_code, sys_ucode, sys_desc, sys_system,
		       sys_notused, sys_created_at, sys_updated_at
		FROM eamsyscodes
		ORDER BY sys_type ASC, sys_code ASC`

	err := infraDB.RunInTenantTx(ctx, r.pool, func(tx pgx.Tx) error {
		rows, err := tx.Query(ctx, query)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var code syscodes.SysCode
			err := rows.Scan(
				&code.ID, &code.Type, &code.Code, &code.UCode, &code.Desc, &code.System,
				&code.NotUsed, &code.CreatedAt, &code.UpdatedAt,
			)
			if err != nil {
				return err
			}
			result = append(result, &code)
		}
		return rows.Err()
	})

	return result, err
}

func (r *PgSysCodeRepository) Create(ctx context.Context, s *syscodes.SysCode) error {
	query := `
		INSERT INTO eamsyscodes (sys_id, sys_type, sys_code, sys_ucode, sys_desc,
		                         sys_system, sys_notused, sys_created_at, sys_updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	return infraDB.RunInTenantTx(ctx, r.pool, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, query,
			s.ID, s.Type, s.Code, s.UCode, s.Desc,
			s.System, s.NotUsed, s.CreatedAt, s.UpdatedAt,
		)
		return err
	})
}

func (r *PgSysCodeRepository) Update(ctx context.Context, s *syscodes.SysCode) error {
	query := `
		UPDATE eamsyscodes 
		SET sys_ucode = $2, sys_desc = $3, sys_notused = $4, sys_updated_at = $5
		WHERE sys_id = $1`

	return infraDB.RunInTenantTx(ctx, r.pool, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, query, s.ID, s.UCode, s.Desc, s.NotUsed, s.UpdatedAt)
		return err
	})
}

func (r *PgSysCodeRepository) Delete(ctx context.Context, id string) error {
	query := `UPDATE eamsyscodes SET sys_notused = '+' WHERE sys_id = $1`
	return infraDB.RunInTenantTx(ctx, r.pool, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, query, id)
		return err
	})
}
