package fieldsdb

import (
	"context"
	"errors"
	"regexp"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	fieldsdomain "github.com/nova/backend/internal/domain/fields"
	infraDB "github.com/nova/backend/internal/infrastructure/db"
)

type PgFieldRepository struct {
	pool *pgxpool.Pool
}

func NewPgFieldRepository(pool *pgxpool.Pool) *PgFieldRepository {
	return &PgFieldRepository{pool: pool}
}

// FindByID returns a field by ID
func (r *PgFieldRepository) FindByID(ctx context.Context, id int) (*fieldsdomain.Field, error) {
	var field *fieldsdomain.Field
	query := `
		SELECT fld_id, fld_fieldname, fld_datatype, fld_tablename,
		       fld_created_at, fld_updated_at
		FROM eamfields WHERE fld_id = $1`

	err := infraDB.RunInTenantTx(ctx, r.pool, func(tx pgx.Tx) error {
		var f fieldsdomain.Field
		err := tx.QueryRow(ctx, query, id).Scan(
			&f.ID, &f.FieldName, &f.DataType, &f.TableName,
			&f.CreatedAt, &f.UpdatedAt,
		)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		if err != nil {
			return err
		}
		field = &f
		return nil
	})

	return field, err
}

// FindByTable returns all fields for a table
func (r *PgFieldRepository) FindByTable(ctx context.Context, tableName string) ([]*fieldsdomain.Field, error) {
	var result []*fieldsdomain.Field
	query := `
		SELECT fld_id, fld_fieldname, fld_datatype, fld_tablename,
		       fld_created_at, fld_updated_at
		FROM eamfields
		WHERE fld_tablename = $1
		ORDER BY fld_id ASC`

	err := infraDB.RunInTenantTx(ctx, r.pool, func(tx pgx.Tx) error {
		rows, err := tx.Query(ctx, query, tableName)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var f fieldsdomain.Field
			err := rows.Scan(
				&f.ID, &f.FieldName, &f.DataType, &f.TableName,
				&f.CreatedAt, &f.UpdatedAt,
			)
			if err != nil {
				return err
			}
			result = append(result, &f)
		}
		return rows.Err()
	})

	return result, err
}

// FindByGrid returns all fields for a grid based on its base query
// Extracts table names from the FROM clause and fetches all in ONE query
func (r *PgFieldRepository) FindByGrid(ctx context.Context, baseQuery string) ([]*fieldsdomain.Field, error) {
	// Extract table names from the base query
	tables := extractTables(baseQuery)
	if len(tables) == 0 {
		return nil, nil
	}

	var result []*fieldsdomain.Field

	// Single query with IN clause instead of multiple queries
	query := `
		SELECT fld_id, fld_fieldname, fld_datatype, fld_tablename,
		       fld_created_at, fld_updated_at
		FROM eamfields
		WHERE fld_tablename = ANY($1)
		ORDER BY fld_id ASC`

	err := infraDB.RunInTenantTx(ctx, r.pool, func(tx pgx.Tx) error {
		rows, err := tx.Query(ctx, query, tables)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var f fieldsdomain.Field
			err := rows.Scan(
				&f.ID, &f.FieldName, &f.DataType, &f.TableName,
				&f.CreatedAt, &f.UpdatedAt,
			)
			if err != nil {
				return err
			}
			result = append(result, &f)
		}
		return rows.Err()
	})

	return result, err
}

// FindByTableAndField returns a specific field
func (r *PgFieldRepository) FindByTableAndField(ctx context.Context, tableName, fieldName string) (*fieldsdomain.Field, error) {
	var field *fieldsdomain.Field
	query := `
		SELECT fld_id, fld_fieldname, fld_datatype, fld_tablename,
		       fld_created_at, fld_updated_at
		FROM eamfields
		WHERE fld_tablename = $1 AND fld_fieldname = $2`

	err := infraDB.RunInTenantTx(ctx, r.pool, func(tx pgx.Tx) error {
		var f fieldsdomain.Field
		err := tx.QueryRow(ctx, query, tableName, fieldName).Scan(
			&f.ID, &f.FieldName, &f.DataType, &f.TableName,
			&f.CreatedAt, &f.UpdatedAt,
		)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		if err != nil {
			return err
		}
		field = &f
		return nil
	})

	return field, err
}

// extractTables extracts table names from a SQL FROM clause
// Handles: FROM table, FROM table alias, FROM table1, table2, FROM table1 JOIN table2
func extractTables(query string) []string {
	// Normalize to lowercase for parsing
	q := strings.ToLower(query)

	// Find FROM keyword
	fromIdx := strings.Index(q, "from")
	if fromIdx == -1 {
		return nil
	}

	// Get content after FROM
	content := strings.TrimSpace(q[fromIdx+4:])

	// Remove WHERE, GROUP, ORDER, LIMIT clauses
	for _, keyword := range []string{"where", "group by", "order by", "limit", "union"} {
		idx := strings.Index(content, keyword)
		if idx != -1 {
			content = content[:idx]
		}
	}

	// Split by comma and JOIN
	// First, replace JOIN variations with comma
	joinPattern := regexp.MustCompile(`\s+(left|right|inner|outer|cross)?\s*join\s+`)
	content = joinPattern.ReplaceAllString(content, ",")

	// Split by comma
	parts := strings.Split(content, ",")
	var tables []string

	for _, part := range parts {
		part = strings.TrimSpace(part)

		// Extract table name (before alias or whitespace)
		words := strings.Fields(part)
		if len(words) > 0 {
			tableName := words[0]
			// Remove any backticks, quotes
			tableName = strings.Trim(tableName, "`\"'")

			if tableName != "" && !strings.HasPrefix(tableName, "(") {
				tables = append(tables, tableName)
			}
		}
	}

	return tables
}
