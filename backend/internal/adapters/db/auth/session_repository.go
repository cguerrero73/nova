package db

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	infraDB "github.com/nova/backend/internal/infrastructure/db"

	"github.com/nova/backend/internal/domain/auth"
)

func nilOnEmpty(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

// PgSessionRepository implements auth.SessionRepository using pgx
type PgSessionRepository struct {
	pool *pgxpool.Pool
}

func NewPgSessionRepository(pool *pgxpool.Pool) *PgSessionRepository {
	return &PgSessionRepository{pool: pool}
}

func (r *PgSessionRepository) Create(ctx context.Context, session *auth.Session) error {
	query := `
		INSERT INTO eamsessions (ses_user_code, ses_refresh_token, 
		                          ses_expires_at, ses_ip_address, ses_user_agent, ses_created_at)
		VALUES ($1, $2, $3, $4, $5, $6)`

	return infraDB.RunInTenantTx(ctx, r.pool, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, query,
			session.UserCode, session.RefreshToken,
			session.ExpiresAt, nilOnEmpty(session.IPAddress),
			nilOnEmpty(session.UserAgent), session.CreatedAt,
		)
		return err
	})
}

func (r *PgSessionRepository) FindByRefreshToken(ctx context.Context, token string) (*auth.Session, error) {
	var session *auth.Session
	query := `
		SELECT ses_id, ses_user_code, ses_refresh_token, ses_expires_at, 
		       COALESCE(ses_ip_address, '') as ses_ip_address, 
		       COALESCE(ses_user_agent, '') as ses_user_agent, 
		       ses_created_at, ses_revoked_at
		FROM eamsessions WHERE ses_refresh_token = $1`

	err := infraDB.RunInTenantTx(ctx, r.pool, func(tx pgx.Tx) error {
		var s auth.Session
		err := tx.QueryRow(ctx, query, token).Scan(
			&s.ID, &s.UserCode, &s.RefreshToken,
			&s.ExpiresAt, &s.IPAddress, &s.UserAgent,
			&s.CreatedAt, &s.RevokedAt,
		)
		if err != nil {
			return err
		}
		session = &s
		return nil
	})

	return session, err
}

func (r *PgSessionRepository) Revoke(ctx context.Context, userCode string) error {
	query := `UPDATE eamsessions SET ses_revoked_at = $2 WHERE ses_user_code = $1`
	return infraDB.RunInTenantTx(ctx, r.pool, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, query, userCode, time.Now())
		return err
	})
}

func (r *PgSessionRepository) RevokeAll(ctx context.Context, userCode string) error {
	return r.Revoke(ctx, userCode)
}
