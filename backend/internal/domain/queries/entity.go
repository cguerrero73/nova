package queries

import (
	"encoding/json"
	"time"
)

// QueryFilter represents a filter condition
type QueryFilter struct {
	Field    int `json:"field"`
	Operator int `json:"operator"`
	Value    any `json:"value"`
}

// QuerySort represents sorting configuration
type QuerySort struct {
	Field     int `json:"field"`
	Direction int `json:"direction"` // 1=ASC, 2=DESC
}

// GridQuery represents the query configuration
type GridQuery struct {
	Fields     []int         `json:"fields"`
	Sort       []QuerySort   `json:"sort"`
	Filters    []QueryFilter `json:"filters"`
	Pagination Pagination    `json:"pagination"`
}

// MarshalJSON implements custom JSON marshaling for GridQuery
func (q GridQuery) MarshalJSON() ([]byte, error) {
	type alias GridQuery
	return json.Marshal(alias(q))
}

// UnmarshalJSON implements custom JSON unmarshaling for GridQuery
func (q *GridQuery) UnmarshalJSON(data []byte) error {
	type alias GridQuery
	var a alias
	if err := json.Unmarshal(data, &a); err != nil {
		return err
	}
	*q = GridQuery(a)
	return nil
}

// Pagination holds pagination settings
type Pagination struct {
	PageSize int `json:"pageSize"`
}

// SavedQuery represents a saved query
type SavedQuery struct {
	ID        string    `json:"id"`
	GridID    int       `json:"gridId"`
	Name      string    `json:"name"`
	UserID    *string   `json:"userId"`
	IsPublic  bool      `json:"isPublic"`
	IsDefault bool      `json:"isDefault"`
	Query     GridQuery `json:"query"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// GridRequest represents the request to execute a grid query
type GridRequest struct {
	GridID int       `json:"gridId"`
	Query  GridQuery `json:"query"`
	Page   int       `json:"page,omitempty"`
}

// GridMeta represents pagination metadata
type GridMeta struct {
	Page       int `json:"page"`
	PageSize   int `json:"pageSize"`
	Total      int `json:"total"`
	TotalPages int `json:"totalPages"`
}

// GridResponse represents the response for grid data
type GridResponse struct {
	Success bool     `json:"success"`
	Data    []any    `json:"data"`
	Meta    GridMeta `json:"meta"`
}

// ExecuteRequest represents the request to execute a custom query
type ExecuteRequest struct {
	Query string `json:"query"`
	Code  string `json:"code"`
}
