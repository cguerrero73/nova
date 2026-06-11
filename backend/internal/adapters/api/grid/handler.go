package api

import (
	"log"

	"github.com/gofiber/fiber/v2"

	griddomain "github.com/nova/backend/internal/domain/grid"
	"github.com/nova/backend/pkg/errors"
)

type GridHandler struct {
	gridService *griddomain.Service
}

func NewGridHandler(gridService *griddomain.Service) *GridHandler {
	return &GridHandler{gridService: gridService}
}

// GridDataRequest represents the request to execute a grid query
type GridDataRequest struct {
	GridID    int                          `json:"gridId"`
	QueryID   string                       `json:"queryId,omitempty"`
	FirstLoad bool                         `json:"firstLoad"`
	Fields    []int                        `json:"fields"`
	Sort      []griddomain.SortCondition   `json:"sort"`
	Filters   []griddomain.FilterCondition `json:"filters"`
	Page      int                          `json:"page"`
	PageSize  int                          `json:"pageSize"`
}

// GetConfig handles GET /grid/config/:name
// Returns grid configuration metadata
func (h *GridHandler) GetConfig(c *fiber.Ctx) error {
	name := c.Params("name")
	if name == "" {
		return c.Status(400).JSON(errors.New("BAD_REQUEST", "grid name is required", 400))
	}

	log.Printf("[GridConfig] Getting config for grid: %s", name)

	// Use c.UserContext() instead of c.Context() to get the Go context with tenant
	ctx := c.UserContext()
	log.Printf("[GridConfig] ctx type=%T", ctx)

	config, err := h.gridService.GetConfig(ctx, name)
	if err != nil {
		log.Printf("[GridConfig] ERROR getting config for '%s': %v", name, err)
		return c.Status(500).JSON(errors.NewWithDetail(
			"INTERNAL",
			"Error getting grid config",
			err.Error(),
			500,
		))
	}
	if config == nil {
		log.Printf("[GridConfig] Grid not found: %s", name)
		return c.Status(404).JSON(errors.New("NOT_FOUND", "Grid not found: "+name, 404))
	}

	log.Printf("[GridConfig] Success for '%s': %d columns", name, len(config.Columns))
	return c.JSON(griddomain.GridConfigResponse{
		Success: true,
		Config:  config,
	})
}

// GetConfigByID handles GET /grid/config/id/:id
// Returns grid configuration by ID
func (h *GridHandler) GetConfigByID(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(400).JSON(errors.New("BAD_REQUEST", "grid id is required", 400))
	}

	var gridID int
	if _, err := parseIntParam(id, &gridID); err != nil {
		return c.Status(400).JSON(errors.New("BAD_REQUEST", "invalid grid id", 400))
	}

	config, err := h.gridService.GetConfigByID(c.UserContext(), gridID)
	if err != nil {
		return c.Status(500).JSON(errors.ErrInternal)
	}
	if config == nil {
		return c.Status(404).JSON(errors.New("NOT_FOUND", "Grid not found", 404))
	}

	return c.JSON(griddomain.GridConfigResponse{
		Success: true,
		Config:  config,
	})
}

// ExecuteData handles POST /grid/data
// Executes a grid query
func (h *GridHandler) ExecuteData(c *fiber.Ctx) error {
	var req GridDataRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(errors.ErrBadRequest)
	}

	// Default pagination
	page := req.Page
	if page < 1 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize < 1 {
		pageSize = 20
	}

	// Use c.UserContext() to get Go context with tenant
	ctx := c.UserContext()

	// Route based on presence of QueryID
	if req.QueryID != "" {
		// Execute by saved query ID
		result, err := h.gridService.ExecuteQueryByID(ctx, req.QueryID, page, pageSize)
		if err != nil {
			log.Printf("[ExecuteData] ERROR QueryID=%s page=%d: %v", req.QueryID, page, err)
			if err == griddomain.ErrQueryNotFound {
				return c.Status(404).JSON(errors.New("NOT_FOUND", "Query not found", 404))
			}
			if err == griddomain.ErrGridNotFound {
				return c.Status(404).JSON(errors.New("NOT_FOUND", "Grid not found", 404))
			}
			return c.Status(500).JSON(errors.NewWithDetail("INTERNAL", "Error executing query by ID", err.Error(), 500))
		}

		totalPages := result.Total / result.PageSize
		if result.Total%result.PageSize > 0 {
			totalPages++
		}
		return c.JSON(griddomain.GridResponse{
			Success: true,
			Data:    convertToSliceAny(result.Data),
			Meta: griddomain.GridMeta{
				Page:       result.Page,
				PageSize:   result.PageSize,
				Total:      result.Total,
				TotalPages: totalPages,
			},
		})
	}

	// Fallback: validate gridId required when no QueryID
	if req.GridID == 0 {
		return c.Status(400).JSON(errors.New("BAD_REQUEST", "gridId is required when queryId not provided", 400))
	}

	// Existing flat-field execution (backward compatible)
	queryConfig := &griddomain.GridQueryConfig{
		Fields:  req.Fields,
		Sort:    req.Sort,
		Filters: req.Filters,
		Pagination: griddomain.Pagination{
			PageSize: pageSize,
		},
	}

	log.Printf("[ExecuteData] GridID=%d page=%d pageSize=%d filters=%d", req.GridID, page, pageSize, len(req.Filters))

	result, err := h.gridService.ExecuteQuery(ctx, req.GridID, queryConfig, page, pageSize)
	if err != nil {
		log.Printf("[ExecuteData] ERROR GridID=%d page=%d: %v", req.GridID, page, err)
		if err == griddomain.ErrGridNotFound {
			return c.Status(404).JSON(errors.New("NOT_FOUND", "Grid not found", 404))
		}
		return c.Status(500).JSON(errors.NewWithDetail("INTERNAL", "Error executing grid query", err.Error(), 500))
	}

	// Calculate total pages
	totalPages := result.Total / result.PageSize
	if result.Total%result.PageSize > 0 {
		totalPages++
	}

	return c.JSON(griddomain.GridResponse{
		Success: true,
		Data:    convertToSliceAny(result.Data),
		Meta: griddomain.GridMeta{
			Page:       result.Page,
			PageSize:   result.PageSize,
			Total:      result.Total,
			TotalPages: totalPages,
		},
	})
}

// convertToSliceAny converts []map[string]any to []any for JSON serialization
func convertToSliceAny(data []map[string]any) []any {
	result := make([]any, len(data))
	for i, m := range data {
		result[i] = m
	}
	return result
}

// parseIntParam parses an int parameter
func parseIntParam(s string, result *int) (bool, error) {
	var n int
	for _, c := range s {
		if c < '0' || c > '9' {
			return false, nil
		}
		n = n*10 + int(c-'0')
	}
	*result = n
	return true, nil
}
