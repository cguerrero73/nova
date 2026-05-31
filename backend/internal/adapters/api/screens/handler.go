package api

import (
	"github.com/gofiber/fiber/v2"
)

type ScreenHandler struct{}

func NewScreenHandler() *ScreenHandler {
	return &ScreenHandler{}
}

func (h *ScreenHandler) GetTranslations(c *fiber.Ctx) error {
	screenID := c.Params("screenId")
	lang := c.Query("lang", "en")

	// Return empty translations (stub - backend endpoint not implemented yet)
	return c.JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"screenId":     screenID,
			"translations": fiber.Map{},
			"language":     lang,
		},
	})
}
