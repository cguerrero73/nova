package queries

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"

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

		log.Printf("[DEBUG] Query List: gridID=%d userID=%s sql=%s", gridID, userID, query)
		rows, err := tx.Query(ctx, query, gridID, userID)
		if err != nil {
			log.Printf("[ERROR] Query List failed: %v", err)
			return fmt.Errorf("query list: %w", err)
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
				log.Printf("[ERROR] Scan failed: %v", err)
				return fmt.Errorf("scan row: %w", err)
			}

			if len(queryJSON) > 0 {
				if err := json.Unmarshal(queryJSON, &q.Query); err != nil {
					log.Printf("[ERROR] JSON unmarshal failed: %v", err)
					return fmt.Errorf("unmarshal query: %w", err)
				}
			}

			result = append(result, &q)
		}
		if rows.Err() != nil {
			log.Printf("[ERROR] Rows iteration error: %v", rows.Err())
			return fmt.Errorf("rows error: %w", rows.Err())
		}
		return nil
	})

	if err != nil {
		log.Printf("[ERROR] List queries failed - gridID=%d userID=%s error=%v", gridID, userID, err)
	}
	return result, err
}

func (r *PgQueryRepository) ListByGridName(ctx context.Context, gridName string, userID string) ([]*queries.SavedQuery, error) {
	var result []*queries.SavedQuery

	err := infraDB.RunInTenantTx(ctx, r.pool, func(tx pgx.Tx) error {
		query := `
			SELECT q.qry_id, q.qry_grid_id, q.qry_name, q.qry_user_id, q.qry_is_public, 
			       q.qry_is_default, q.qry_query, q.qry_created_at, q.qry_updated_at
			FROM eamqueries q
			INNER JOIN eamgrids g ON g.grd_id = q.qry_grid_id
			WHERE g.grd_name = $1 
			  AND (q.qry_is_public = true OR q.qry_user_id = $2)
			ORDER BY q.qry_is_default DESC, q.qry_name ASC`

		rows, err := tx.Query(ctx, query, gridName, userID)
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
	var q queries.SavedQuery
	var queryJSON []byte

	err := infraDB.RunInTenantTx(ctx, r.pool, func(tx pgx.Tx) error {
		sql := `
			SELECT qry_id, qry_grid_id, qry_name, qry_user_id, qry_is_public,
			       qry_is_default, qry_query, qry_created_at, qry_updated_at
			FROM eamqueries WHERE qry_id = $1`

		err := tx.QueryRow(ctx, sql, id).Scan(
			&q.ID, &q.GridID, &q.Name, &q.UserID, &q.IsPublic, &q.IsDefault,
			&queryJSON, &q.CreatedAt, &q.UpdatedAt,
		)
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("query not found: %w", err)
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

	return &q, err
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
