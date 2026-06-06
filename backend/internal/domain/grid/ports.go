package grid

import "context"

// GridRepository defines the interface for grid data access
type GridRepository interface {
	// FindByID returns a grid by ID
	FindByID(ctx context.Context, id int) (*Grid, error)

	// FindByName returns a grid by name
	FindByName(ctx context.Context, name string) (*Grid, error)

	// FindAll returns all grids
	FindAll(ctx context.Context) ([]*Grid, error)

	// ExecuteQuery executes a custom query with parameters
	ExecuteQuery(ctx context.Context, baseQuery string, fields []string,
		filters []FilterCondition, sort []SortCondition, page, pageSize int) (*GridResult, error)
}
