package grid

// FilterCondition represents a filter condition
type FilterCondition struct {
	Field    string `json:"field"`    // field name, not ID
	Operator int    `json:"operator"` // 1=EQ, 2=NE, 3=CONTAINS, 4=GT, 5=LT, 6=GTE, 7=LTE, 8=IN, 9=IS_NULL, 10=IS_NOT_NULL
	Value    any    `json:"value"`
}

// SortCondition represents sorting configuration
type SortCondition struct {
	Field     string `json:"field"`     // field name, not ID
	Direction int    `json:"direction"` // 1=ASC, 2=DESC
}

// GridQueryConfig holds query configuration from SavedQuery
type GridQueryConfig struct {
	Fields     []int             `json:"fields"`
	Sort       []SortCondition   `json:"sort"`
	Filters    []FilterCondition `json:"filters"`
	Pagination Pagination        `json:"pagination"`
}

// Pagination holds pagination settings
type Pagination struct {
	PageSize int `json:"pageSize"`
}

// GridMeta represents pagination metadata
type GridMeta struct {
	Page       int `json:"page"`
	PageSize   int `json:"pageSize"`
	Total      int `json:"total"`
	TotalPages int `json:"totalPages"`
}

// GridResult holds query results
type GridResult struct {
	Data     []map[string]any `json:"data"`
	Total    int              `json:"total"`
	Page     int              `json:"page"`
	PageSize int              `json:"pageSize"`
}

// GridResponse represents the API response for grid data
type GridResponse struct {
	Success bool        `json:"success"`
	Data    []any       `json:"data"`
	Meta    GridMeta    `json:"meta"`
	Columns []GridCol   `json:"columns,omitempty"`
	Config  *GridConfig `json:"gridConfig,omitempty"`
}

// GridCol represents a column in the grid
type GridCol struct {
	ID         int    `json:"id"`
	Key        string `json:"key"`
	Label      string `json:"label"`
	Type       string `json:"type"` // string, number, date, boolean, select
	Sortable   bool   `json:"sortable"`
	Filterable bool   `json:"filterable"`
}

// GridColumnRef pairs a database column name with its exposed domain key.
type GridColumnRef struct {
	DBName    string
	DomainKey string
}

// GridConfig represents grid configuration metadata
type GridConfig struct {
	GridID           int       `json:"gridId"`
	GridName         string    `json:"gridName"`
	BaseQuery        string    `json:"baseQuery"`
	OrgColumn        string    `json:"orgColumn"`
	BotFunction      string    `json:"botFunction"`
	SecEntity        string    `json:"secEntity"`
	Hints            string    `json:"hints"`
	AvailableFilters []int     `json:"availableFilters"`
	AvailableSort    []int     `json:"availableSort"`
	AvailableDisplay []int     `json:"availableDisplay"`
	Columns          []GridCol `json:"columns,omitempty"`
}

// GridConfigResponse represents the API response for grid config
type GridConfigResponse struct {
	Success bool        `json:"success"`
	Config  *GridConfig `json:"config"`
}
