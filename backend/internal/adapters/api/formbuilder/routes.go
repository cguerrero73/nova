package formbuilder

import "github.com/gofiber/fiber/v2"

// RegisterRoutes registers all form builder routes under the given group.
// The group should already have auth + context loader middleware applied.
func RegisterRoutes(group fiber.Router, handler *FormHandler) {
	forms := group.Group("/formbuilder")

	// Form-level routes
	forms.Get("/forms", handler.ListForms)
	forms.Post("/forms", handler.CreateForm)
	forms.Post("/forms/:formKey/archive", handler.ArchiveForm)
	forms.Get("/forms/:formKey", handler.ResolveForm)

	// Layout routes
	forms.Get("/forms/:formKey/layouts", handler.ListLayouts)
	forms.Post("/forms/:formKey/layouts", handler.CreateLayout)
	forms.Post("/forms/:formKey/layouts/:layoutName/archive", handler.ArchiveLayout)

	// Draft routes
	forms.Get("/forms/:formKey/layouts/:layoutName/draft", handler.GetDraft)
	forms.Put("/forms/:formKey/layouts/:layoutName/draft", handler.SaveDraft)

	// Publish & revert
	forms.Post("/forms/:formKey/layouts/:layoutName/publish", handler.Publish)
	forms.Post("/forms/:formKey/layouts/:layoutName/revert", handler.Revert)

	// Version history
	forms.Get("/forms/:formKey/layouts/:layoutName/versions", handler.ListVersions)
	forms.Get("/forms/:formKey/layouts/:layoutName/versions/:n", handler.GetVersion)

	// Assignment routes
	forms.Get("/forms/:formKey/assignments", handler.ListAssignments)
	forms.Put("/forms/:formKey/assignments/:roleName", handler.AssignRole)
	forms.Delete("/forms/:formKey/assignments/:roleName", handler.RevokeAssignment)
}
