package fields

import "context"

// FieldRepository defines the interface for fields data access
type FieldRepository interface {
	// FindByID returns a field by ID
	FindByID(ctx context.Context, id int) (*Field, error)

	// FindByTable returns all fields for a table
	FindByTable(ctx context.Context, tableName string) ([]*Field, error)

	// FindByGrid returns all fields for a grid's tables
	FindByGrid(ctx context.Context, baseQuery string) ([]*Field, error)

	// FindByTableAndField returns a specific field
	FindByTableAndField(ctx context.Context, tableName, fieldName string) (*Field, error)
}
