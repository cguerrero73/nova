package queries

import (
	"context"
)

// Repository defines the interface for saved queries data access
type Repository interface {
	// List returns saved queries for a grid (public + user's own)
	List(ctx context.Context, gridID int, userID string) ([]*SavedQuery, error)

	// ListByGridName returns saved queries for a grid by name (public + user's own)
	ListByGridName(ctx context.Context, gridName string, userID string) ([]*SavedQuery, error)

	// GetByID returns a specific saved query
	GetByID(ctx context.Context, id string) (*SavedQuery, error)

	// Create saves a new query
	Create(ctx context.Context, query *SavedQuery) error

	// Update updates an existing query
	Update(ctx context.Context, query *SavedQuery) error

	// Delete removes a saved query
	Delete(ctx context.Context, id string) error

	// ClearDefaultForGrid removes default flag from all queries of a grid
	ClearDefaultForGrid(ctx context.Context, gridID int) error
}
