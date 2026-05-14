package users

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type pgUserRepository struct {
	pool *pgxpool.Pool
}

func NewPgUserRepository(pool *pgxpool.Pool) *pgUserRepository {
	return &pgUserRepository{pool: pool}
}

func (r *pgUserRepository) FindByID(ctx context.Context, id string) (*User, error) {
	query := `
		SELECT usr_id, usr_code, usr_name, usr_email, usr_password, usr_phone, 
		       usr_status, usr_default_org, usr_notused, usr_tenant_id,
		       usr_created_at, usr_updated_at, usr_created_by, usr_updated_by
		FROM eamusers WHERE usr_id = $1`

	var user User
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&user.ID, &user.Code, &user.Name, &user.Email, &user.Password, &user.Phone,
		&user.Status, &user.DefaultOrg, &user.NotUsed, &user.TenantID,
		&user.CreatedAt, &user.UpdatedAt, &user.CreatedBy, &user.UpdatedBy,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return &user, err
}

func (r *pgUserRepository) FindByCode(ctx context.Context, code string) (*User, error) {
	query := `
		SELECT usr_id, usr_code, usr_name, usr_email, usr_password, usr_phone, 
		       usr_status, usr_default_org, usr_notused, usr_tenant_id,
		       usr_created_at, usr_updated_at, usr_created_by, usr_updated_by
		FROM eamusers WHERE usr_code = $1`

	var user User
	err := r.pool.QueryRow(ctx, query, code).Scan(
		&user.ID, &user.Code, &user.Name, &user.Email, &user.Password, &user.Phone,
		&user.Status, &user.DefaultOrg, &user.NotUsed, &user.TenantID,
		&user.CreatedAt, &user.UpdatedAt, &user.CreatedBy, &user.UpdatedBy,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return &user, err
}

func (r *pgUserRepository) FindByEmail(ctx context.Context, email string) (*User, error) {
	query := `
		SELECT usr_id, usr_code, usr_name, usr_email, usr_password, usr_phone, 
		       usr_status, usr_default_org, usr_notused, usr_tenant_id,
		       usr_created_at, usr_updated_at, usr_created_by, usr_updated_by
		FROM eamusers WHERE usr_email = $1`

	var user User
	err := r.pool.QueryRow(ctx, query, email).Scan(
		&user.ID, &user.Code, &user.Name, &user.Email, &user.Password, &user.Phone,
		&user.Status, &user.DefaultOrg, &user.NotUsed, &user.TenantID,
		&user.CreatedAt, &user.UpdatedAt, &user.CreatedBy, &user.UpdatedBy,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return &user, err
}

func (r *pgUserRepository) FindAll(ctx context.Context, tenantID string, limit, offset int) ([]*User, int, error) {
	countQuery := `SELECT COUNT(*) FROM eamusers WHERE usr_tenant_id = $1`
	var total int
	if err := r.pool.QueryRow(ctx, countQuery, tenantID).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := `
		SELECT usr_id, usr_code, usr_name, usr_email, usr_password, usr_phone, 
		       usr_status, usr_default_org, usr_notused, usr_tenant_id,
		       usr_created_at, usr_updated_at, usr_created_by, usr_updated_by
		FROM eamusers 
		WHERE usr_tenant_id = $1
		ORDER BY usr_created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.pool.Query(ctx, query, tenantID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var users []*User
	for rows.Next() {
		var user User
		err := rows.Scan(
			&user.ID, &user.Code, &user.Name, &user.Email, &user.Password, &user.Phone,
			&user.Status, &user.DefaultOrg, &user.NotUsed, &user.TenantID,
			&user.CreatedAt, &user.UpdatedAt, &user.CreatedBy, &user.UpdatedBy,
		)
		if err != nil {
			return nil, 0, err
		}
		users = append(users, &user)
	}

	return users, total, nil
}

func (r *pgUserRepository) Create(ctx context.Context, user *User) error {
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

func (r *pgUserRepository) Update(ctx context.Context, user *User) error {
	query := `
		UPDATE eamusers 
		SET usr_name = $2, usr_email = $3, usr_phone = $4, usr_status = $5,
		    usr_default_org = $6, usr_notused = $7, usr_updated_at = $8, 
		    usr_updated_by = $9
		WHERE usr_id = $1`

	_, err := r.pool.Exec(ctx, query,
		user.ID, user.Name, user.Email, user.Phone, user.Status,
		user.DefaultOrg, user.NotUsed, user.UpdatedAt, user.UpdatedBy,
	)
	return err
}

func (r *pgUserRepository) Delete(ctx context.Context, id string) error {
	query := `UPDATE eamusers SET usr_notused = '+', usr_updated_at = $2 WHERE usr_id = $1`
	_, err := r.pool.Exec(ctx, query, id, time.Now())
	return err
}
