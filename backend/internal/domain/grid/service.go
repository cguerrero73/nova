package grid

import (
	"context"
	"strings"

	fieldsdomain "github.com/nova/backend/internal/domain/fields"
	queriesdomain "github.com/nova/backend/internal/domain/queries"
)

// Service handles grid business logic
type Service struct {
	repo        GridRepository
	fieldRepo   FieldsRepository
	queriesRepo QueriesRepository
}

// FieldsRepository interface for fields data access (domain-level interface)
type FieldsRepository interface {
	FindByGrid(ctx context.Context, baseQuery string) ([]*fieldsdomain.Field, error)
}

// QueriesRepository interface for saved queries data access
type QueriesRepository interface {
	GetByID(ctx context.Context, id string) (*queriesdomain.SavedQuery, error)
}

// NewService creates a new grid service
func NewService(repo GridRepository) *Service {
	return &Service{repo: repo}
}

// NewServiceWithFields creates a new grid service with field repository
func NewServiceWithFields(repo GridRepository, fieldRepo FieldsRepository) *Service {
	return &Service{repo: repo, fieldRepo: fieldRepo}
}

// NewServiceWithQueries creates a new grid service with queries repository
func NewServiceWithQueries(repo GridRepository, queriesRepo QueriesRepository) *Service {
	return &Service{repo: repo, queriesRepo: queriesRepo}
}

// NewServiceWithAll creates a new grid service with all repositories
func NewServiceWithAll(repo GridRepository, fieldRepo FieldsRepository, queriesRepo QueriesRepository) *Service {
	return &Service{repo: repo, fieldRepo: fieldRepo, queriesRepo: queriesRepo}
}

// formatLabel converts field name to human-readable label
// e.g., "usr_name" -> "Usr Name", "par_code" -> "Par Code"
func formatLabel(fieldName string) string {
	// Remove prefix (first 3 chars like usr_, par_, str_, etc.)
	if len(fieldName) > 4 {
		fieldName = fieldName[4:]
	}

	// Split by underscore and capitalize each word
	parts := strings.Split(fieldName, "_")
	for i, part := range parts {
		if len(part) > 0 {
			parts[i] = strings.ToUpper(string(part[0])) + strings.ToLower(part[1:])
		}
	}

	return strings.Join(parts, " ")
}

// GetConfig returns grid configuration metadata by name
func (s *Service) GetConfig(ctx context.Context, gridName string) (*GridConfig, error) {
	grid, err := s.repo.FindByName(ctx, gridName)
	if err != nil {
		return nil, err
	}
	if grid == nil {
		return nil, ErrGridNotFound
	}

	// Get field metadata in a single query
	var columns []GridCol
	if s.fieldRepo != nil {
		fields, err := s.fieldRepo.FindByGrid(ctx, grid.BaseQuery)
		if err == nil && fields != nil {
			for _, f := range fields {
				columns = append(columns, GridCol{
					ID:         f.ID,
					Key:        f.FieldName,
					Label:      formatLabel(f.FieldName),
					Type:       f.DataType,
					Sortable:   grid.IsFieldSortable(f.ID),
					Filterable: grid.IsFieldFilterable(f.ID),
				})
			}
		}
	}

	return &GridConfig{
		GridID:           grid.ID,
		GridName:         grid.Name,
		BaseQuery:        grid.BaseQuery,
		OrgColumn:        grid.OrgColumn,
		BotFunction:      grid.BotFunction,
		SecEntity:        grid.SecEntity,
		Hints:            grid.Hints,
		AvailableFilters: grid.GetFilterableFields(),
		AvailableSort:    grid.GetSortableFields(),
		AvailableDisplay: grid.GetDisplayableFields(),
		Columns:          columns,
	}, nil
}

// GetConfigByID returns grid configuration by ID
func (s *Service) GetConfigByID(ctx context.Context, gridID int) (*GridConfig, error) {
	grid, err := s.repo.FindByID(ctx, gridID)
	if err != nil {
		return nil, err
	}
	if grid == nil {
		return nil, ErrGridNotFound
	}

	// Get field metadata if fieldRepo is available
	var columns []GridCol
	if s.fieldRepo != nil {
		fields, err := s.fieldRepo.FindByGrid(ctx, grid.BaseQuery)
		if err == nil && fields != nil {
			for _, f := range fields {
				columns = append(columns, GridCol{
					ID:         f.ID,
					Key:        f.FieldName,
					Label:      formatLabel(f.FieldName),
					Type:       f.DataType,
					Sortable:   grid.IsFieldSortable(f.ID),
					Filterable: grid.IsFieldFilterable(f.ID),
				})
			}
		}
	}

	return &GridConfig{
		GridID:           grid.ID,
		GridName:         grid.Name,
		BaseQuery:        grid.BaseQuery,
		OrgColumn:        grid.OrgColumn,
		BotFunction:      grid.BotFunction,
		SecEntity:        grid.SecEntity,
		Hints:            grid.Hints,
		AvailableFilters: grid.GetFilterableFields(),
		AvailableSort:    grid.GetSortableFields(),
		AvailableDisplay: grid.GetDisplayableFields(),
		Columns:          columns,
	}, nil
}

// ExecuteQuery executes a grid query
func (s *Service) ExecuteQuery(ctx context.Context, gridID int, config *GridQueryConfig, page, pageSize int) (*GridResult, error) {
	// Default pagination
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}

	return s.repo.ExecuteQuery(ctx, "", nil, config.Filters, config.Sort, page, pageSize)
}

// GetGridByID returns a grid by ID
func (s *Service) GetGridByID(ctx context.Context, id int) (*Grid, error) {
	return s.repo.FindByID(ctx, id)
}

// GetGridByName returns a grid by name
func (s *Service) GetGridByName(ctx context.Context, name string) (*Grid, error) {
	return s.repo.FindByName(ctx, name)
}

// ExecuteQueryByID fetches saved query, resolves grid, and executes
func (s *Service) ExecuteQueryByID(ctx context.Context, queryID string, page, pageSize int) (*GridResult, error) {
	// Default pagination
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}

	// 1. Fetch SavedQuery by ID
	if s.queriesRepo == nil {
		return nil, ErrQueryNotFound
	}
	savedQuery, err := s.queriesRepo.GetByID(ctx, queryID)
	if err != nil {
		if strings.Contains(err.Error(), "query not found") {
			return nil, ErrQueryNotFound
		}
		return nil, err
	}

	// 2. Fetch Grid by gridId to get baseQuery
	grid, err := s.repo.FindByID(ctx, savedQuery.GridID)
	if err != nil {
		return nil, err
	}
	if grid == nil {
		return nil, ErrGridNotFound
	}

	// 3. Build config from saved query's fields/filters/sort
	config := &GridQueryConfig{
		Fields:  savedQuery.Query.Fields,
		Sort:    convertSort(savedQuery.Query.Sort),
		Filters: convertFilters(savedQuery.Query.Filters),
	}

	// 4. Call repo.ExecuteQuery with baseQuery from grid
	// Note: fields is passed as nil to select all columns (matching existing ExecuteQuery behavior)
	return s.repo.ExecuteQuery(ctx, grid.BaseQuery, nil, config.Filters, config.Sort, page, pageSize)
}

// convertSort converts queries.QuerySort to grid.SortCondition
func convertSort(sort []queriesdomain.QuerySort) []SortCondition {
	if len(sort) == 0 {
		return nil
	}
	result := make([]SortCondition, len(sort))
	for i, s := range sort {
		result[i] = SortCondition{
			Field:     s.Field,
			Direction: s.Direction,
		}
	}
	return result
}

// convertFilters converts queries.QueryFilter to grid.FilterCondition
func convertFilters(filters []queriesdomain.QueryFilter) []FilterCondition {
	if len(filters) == 0 {
		return nil
	}
	result := make([]FilterCondition, len(filters))
	for i, f := range filters {
		result[i] = FilterCondition{
			Field:    f.Field,
			Operator: f.Operator,
			Value:    f.Value,
		}
	}
	return result
}
