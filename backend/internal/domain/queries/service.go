package queries

import (
	"context"
	"time"
)

// Service handles saved queries business logic
type Service struct {
	repo QueryRepository
}

// QueryRepository defines the interface (same as domain repo)
type QueryRepository interface {
	List(ctx context.Context, gridID int, userID string) ([]*SavedQuery, error)
	ListByGridName(ctx context.Context, gridName string, userID string) ([]*SavedQuery, error)
	GetByID(ctx context.Context, id string) (*SavedQuery, error)
	Create(ctx context.Context, query *SavedQuery) error
	Update(ctx context.Context, query *SavedQuery) error
	Delete(ctx context.Context, id string) error
	ClearDefaultForGrid(ctx context.Context, gridID int) error
}

// NewService creates a new queries service
func NewService(repo QueryRepository) *Service {
	return &Service{repo: repo}
}

// List returns saved queries for a grid by ID
func (s *Service) List(ctx context.Context, gridID int, userID string) ([]*SavedQuery, error) {
	return s.repo.List(ctx, gridID, userID)
}

// ListByGridName returns saved queries for a grid by name
func (s *Service) ListByGridName(ctx context.Context, gridName string, userID string) ([]*SavedQuery, error) {
	return s.repo.ListByGridName(ctx, gridName, userID)
}

// GetByID returns a saved query by ID
func (s *Service) GetByID(ctx context.Context, id string) (*SavedQuery, error) {
	return s.repo.GetByID(ctx, id)
}

// Create creates a new saved query
func (s *Service) Create(ctx context.Context, req *SaveRequest) (*SavedQuery, error) {
	// If isDefault, clear other defaults first
	if req.IsDefault {
		if err := s.repo.ClearDefaultForGrid(ctx, req.GridID); err != nil {
			return nil, err
		}
	}

	now := time.Now()
	query := &SavedQuery{
		ID:        generateID(),
		GridID:    req.GridID,
		Name:      req.Name,
		UserID:    req.UserID,
		IsPublic:  req.IsPublic,
		IsDefault: req.IsDefault,
		Query:     req.Query,
		CreatedAt: &now,
		UpdatedAt: &now,
	}

	if err := s.repo.Create(ctx, query); err != nil {
		return nil, err
	}

	return query, nil
}

// Update updates a saved query
func (s *Service) Update(ctx context.Context, id string, req *UpdateRequest) (*SavedQuery, error) {
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// If setting as default, clear others first
	if req.IsDefault != nil && *req.IsDefault && !existing.IsDefault {
		if err := s.repo.ClearDefaultForGrid(ctx, existing.GridID); err != nil {
			return nil, err
		}
	}

	// Apply updates
	if req.Name != nil {
		existing.Name = *req.Name
	}
	if req.IsPublic != nil {
		existing.IsPublic = *req.IsPublic
	}
	if req.IsDefault != nil {
		existing.IsDefault = *req.IsDefault
	}
	if req.Query != nil {
		existing.Query = *req.Query
	}
	existing.UpdatedAt = func() *time.Time { t := time.Now(); return &t }()

	if err := s.repo.Update(ctx, existing); err != nil {
		return nil, err
	}

	return existing, nil
}

// Delete removes a saved query
func (s *Service) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

// SaveRequest represents the request to save a query
type SaveRequest struct {
	GridID    int
	Name      string
	UserID    *string
	IsPublic  bool
	IsDefault bool
	Query     GridQuery
}

// UpdateRequest represents the request to update a query
type UpdateRequest struct {
	Name      *string
	IsPublic  *bool
	IsDefault *bool
	Query     *GridQuery
}

// generateID generates a unique ID for queries
func generateID() string {
	return "qry-" + randomString(16)
}

// randomString generates a random alphanumeric string
func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[time.Now().UnixNano()%int64(len(letters))]
	}
	return string(b)
}
