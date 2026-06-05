package db

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	infraDB "github.com/nova/backend/internal/infrastructure/db"

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
	var user *auth.User
	query := `
		SELECT usr_id, usr_code, usr_name, usr_email, usr_password, usr_phone, 
		       usr_status, usr_default_org, usr_notused, usr_tenant_id,
		       usr_created_at, usr_updated_at, usr_created_by, usr_updated_by
		FROM eamusers WHERE usr_email = $1`

	err := infraDB.RunInTenantTx(ctx, r.pool, func(tx pgx.Tx) error {
		var u auth.User
		err := tx.QueryRow(ctx, query, email).Scan(
			&u.ID, &u.Code, &u.Name, &u.Email, &u.Password, &u.Phone,
			&u.Status, &u.DefaultOrg, &u.NotUsed, &u.TenantID,
			&u.CreatedAt, &u.UpdatedAt, &u.CreatedBy, &u.UpdatedBy,
		)
		if err != nil {
			return err
		}
		user = &u
		return nil
	})

	return user, err
}

func (r *PgUserRepository) FindByCode(ctx context.Context, code string) (*auth.User, error) {
	var user *auth.User
	query := `
		SELECT usr_id, usr_code, usr_name, usr_email, usr_password, usr_phone, 
		       usr_status, usr_default_org, usr_notused, usr_tenant_id,
		       usr_created_at, usr_updated_at, usr_created_by, usr_updated_by
		FROM eamusers WHERE usr_code = $1`

	err := infraDB.RunInTenantTx(ctx, r.pool, func(tx pgx.Tx) error {
		var u auth.User
		err := tx.QueryRow(ctx, query, code).Scan(
			&u.ID, &u.Code, &u.Name, &u.Email, &u.Password, &u.Phone,
			&u.Status, &u.DefaultOrg, &u.NotUsed, &u.TenantID,
			&u.CreatedAt, &u.UpdatedAt, &u.CreatedBy, &u.UpdatedBy,
		)
		if err != nil {
			return err
		}
		user = &u
		return nil
	})

	return user, err
}

func (r *PgUserRepository) Create(ctx context.Context, user *auth.User) error {
	query := `
		INSERT INTO eamusers (usr_id, usr_code, usr_name, usr_email, usr_password, 
		                      usr_phone, usr_status, usr_default_org, usr_notused,
		                      usr_tenant_id, usr_created_at, usr_updated_at, 
		                      usr_created_by, usr_updated_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)`

	return infraDB.RunInTenantTx(ctx, r.pool, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, query,
			user.ID, user.Code, user.Name, user.Email, user.Password, user.Phone,
			user.Status, user.DefaultOrg, user.NotUsed, user.TenantID,
			user.CreatedAt, user.UpdatedAt, user.CreatedBy, user.UpdatedBy,
		)
		return err
	})
}
