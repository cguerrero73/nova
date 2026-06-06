package api

import (
	"github.com/gofiber/fiber/v2"

	"github.com/nova/backend/internal/domain/queries"
	"github.com/nova/backend/pkg/errors"
)

type QueriesHandler struct {
	queryService *queries.Service
}

func NewQueriesHandler(queryService *queries.Service) *QueriesHandler {
	return &QueriesHandler{queryService: queryService}
}

// List returns saved queries for a grid
// GET /queries?gridId=1
func (h *QueriesHandler) List(c *fiber.Ctx) error {
	gridID := c.QueryInt("gridId", 0)
	if gridID == 0 {
		return c.Status(400).JSON(errors.New("BAD_REQUEST", "gridId is required", 400))
	}

	// Get user from context (set by auth middleware)
	userCode := getUserCode(c)
	if userCode == "" {
		userCode = "anonymous"
	}

	result, err := h.queryService.List(c.Context(), gridID, userCode)
	if err != nil {
		return c.Status(500).JSON(errors.ErrInternal)
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    result,
	})
}

// Get returns a specific saved query
// GET /queries/:id
func (h *QueriesHandler) Get(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(400).JSON(errors.New("BAD_REQUEST", "id is required", 400))
	}

	query, err := h.queryService.GetByID(c.Context(), id)
	if err != nil {
		return c.Status(404).JSON(errors.ErrNotFound)
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    query,
	})
}

// Create creates a new saved query
// POST /queries
func (h *QueriesHandler) Create(c *fiber.Ctx) error {
	var req queries.SaveRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(errors.ErrBadRequest)
	}

	if req.Name == "" {
		return c.Status(400).JSON(errors.New("BAD_REQUEST", "name is required", 400))
	}

	// Get user from context
	userCode := getUserCode(c)
	if userCode != "" {
		req.UserID = &userCode
	}

	query, err := h.queryService.Create(c.Context(), &req)
	if err != nil {
		return c.Status(500).JSON(errors.ErrInternal)
	}

	return c.Status(201).JSON(fiber.Map{
		"success": true,
		"data":    query,
	})
}

// Update updates a saved query
// PUT /queries/:id
func (h *QueriesHandler) Update(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(400).JSON(errors.New("BAD_REQUEST", "id is required", 400))
	}

	var req queries.UpdateRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(errors.ErrBadRequest)
	}

	query, err := h.queryService.Update(c.Context(), id, &req)
	if err != nil {
		return c.Status(500).JSON(errors.ErrInternal)
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    query,
	})
}

// Delete removes a saved query
// DELETE /queries/:id
func (h *QueriesHandler) Delete(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(400).JSON(errors.New("BAD_REQUEST", "id is required", 400))
	}

	if err := h.queryService.Delete(c.Context(), id); err != nil {
		return c.Status(500).JSON(errors.ErrInternal)
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    nil,
	})
}

// Helper to get user code from context
func getUserCode(c *fiber.Ctx) string {
	// Check for user claims set by auth middleware
	if user := c.Locals("user"); user != nil {
		if claims, ok := user.(map[string]interface{}); ok {
			if code, ok := claims["userCode"]; ok {
				return code.(string)
			}
		}
	}
	return ""
}
