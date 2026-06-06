package fields

import "context"

// Service handles fields business logic
type Service struct {
	repo FieldRepository
}

// NewService creates a new fields service
func NewService(repo FieldRepository) *Service {
	return &Service{repo: repo}
}

// GetByTable returns all fields for a table
func (s *Service) GetByTable(ctx context.Context, tableName string) ([]*Field, error) {
	return s.repo.FindByTable(ctx, tableName)
}

// GetByGrid returns all fields needed for a grid based on its base query
func (s *Service) GetByGrid(ctx context.Context, baseQuery string) ([]*Field, error) {
	return s.repo.FindByGrid(ctx, baseQuery)
}

// GetField returns a specific field
func (s *Service) GetField(ctx context.Context, tableName, fieldName string) (*Field, error) {
	return s.repo.FindByTableAndField(ctx, tableName, fieldName)
}
