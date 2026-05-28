package db

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/nova/backend/internal/domain/syscodes"
)

type PgSysCodeRepository struct {
	pool *pgxpool.Pool
}

func NewPgSysCodeRepository(pool *pgxpool.Pool) *PgSysCodeRepository {
	return &PgSysCodeRepository{pool: pool}
}

func (r *PgSysCodeRepository) FindByID(ctx context.Context, id string) (*syscodes.SysCode, error) {
	query := `
		SELECT sys_id, sys_type, sys_code, sys_ucode, sys_desc, sys_system,
		       sys_notused, sys_created_at, sys_updated_at
		FROM eamsyscodes WHERE sys_id = $1`

	var s syscodes.SysCode
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&s.ID, &s.Type, &s.Code, &s.UCode, &s.Desc, &s.System,
		&s.NotUsed, &s.CreatedAt, &s.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return &s, err
}

func (r *PgSysCodeRepository) FindByTypeAndCode(ctx context.Context, codeType, code string) (*syscodes.SysCode, error) {
	query := `
		SELECT sys_id, sys_type, sys_code, sys_ucode, sys_desc, sys_system,
		       sys_notused, sys_created_at, sys_updated_at
		FROM eamsyscodes WHERE sys_type = $1 AND sys_code = $2`

	var s syscodes.SysCode
	err := r.pool.QueryRow(ctx, query, codeType, code).Scan(
		&s.ID, &s.Type, &s.Code, &s.UCode, &s.Desc, &s.System,
		&s.NotUsed, &s.CreatedAt, &s.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return &s, err
}

func (r *PgSysCodeRepository) FindByType(ctx context.Context, codeType string) ([]*syscodes.SysCode, error) {
	query := `
		SELECT sys_id, sys_type, sys_code, sys_ucode, sys_desc, sys_system,
		       sys_notused, sys_created_at, sys_updated_at
		FROM eamsyscodes WHERE sys_type = $1
		ORDER BY sys_code ASC`

	rows, err := r.pool.Query(ctx, query, codeType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*syscodes.SysCode
	for rows.Next() {
		var s syscodes.SysCode
		err := rows.Scan(
			&s.ID, &s.Type, &s.Code, &s.UCode, &s.Desc, &s.System,
			&s.NotUsed, &s.CreatedAt, &s.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		result = append(result, &s)
	}

	return result, nil
}

func (r *PgSysCodeRepository) FindByUCode(ctx context.Context, ucode string) (*syscodes.SysCode, error) {
	query := `
		SELECT sys_id, sys_type, sys_code, sys_ucode, sys_desc, sys_system,
		       sys_notused, sys_created_at, sys_updated_at
		FROM eamsyscodes WHERE sys_ucode = $1`

	var s syscodes.SysCode
	err := r.pool.QueryRow(ctx, query, ucode).Scan(
		&s.ID, &s.Type, &s.Code, &s.UCode, &s.Desc, &s.System,
		&s.NotUsed, &s.CreatedAt, &s.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return &s, err
}

func (r *PgSysCodeRepository) FindAll(ctx context.Context) ([]*syscodes.SysCode, error) {
	query := `
		SELECT sys_id, sys_type, sys_code, sys_ucode, sys_desc, sys_system,
		       sys_notused, sys_created_at, sys_updated_at
		FROM eamsyscodes
		ORDER BY sys_type ASC, sys_code ASC`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*syscodes.SysCode
	for rows.Next() {
		var s syscodes.SysCode
		err := rows.Scan(
			&s.ID, &s.Type, &s.Code, &s.UCode, &s.Desc, &s.System,
			&s.NotUsed, &s.CreatedAt, &s.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		result = append(result, &s)
	}

	return result, nil
}

func (r *PgSysCodeRepository) Create(ctx context.Context, s *syscodes.SysCode) error {
	query := `
		INSERT INTO eamsyscodes (sys_id, sys_type, sys_code, sys_ucode, sys_desc,
		                         sys_system, sys_notused, sys_created_at, sys_updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	_, err := r.pool.Exec(ctx, query,
		s.ID, s.Type, s.Code, s.UCode, s.Desc,
		s.System, s.NotUsed, s.CreatedAt, s.UpdatedAt,
	)
	return err
}

func (r *PgSysCodeRepository) Update(ctx context.Context, s *syscodes.SysCode) error {
	query := `
		UPDATE eamsyscodes 
		SET sys_ucode = $2, sys_desc = $3, sys_notused = $4, sys_updated_at = $5
		WHERE sys_id = $1`

	_, err := r.pool.Exec(ctx, query, s.ID, s.UCode, s.Desc, s.NotUsed, s.UpdatedAt)
	return err
}

func (r *PgSysCodeRepository) Delete(ctx context.Context, id string) error {
	query := `UPDATE eamsyscodes SET sys_notused = '+' WHERE sys_id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	return err
}
