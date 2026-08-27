package db

import (
	"context"
	"errors"
	"strconv"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	infraDB "github.com/nova/backend/internal/infrastructure/db"

	"github.com/nova/backend/internal/domain/users"
)

type PgUserRepository struct {
	pool *pgxpool.Pool
}

func NewPgUserRepository(pool *pgxpool.Pool) *PgUserRepository {
	return &PgUserRepository{pool: pool}
}

func (r *PgUserRepository) FindByID(ctx context.Context, id string) (*users.User, error) {
	var user *users.User
	query := `
		SELECT usr_id, usr_code, usr_name, usr_email, usr_password, usr_phone, 
		       usr_status, usr_default_org, usr_notused, usr_tenant_id,
		       usr_created_at, usr_updated_at, usr_created_by, usr_updated_by
		FROM eamusers WHERE usr_id = $1`

	err := infraDB.RunInTenantTx(ctx, r.pool, func(tx pgx.Tx) error {
		var u users.User
		err := tx.QueryRow(ctx, query, id).Scan(
			&u.ID, &u.Code, &u.Name, &u.Email, &u.Password, &u.Phone,
			&u.Status, &u.DefaultOrg, &u.NotUsed, &u.TenantID,
			&u.CreatedAt, &u.UpdatedAt, &u.CreatedBy, &u.UpdatedBy,
		)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		if err != nil {
			return err
		}
		user = &u
		return nil
	})

	return user, err
}

func (r *PgUserRepository) FindByCode(ctx context.Context, code string) (*users.User, error) {
	var user *users.User
	query := `
		SELECT usr_id, usr_code, usr_name, usr_email, usr_password, usr_phone, 
		       usr_status, usr_default_org, usr_notused, usr_tenant_id,
		       usr_created_at, usr_updated_at, usr_created_by, usr_updated_by
		FROM eamusers WHERE usr_code = $1`

	err := infraDB.RunInTenantTx(ctx, r.pool, func(tx pgx.Tx) error {
		var u users.User
		err := tx.QueryRow(ctx, query, code).Scan(
			&u.ID, &u.Code, &u.Name, &u.Email, &u.Password, &u.Phone,
			&u.Status, &u.DefaultOrg, &u.NotUsed, &u.TenantID,
			&u.CreatedAt, &u.UpdatedAt, &u.CreatedBy, &u.UpdatedBy,
		)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		if err != nil {
			return err
		}
		user = &u
		return nil
	})

	return user, err
}

func (r *PgUserRepository) FindByEmail(ctx context.Context, email string) (*users.User, error) {
	var user *users.User
	query := `
		SELECT usr_id, usr_code, usr_name, usr_email, usr_password, usr_phone, 
		       usr_status, usr_default_org, usr_notused, usr_tenant_id,
		       usr_created_at, usr_updated_at, usr_created_by, usr_updated_by
		FROM eamusers WHERE usr_email = $1`

	err := infraDB.RunInTenantTx(ctx, r.pool, func(tx pgx.Tx) error {
		var u users.User
		err := tx.QueryRow(ctx, query, email).Scan(
			&u.ID, &u.Code, &u.Name, &u.Email, &u.Password, &u.Phone,
			&u.Status, &u.DefaultOrg, &u.NotUsed, &u.TenantID,
			&u.CreatedAt, &u.UpdatedAt, &u.CreatedBy, &u.UpdatedBy,
		)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		if err != nil {
			return err
		}
		user = &u
		return nil
	})

	return user, err
}

func (r *PgUserRepository) FindAll(ctx context.Context, tenantID string, limit, offset int) ([]*users.User, int, error) {
	var total int
	var result []*users.User

	err := infraDB.RunInTenantTx(ctx, r.pool, func(tx pgx.Tx) error {
		// Count query
		countQuery := `SELECT COUNT(*) FROM eamusers WHERE usr_tenant_id = $1`
		if err := tx.QueryRow(ctx, countQuery, tenantID).Scan(&total); err != nil {
			return err
		}

		// Find all query
		query := `
			SELECT usr_id, usr_code, usr_name, usr_email, usr_password, usr_phone, 
			       usr_status, usr_default_org, usr_notused, usr_tenant_id,
			       usr_created_at, usr_updated_at, usr_created_by, usr_updated_by
			FROM eamusers 
			WHERE usr_tenant_id = $1
			ORDER BY usr_created_at DESC
			LIMIT $2 OFFSET $3`

		rows, err := tx.Query(ctx, query, tenantID, limit, offset)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var u users.User
			err := rows.Scan(
				&u.ID, &u.Code, &u.Name, &u.Email, &u.Password, &u.Phone,
				&u.Status, &u.DefaultOrg, &u.NotUsed, &u.TenantID,
				&u.CreatedAt, &u.UpdatedAt, &u.CreatedBy, &u.UpdatedBy,
			)
			if err != nil {
				return err
			}
			result = append(result, &u)
		}
		return rows.Err()
	})

	return result, total, err
}

func (r *PgUserRepository) Create(ctx context.Context, user *users.User) error {
	query := `
		INSERT INTO eamusers (usr_code, usr_name, usr_email, usr_password, 
		                      usr_phone, usr_status, usr_default_org, usr_notused,
		                      usr_tenant_id, usr_created_at, usr_updated_at, 
		                      usr_created_by, usr_updated_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING usr_id`

	return infraDB.RunInTenantTx(ctx, r.pool, func(tx pgx.Tx) error {
		var id int64
		err := tx.QueryRow(ctx, query,
			user.Code, user.Name, user.Email, user.Password, user.Phone,
			user.Status, user.DefaultOrg, user.NotUsed, user.TenantID,
			user.CreatedAt, user.UpdatedAt, user.CreatedBy, user.UpdatedBy,
		).Scan(&id)
		if err != nil {
			return err
		}
		user.ID = strconv.FormatInt(id, 10)
		return nil
	})
}

func (r *PgUserRepository) Update(ctx context.Context, user *users.User) error {
	query := `
		UPDATE eamusers 
		SET usr_name = $2, usr_email = $3, usr_phone = $4, usr_status = $5,
		    usr_default_org = $6, usr_notused = $7, usr_updated_at = $8, 
		    usr_updated_by = $9
		WHERE usr_id = $1`

	return infraDB.RunInTenantTx(ctx, r.pool, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, query,
			user.ID, user.Name, user.Email, user.Phone, user.Status,
			user.DefaultOrg, user.NotUsed, user.UpdatedAt, user.UpdatedBy,
		)
		return err
	})
}

func (r *PgUserRepository) Delete(ctx context.Context, id string) error {
	// Hard delete: if the user is referenced by other tables (e.g. eamuser_organizations),
	// the database will raise a foreign-key violation error intentionally.
	query := `DELETE FROM eamusers WHERE usr_id = $1`
	return infraDB.RunInTenantTx(ctx, r.pool, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, query, id)
		return err
	})
}
