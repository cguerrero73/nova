package queries

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nova/backend/internal/domain/queries"
	infraDB "github.com/nova/backend/internal/infrastructure/db"
)

type PgQueryRepository struct {
	pool *pgxpool.Pool
}

func NewPgQueryRepository(pool *pgxpool.Pool) *PgQueryRepository {
	return &PgQueryRepository{pool: pool}
}

func (r *PgQueryRepository) List(ctx context.Context, gridID int, userID string) ([]*queries.SavedQuery, error) {
	var result []*queries.SavedQuery

	err := infraDB.RunInTenantTx(ctx, r.pool, func(tx pgx.Tx) error {
		query := `
			SELECT qry_id, qry_grid_id, qry_name, qry_user_id, qry_is_public, 
			       qry_is_default, qry_query, qry_created_at, qry_updated_at
			FROM eamqueries
			WHERE qry_grid_id = $1 
			  AND (qry_is_public = true OR qry_user_id = $2)
			ORDER BY qry_is_default DESC, qry_name ASC`

		rows, err := tx.Query(ctx, query, gridID, userID)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var q queries.SavedQuery
			var queryJSON []byte
			err := rows.Scan(
				&q.ID, &q.GridID, &q.Name, &q.UserID, &q.IsPublic, &q.IsDefault,
				&queryJSON, &q.CreatedAt, &q.UpdatedAt,
			)
			if err != nil {
				return err
			}

			if len(queryJSON) > 0 {
				if err := json.Unmarshal(queryJSON, &q.Query); err != nil {
					return err
				}
			}

			result = append(result, &q)
		}
		return rows.Err()
	})

	return result, err
}

func (r *PgQueryRepository) GetByID(ctx context.Context, id string) (*queries.SavedQuery, error) {
	err := infraDB.RunInTenantTx(ctx, r.pool, func(tx pgx.Tx) error {
		sql := `
			SELECT qry_id, qry_grid_id, qry_name, qry_user_id, qry_is_public,
			       qry_is_default, qry_query, qry_created_at, qry_updated_at
			FROM eamqueries WHERE qry_id = $1`

		var q queries.SavedQuery
		var queryJSON []byte
		err := tx.QueryRow(ctx, sql, id).Scan(
			&q.ID, &q.GridID, &q.Name, &q.UserID, &q.IsPublic, &q.IsDefault,
			&queryJSON, &q.CreatedAt, &q.UpdatedAt,
		)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		if err != nil {
			return err
		}

		if len(queryJSON) > 0 {
			if err := json.Unmarshal(queryJSON, &q.Query); err != nil {
				return err
			}
		}

		return nil
	})

	return nil, err
}

func (r *PgQueryRepository) Create(ctx context.Context, q *queries.SavedQuery) error {
	queryJSON, err := json.Marshal(q.Query)
	if err != nil {
		return err
	}

	sql := `
		INSERT INTO eamqueries (qry_id, qry_grid_id, qry_name, qry_user_id, qry_is_public,
		                        qry_is_default, qry_query, qry_created_at, qry_updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	return infraDB.RunInTenantTx(ctx, r.pool, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, sql,
			q.ID, q.GridID, q.Name, q.UserID, q.IsPublic, q.IsDefault,
			queryJSON, q.CreatedAt, q.UpdatedAt,
		)
		return err
	})
}

func (r *PgQueryRepository) Update(ctx context.Context, q *queries.SavedQuery) error {
	queryJSON, err := json.Marshal(q.Query)
	if err != nil {
		return err
	}

	sql := `
		UPDATE eamqueries 
		SET qry_name = $2, qry_is_public = $3, qry_is_default = $4, 
		    qry_query = $5, qry_updated_at = $6
		WHERE qry_id = $1`

	return infraDB.RunInTenantTx(ctx, r.pool, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, sql,
			q.ID, q.Name, q.IsPublic, q.IsDefault, queryJSON, q.UpdatedAt,
		)
		return err
	})
}

func (r *PgQueryRepository) Delete(ctx context.Context, id string) error {
	return infraDB.RunInTenantTx(ctx, r.pool, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, "DELETE FROM eamqueries WHERE qry_id = $1", id)
		return err
	})
}

func (r *PgQueryRepository) ClearDefaultForGrid(ctx context.Context, gridID int) error {
	return infraDB.RunInTenantTx(ctx, r.pool, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx,
			"UPDATE eamqueries SET qry_is_default = false WHERE qry_grid_id = $1",
			gridID,
		)
		return err
	})
}
