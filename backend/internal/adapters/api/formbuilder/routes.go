package formbuilder

import "github.com/gofiber/fiber/v2"

// RegisterRoutes registers all form builder routes under the given group.
// The group should already have auth + context loader middleware applied.
func RegisterRoutes(group fiber.Router, handler *FormHandler) {
	forms := group.Group("/formbuilder")

	// Form-level routes
	forms.Get("/forms", handler.ListForms)
	forms.Post("/forms", handler.CreateForm)
	forms.Get("/forms/:formKey", handler.ResolveForm)

	// Layout routes
	forms.Get("/forms/:formKey/layouts", handler.ListLayouts)
	forms.Post("/forms/:formKey/layouts", handler.CreateLayout)
	forms.Get("/forms/:formKey/layouts/:layoutName/draft", handler.GetDraft)

	// Assignment routes
	forms.Get("/forms/:formKey/assignments", handler.ListAssignments)
}
