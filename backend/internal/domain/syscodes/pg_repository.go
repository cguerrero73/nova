package syscodes

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type pgSysCodeRepository struct {
	pool *pgxpool.Pool
}

func NewPgSysCodeRepository(pool *pgxpool.Pool) *pgSysCodeRepository {
	return &pgSysCodeRepository{pool: pool}
}

func (r *pgSysCodeRepository) FindByID(ctx context.Context, id string) (*SysCode, error) {
	query := `
		SELECT sys_id, sys_type, sys_code, sys_ucode, sys_desc, sys_system,
		       sys_notused, sys_created_at, sys_updated_at
		FROM eamsyscodes WHERE sys_id = $1`

	var s SysCode
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&s.ID, &s.Type, &s.Code, &s.UCode, &s.Desc, &s.System,
		&s.NotUsed, &s.CreatedAt, &s.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return &s, err
}

func (r *pgSysCodeRepository) FindByTypeAndCode(ctx context.Context, codeType, code string) (*SysCode, error) {
	query := `
		SELECT sys_id, sys_type, sys_code, sys_ucode, sys_desc, sys_system,
		       sys_notused, sys_created_at, sys_updated_at
		FROM eamsyscodes WHERE sys_type = $1 AND sys_code = $2`

	var s SysCode
	err := r.pool.QueryRow(ctx, query, codeType, code).Scan(
		&s.ID, &s.Type, &s.Code, &s.UCode, &s.Desc, &s.System,
		&s.NotUsed, &s.CreatedAt, &s.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return &s, err
}

func (r *pgSysCodeRepository) FindByType(ctx context.Context, codeType string) ([]*SysCode, error) {
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

	var syscodes []*SysCode
	for rows.Next() {
		var s SysCode
		err := rows.Scan(
			&s.ID, &s.Type, &s.Code, &s.UCode, &s.Desc, &s.System,
			&s.NotUsed, &s.CreatedAt, &s.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		syscodes = append(syscodes, &s)
	}

	return syscodes, nil
}

func (r *pgSysCodeRepository) FindByUCode(ctx context.Context, ucode string) (*SysCode, error) {
	query := `
		SELECT sys_id, sys_type, sys_code, sys_ucode, sys_desc, sys_system,
		       sys_notused, sys_created_at, sys_updated_at
		FROM eamsyscodes WHERE sys_ucode = $1`

	var s SysCode
	err := r.pool.QueryRow(ctx, query, ucode).Scan(
		&s.ID, &s.Type, &s.Code, &s.UCode, &s.Desc, &s.System,
		&s.NotUsed, &s.CreatedAt, &s.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return &s, err
}

func (r *pgSysCodeRepository) FindAll(ctx context.Context) ([]*SysCode, error) {
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

	var syscodes []*SysCode
	for rows.Next() {
		var s SysCode
		err := rows.Scan(
			&s.ID, &s.Type, &s.Code, &s.UCode, &s.Desc, &s.System,
			&s.NotUsed, &s.CreatedAt, &s.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		syscodes = append(syscodes, &s)
	}

	return syscodes, nil
}

func (r *pgSysCodeRepository) Create(ctx context.Context, s *SysCode) error {
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

func (r *pgSysCodeRepository) Update(ctx context.Context, s *SysCode) error {
	query := `
		UPDATE eamsyscodes 
		SET sys_ucode = $2, sys_desc = $3, sys_notused = $4, sys_updated_at = $5
		WHERE sys_id = $1`

	_, err := r.pool.Exec(ctx, query, s.ID, s.UCode, s.Desc, s.NotUsed, s.UpdatedAt)
	return err
}

func (r *pgSysCodeRepository) Delete(ctx context.Context, id string) error {
	query := `UPDATE eamsyscodes SET sys_notused = '+' WHERE sys_id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	return err
}
