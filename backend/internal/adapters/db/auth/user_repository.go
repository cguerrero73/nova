package db

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/nova/backend/internal/domain/auth"
)

// PgUserRepository implements auth.UserRepository using pgx
type PgUserRepository struct {
	pool *pgxpool.Pool
}

func NewPgUserRepository(pool *pgxpool.Pool) *PgUserRepository {
	return &PgUserRepository{pool: pool}
}

func (r *PgUserRepository) FindByEmail(ctx context.Context, email string) (*auth.User, error) {
	query := `
		SELECT usr_id, usr_code, usr_name, usr_email, usr_password, usr_phone, 
		       usr_status, usr_default_org, usr_notused, usr_tenant_id,
		       usr_created_at, usr_updated_at, usr_created_by, usr_updated_by
		FROM eamusers WHERE usr_email = $1`

	var user auth.User
	err := r.pool.QueryRow(ctx, query, email).Scan(
		&user.ID, &user.Code, &user.Name, &user.Email, &user.Password, &user.Phone,
		&user.Status, &user.DefaultOrg, &user.NotUsed, &user.TenantID,
		&user.CreatedAt, &user.UpdatedAt, &user.CreatedBy, &user.UpdatedBy,
	)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *PgUserRepository) FindByCode(ctx context.Context, code string) (*auth.User, error) {
	query := `
		SELECT usr_id, usr_code, usr_name, usr_email, usr_password, usr_phone, 
		       usr_status, usr_default_org, usr_notused, usr_tenant_id,
		       usr_created_at, usr_updated_at, usr_created_by, usr_updated_by
		FROM eamusers WHERE usr_code = $1`

	var user auth.User
	err := r.pool.QueryRow(ctx, query, code).Scan(
		&user.ID, &user.Code, &user.Name, &user.Email, &user.Password, &user.Phone,
		&user.Status, &user.DefaultOrg, &user.NotUsed, &user.TenantID,
		&user.CreatedAt, &user.UpdatedAt, &user.CreatedBy, &user.UpdatedBy,
	)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *PgUserRepository) Create(ctx context.Context, user *auth.User) error {
	query := `
		INSERT INTO eamusers (usr_id, usr_code, usr_name, usr_email, usr_password, 
		                      usr_phone, usr_status, usr_default_org, usr_notused,
		                      usr_tenant_id, usr_created_at, usr_updated_at, 
		                      usr_created_by, usr_updated_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)`

	_, err := r.pool.Exec(ctx, query,
		user.ID, user.Code, user.Name, user.Email, user.Password, user.Phone,
		user.Status, user.DefaultOrg, user.NotUsed, user.TenantID,
		user.CreatedAt, user.UpdatedAt, user.CreatedBy, user.UpdatedBy,
	)
	return err
}
