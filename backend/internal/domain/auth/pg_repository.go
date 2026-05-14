package auth

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// PgUserRepository implements auth.UserRepository using pgx
type PgUserRepository struct {
	pool *pgxpool.Pool
}

func NewPgUserRepository(pool *pgxpool.Pool) *PgUserRepository {
	return &PgUserRepository{pool: pool}
}

func (r *PgUserRepository) FindByEmail(ctx context.Context, email string) (*User, error) {
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
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *PgUserRepository) FindByCode(ctx context.Context, code string) (*User, error) {
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
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *PgUserRepository) Create(ctx context.Context, user *User) error {
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

// PgSessionRepository implements auth.SessionRepository using pgx
type PgSessionRepository struct {
	pool *pgxpool.Pool
}

func NewPgSessionRepository(pool *pgxpool.Pool) *PgSessionRepository {
	return &PgSessionRepository{pool: pool}
}

func (r *PgSessionRepository) Create(ctx context.Context, session *Session) error {
	query := `
		INSERT INTO eamsessions (ses_id, ses_user_code, ses_refresh_token, 
		                          ses_expires_at, ses_ip_address, ses_user_agent, ses_created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err := r.pool.Exec(ctx, query,
		session.ID, session.UserCode, session.RefreshToken,
		session.ExpiresAt, session.IPAddress, session.UserAgent, session.CreatedAt,
	)
	return err
}

func (r *PgSessionRepository) FindByRefreshToken(ctx context.Context, token string) (*Session, error) {
	query := `
		SELECT ses_id, ses_user_code, ses_refresh_token, ses_expires_at, 
		       ses_ip_address, ses_user_agent, ses_created_at, ses_revoked_at
		FROM eamsessions WHERE ses_refresh_token = $1`

	var session Session
	err := r.pool.QueryRow(ctx, query, token).Scan(
		&session.ID, &session.UserCode, &session.RefreshToken,
		&session.ExpiresAt, &session.IPAddress, &session.UserAgent,
		&session.CreatedAt, &session.RevokedAt,
	)
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *PgSessionRepository) Revoke(ctx context.Context, userCode string) error {
	query := `UPDATE eamsessions SET ses_revoked_at = $2 WHERE ses_user_code = $1`
	_, err := r.pool.Exec(ctx, query, userCode, time.Now())
	return err
}

func (r *PgSessionRepository) RevokeAll(ctx context.Context, userCode string) error {
	return r.Revoke(ctx, userCode)
}
