package griddb

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	griddomain "github.com/nova/backend/internal/domain/grid"
	infraDB "github.com/nova/backend/internal/infrastructure/db"
)

type PgGridRepository struct {
	pool *pgxpool.Pool
}

func NewPgGridRepository(pool *pgxpool.Pool) *PgGridRepository {
	return &PgGridRepository{pool: pool}
}

// FindByID returns a grid by ID
func (r *PgGridRepository) FindByID(ctx context.Context, id int) (*griddomain.Grid, error) {
	var grid *griddomain.Grid
	query := `
		SELECT grd_id, grd_name, grd_desc, grd_base_query,
		       COALESCE(grd_key_fields, ''), COALESCE(grd_filterable_list, ''), COALESCE(grd_sortable_list, ''), COALESCE(grd_displayable_list, ''),
		       COALESCE(grd_org_column, ''), COALESCE(grd_bot_function, ''), COALESCE(grd_sec_entity, ''), COALESCE(grd_hints, ''), COALESCE(grd_type, 0),
		       COALESCE(grd_created_at, NOW()), COALESCE(grd_updated_at, NOW())
		FROM eamgrids WHERE grd_id = $1`

	err := infraDB.RunInTenantTx(ctx, r.pool, func(tx pgx.Tx) error {
		var g griddomain.Grid
		err := tx.QueryRow(ctx, query, id).Scan(
			&g.ID, &g.Name, &g.Description, &g.BaseQuery,
			&g.KeyFields, &g.FilterableList, &g.SortableList, &g.DisplayableList,
			&g.OrgColumn, &g.BotFunction, &g.SecEntity, &g.Hints, &g.GridType,
			&g.CreatedAt, &g.UpdatedAt,
		)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		if err != nil {
			return err
		}
		grid = &g
		return nil
	})

	return grid, err
}

// FindByName returns a grid by name
func (r *PgGridRepository) FindByName(ctx context.Context, name string) (*griddomain.Grid, error) {
	var grid *griddomain.Grid
	query := `
		SELECT grd_id, grd_name, grd_desc, grd_base_query,
		       COALESCE(grd_key_fields, ''), COALESCE(grd_filterable_list, ''), COALESCE(grd_sortable_list, ''), COALESCE(grd_displayable_list, ''),
		       COALESCE(grd_org_column, ''), COALESCE(grd_bot_function, ''), COALESCE(grd_sec_entity, ''), COALESCE(grd_hints, ''), COALESCE(grd_type, 0),
		       COALESCE(grd_created_at, NOW()), COALESCE(grd_updated_at, NOW())
		FROM eamgrids WHERE grd_name = $1`

	err := infraDB.RunInTenantTx(ctx, r.pool, func(tx pgx.Tx) error {
		var g griddomain.Grid
		err := tx.QueryRow(ctx, query, name).Scan(
			&g.ID, &g.Name, &g.Description, &g.BaseQuery,
			&g.KeyFields, &g.FilterableList, &g.SortableList, &g.DisplayableList,
			&g.OrgColumn, &g.BotFunction, &g.SecEntity, &g.Hints, &g.GridType,
			&g.CreatedAt, &g.UpdatedAt,
		)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		if err != nil {
			return err
		}
		grid = &g
		return nil
	})

	return grid, err
}

// FindAll returns all grids
func (r *PgGridRepository) FindAll(ctx context.Context) ([]*griddomain.Grid, error) {
	var result []*griddomain.Grid
	query := `
		SELECT grd_id, grd_name, grd_desc, grd_base_query,
		       COALESCE(grd_key_fields, ''), COALESCE(grd_filterable_list, ''), COALESCE(grd_sortable_list, ''), COALESCE(grd_displayable_list, ''),
		       COALESCE(grd_org_column, ''), COALESCE(grd_bot_function, ''), COALESCE(grd_sec_entity, ''), COALESCE(grd_hints, ''), COALESCE(grd_type, 0),
		       COALESCE(grd_created_at, NOW()), COALESCE(grd_updated_at, NOW())
		FROM eamgrids ORDER BY grd_name ASC`

	err := infraDB.RunInTenantTx(ctx, r.pool, func(tx pgx.Tx) error {
		rows, err := tx.Query(ctx, query)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var g griddomain.Grid
			err := rows.Scan(
				&g.ID, &g.Name, &g.Description, &g.BaseQuery,
				&g.KeyFields, &g.FilterableList, &g.SortableList, &g.DisplayableList,
				&g.OrgColumn, &g.BotFunction, &g.SecEntity, &g.Hints, &g.GridType,
				&g.CreatedAt, &g.UpdatedAt,
			)
			if err != nil {
				return err
			}
			result = append(result, &g)
		}
		return rows.Err()
	})

	return result, err
}

// ExecuteQuery executes a query with the given base query and parameters
func (r *PgGridRepository) ExecuteQuery(ctx context.Context, baseQuery string, columns []string,
	filters []griddomain.FilterCondition, sort []griddomain.SortCondition, page, pageSize int) (*griddomain.GridResult, error) {

	var result *griddomain.GridResult

	err := infraDB.RunInTenantTx(ctx, r.pool, func(tx pgx.Tx) error {
		// Build SELECT clause - use columns if provided, else *
		selectClause := "*"
		if len(columns) > 0 {
			selectClause = strings.Join(columns, ", ")
		}

		// Build WHERE clause from filters
		whereClause := ""
		args := make([]any, 0)
		argIndex := 1

		for i, f := range filters {
			if i == 0 {
				whereClause = " WHERE "
			} else {
				whereClause += " AND "
			}

			switch f.Operator {
			case griddomain.OP_EQ:
				whereClause += fmt.Sprintf("%s = $%d", f.FieldID, argIndex)
				args = append(args, f.Value)
				argIndex++
			case griddomain.OP_NE:
				whereClause += fmt.Sprintf("%s != $%d", f.FieldID, argIndex)
				args = append(args, f.Value)
				argIndex++
			case griddomain.OP_CONTAINS:
				whereClause += fmt.Sprintf("%s ILIKE $%d", f.FieldID, argIndex)
				args = append(args, "%"+fmt.Sprintf("%v", f.Value)+"%")
				argIndex++
			case griddomain.OP_GT:
				whereClause += fmt.Sprintf("%s > $%d", f.FieldID, argIndex)
				args = append(args, f.Value)
				argIndex++
			case griddomain.OP_LT:
				whereClause += fmt.Sprintf("%s < $%d", f.FieldID, argIndex)
				args = append(args, f.Value)
				argIndex++
			case griddomain.OP_GTE:
				whereClause += fmt.Sprintf("%s >= $%d", f.FieldID, argIndex)
				args = append(args, f.Value)
				argIndex++
			case griddomain.OP_LTE:
				whereClause += fmt.Sprintf("%s <= $%d", f.FieldID, argIndex)
				args = append(args, f.Value)
				argIndex++
			case griddomain.OP_IN:
				whereClause += fmt.Sprintf("%s = ANY($%d)", f.FieldID, argIndex)
				args = append(args, f.Value)
				argIndex++
			case griddomain.OP_IS_NULL:
				whereClause += fmt.Sprintf("%s IS NULL", f.FieldID)
			case griddomain.OP_IS_NOT_NULL:
				whereClause += fmt.Sprintf("%s IS NOT NULL", f.FieldID)
			}
		}

		// Count total
		countQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s%s", baseQuery, whereClause)
		var total int
		if err := tx.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
			return err
		}

		// Build ORDER BY clause
		orderClause := ""
		for i, s := range sort {
			if i == 0 {
				orderClause = " ORDER BY "
			} else {
				orderClause += ", "
			}

			dir := "ASC"
			if s.Direction == griddomain.SORT_DESC {
				dir = "DESC"
			}
			orderClause += fmt.Sprintf("%s %s", s.FieldID, dir)
		}

		// Build LIMIT/OFFSET
		offset := (page - 1) * pageSize
		limitOffset := fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
		args = append(args, pageSize, offset)

		// Full query
		fullQuery := fmt.Sprintf("SELECT %s FROM %s%s%s%s",
			selectClause, baseQuery, whereClause, orderClause, limitOffset)

		// Execute query
		rows, err := tx.Query(ctx, fullQuery, args...)
		if err != nil {
			return err
		}
		defer rows.Close()

		// Get column names from result
		colDescs := rows.FieldDescriptions()
		colNames := make([]string, len(colDescs))
		for i, col := range colDescs {
			colNames[i] = string(col.Name)
		}

		// Fetch rows
		data := make([]map[string]any, 0)
		for rows.Next() {
			values, err := rows.Values()
			if err != nil {
				return err
			}

			row := make(map[string]any)
			for i, val := range values {
				row[colNames[i]] = val
			}
			data = append(data, row)
		}

		if err := rows.Err(); err != nil {
			return err
		}

		result = &griddomain.GridResult{
			Data:     data,
			Total:    total,
			Page:     page,
			PageSize: pageSize,
		}
		return nil
	})

	return result, err
}
