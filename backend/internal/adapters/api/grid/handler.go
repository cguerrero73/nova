package api

import (
	"github.com/gofiber/fiber/v2"
)

// GridHandler handles grid data queries (stub until feature is implemented)
type GridHandler struct{}

func NewGridHandler() *GridHandler {
	return &GridHandler{}
}

// ExecuteQuery processes a grid data request and returns empty data
func (h *GridHandler) ExecuteQuery(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"success": true,
		"data":    []any{},
		"meta": fiber.Map{
			"page":       1,
			"pageSize":   20,
			"total":      0,
			"totalPages": 0,
		},
	})
}
