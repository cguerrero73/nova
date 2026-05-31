package api

import (
	"github.com/gofiber/fiber/v2"
)

// QueriesHandler handles saved query endpoints (stub)
type QueriesHandler struct{}

func NewQueriesHandler() *QueriesHandler {
	return &QueriesHandler{}
}

// List returns saved queries for a grid (empty until feature is implemented)
func (h *QueriesHandler) List(c *fiber.Ctx) error {
	// Return empty list - no saved queries yet
	return c.JSON(fiber.Map{
		"success": true,
		"data":    []any{},
	})
}
