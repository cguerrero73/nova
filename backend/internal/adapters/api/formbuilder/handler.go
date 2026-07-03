package formbuilder

import (
	"github.com/gofiber/fiber/v2"

	"github.com/nova/backend/internal/domain/formbuilder"
	"github.com/nova/backend/internal/handler/dto"
	"github.com/nova/backend/internal/infrastructure/middleware"
	"github.com/nova/backend/pkg/errors"
)

// FormHandler handles HTTP requests for the form builder module.
type FormHandler struct {
	service *formbuilder.FormService
}

// NewFormHandler creates a new FormHandler.
func NewFormHandler(service *formbuilder.FormService) *FormHandler {
	return &FormHandler{service: service}
}

// ListForms handles GET /api/formbuilder/forms
func (h *FormHandler) ListForms(c *fiber.Ctx) error {
	if !requirePermission(c, "view") {
		return nil
	}

	status := c.Query("status")
	forms, err := h.service.ListForms(c.UserContext(), status)
	if err != nil {
		return c.Status(500).JSON(dto.NewErrorResponse("INTERNAL", err.Error()))
	}

	return c.JSON(dto.NewSuccessResponse(forms))
}

// CreateForm handles POST /api/formbuilder/forms
func (h *FormHandler) CreateForm(c *fiber.Ctx) error {
	if !requirePermission(c, "design") {
		return nil
	}

	var req formbuilder.CreateFormRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(errors.ErrBadRequest)
	}

	if req.Key == "" || req.Name == "" {
		return c.Status(400).JSON(dto.NewErrorResponse("VALIDATION_ERROR", "key and name are required"))
	}

	actor := getActor(c)
	form, layout, err := h.service.CreateForm(c.UserContext(), &req, actor)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return c.Status(appErr.Status).JSON(appErr)
		}
		return c.Status(500).JSON(dto.NewErrorResponse("INTERNAL", err.Error()))
	}

	_ = layout // default layout created as side-effect
	return c.Status(201).JSON(dto.NewSuccessResponse(form))
}

// ResolveForm handles GET /api/formbuilder/forms/:formKey (runtime resolution)
func (h *FormHandler) ResolveForm(c *fiber.Ctx) error {
	if !requirePermission(c, "view") {
		return nil
	}

	formKey := c.Params("formKey")
	roleName := middleware.GetActiveRole(c)

	result, err := h.service.Resolve(c.UserContext(), formKey, roleName)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return c.Status(appErr.Status).JSON(appErr)
		}
		return c.Status(500).JSON(dto.NewErrorResponse("INTERNAL", err.Error()))
	}

	return c.JSON(dto.NewSuccessResponse(result))
}

// ListLayouts handles GET /api/formbuilder/forms/:formKey/layouts
func (h *FormHandler) ListLayouts(c *fiber.Ctx) error {
	if !requirePermission(c, "view") {
		return nil
	}

	formKey := c.Params("formKey")
	layouts, err := h.service.ListLayouts(c.UserContext(), formKey)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return c.Status(appErr.Status).JSON(appErr)
		}
		return c.Status(500).JSON(dto.NewErrorResponse("INTERNAL", err.Error()))
	}

	return c.JSON(dto.NewSuccessResponse(layouts))
}

// CreateLayout handles POST /api/formbuilder/forms/:formKey/layouts
func (h *FormHandler) CreateLayout(c *fiber.Ctx) error {
	if !requirePermission(c, "design") {
		return nil
	}

	formKey := c.Params("formKey")
	var req formbuilder.CreateLayoutRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(errors.ErrBadRequest)
	}

	if req.Name == "" || req.DisplayName == "" {
		return c.Status(400).JSON(dto.NewErrorResponse("VALIDATION_ERROR", "name and displayName are required"))
	}

	actor := getActor(c)
	layout, err := h.service.CreateLayout(c.UserContext(), formKey, &req, actor)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return c.Status(appErr.Status).JSON(appErr)
		}
		return c.Status(500).JSON(dto.NewErrorResponse("INTERNAL", err.Error()))
	}

	return c.Status(201).JSON(dto.NewSuccessResponse(layout))
}

// GetDraft handles GET /api/formbuilder/forms/:formKey/layouts/:layoutName/draft
func (h *FormHandler) GetDraft(c *fiber.Ctx) error {
	if !requirePermission(c, "view_draft") {
		return nil
	}

	formKey := c.Params("formKey")
	layoutName := c.Params("layoutName")

	version, err := h.service.GetDraft(c.UserContext(), formKey, layoutName)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return c.Status(appErr.Status).JSON(appErr)
		}
		return c.Status(500).JSON(dto.NewErrorResponse("INTERNAL", err.Error()))
	}

	return c.JSON(dto.NewSuccessResponse(version))
}

// ListAssignments handles GET /api/formbuilder/forms/:formKey/assignments
func (h *FormHandler) ListAssignments(c *fiber.Ctx) error {
	if !requirePermission(c, "view") {
		return nil
	}

	formKey := c.Params("formKey")
	assignments, err := h.service.ListAssignments(c.UserContext(), formKey)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return c.Status(appErr.Status).JSON(appErr)
		}
		return c.Status(500).JSON(dto.NewErrorResponse("INTERNAL", err.Error()))
	}

	return c.JSON(dto.NewSuccessResponse(assignments))
}

// SaveDraft handles PUT /api/formbuilder/forms/:formKey/layouts/:layoutName/draft
func (h *FormHandler) SaveDraft(c *fiber.Ctx) error {
	if !requirePermission(c, "design") {
		return nil
	}

	formKey := c.Params("formKey")
	layoutName := c.Params("layoutName")

	// The body is raw layout JSON
	definition := c.Body()
	if len(definition) == 0 {
		return c.Status(400).JSON(dto.NewErrorResponse("VALIDATION_ERROR", "Request body must contain layout JSON"))
	}

	actor := getActor(c)
	version, err := h.service.SaveDraft(c.UserContext(), formKey, layoutName, definition, actor)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return c.Status(appErr.Status).JSON(appErr)
		}
		return c.Status(500).JSON(dto.NewErrorResponse("INTERNAL", err.Error()))
	}

	return c.JSON(dto.NewSuccessResponse(version))
}

// Publish handles POST /api/formbuilder/forms/:formKey/layouts/:layoutName/publish
func (h *FormHandler) Publish(c *fiber.Ctx) error {
	if !requirePermission(c, "publish") {
		return nil
	}

	formKey := c.Params("formKey")
	layoutName := c.Params("layoutName")

	var req formbuilder.PublishRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(errors.ErrBadRequest)
	}

	actor := getActor(c)
	version, err := h.service.Publish(c.UserContext(), formKey, layoutName, req.Description, actor)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return c.Status(appErr.Status).JSON(appErr)
		}
		return c.Status(500).JSON(dto.NewErrorResponse("INTERNAL", err.Error()))
	}

	return c.Status(201).JSON(dto.NewSuccessResponse(version))
}

// Revert handles POST /api/formbuilder/forms/:formKey/layouts/:layoutName/revert
func (h *FormHandler) Revert(c *fiber.Ctx) error {
	if !requirePermission(c, "design") {
		return nil
	}

	formKey := c.Params("formKey")
	layoutName := c.Params("layoutName")

	var req formbuilder.RevertRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(errors.ErrBadRequest)
	}

	if req.VersionNumber <= 0 {
		return c.Status(400).JSON(dto.NewErrorResponse("VALIDATION_ERROR", "versionNumber is required and must be positive"))
	}

	actor := getActor(c)
	version, err := h.service.Revert(c.UserContext(), formKey, layoutName, req.VersionNumber, actor)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return c.Status(appErr.Status).JSON(appErr)
		}
		return c.Status(500).JSON(dto.NewErrorResponse("INTERNAL", err.Error()))
	}

	return c.JSON(dto.NewSuccessResponse(version))
}

// ArchiveLayout handles POST /api/formbuilder/forms/:formKey/layouts/:layoutName/archive
func (h *FormHandler) ArchiveLayout(c *fiber.Ctx) error {
	if !requirePermission(c, "publish") {
		return nil
	}

	formKey := c.Params("formKey")
	layoutName := c.Params("layoutName")

	actor := getActor(c)
	if err := h.service.ArchiveLayout(c.UserContext(), formKey, layoutName, actor); err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return c.Status(appErr.Status).JSON(appErr)
		}
		return c.Status(500).JSON(dto.NewErrorResponse("INTERNAL", err.Error()))
	}

	return c.JSON(dto.NewMessageResponse("Layout archived"))
}

// ArchiveForm handles POST /api/formbuilder/forms/:formKey/archive
func (h *FormHandler) ArchiveForm(c *fiber.Ctx) error {
	if !requirePermission(c, "publish") {
		return nil
	}

	formKey := c.Params("formKey")
	actor := getActor(c)

	if err := h.service.ArchiveForm(c.UserContext(), formKey, actor); err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return c.Status(appErr.Status).JSON(appErr)
		}
		return c.Status(500).JSON(dto.NewErrorResponse("INTERNAL", err.Error()))
	}

	return c.JSON(dto.NewMessageResponse("Form archived"))
}

// ListVersions handles GET /api/formbuilder/forms/:formKey/layouts/:layoutName/versions
func (h *FormHandler) ListVersions(c *fiber.Ctx) error {
	if !requirePermission(c, "view") {
		return nil
	}

	formKey := c.Params("formKey")
	layoutName := c.Params("layoutName")

	versions, err := h.service.ListVersions(c.UserContext(), formKey, layoutName)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return c.Status(appErr.Status).JSON(appErr)
		}
		return c.Status(500).JSON(dto.NewErrorResponse("INTERNAL", err.Error()))
	}

	return c.JSON(dto.NewSuccessResponse(versions))
}

// GetVersion handles GET /api/formbuilder/forms/:formKey/layouts/:layoutName/versions/:n
func (h *FormHandler) GetVersion(c *fiber.Ctx) error {
	if !requirePermission(c, "view") {
		return nil
	}

	formKey := c.Params("formKey")
	layoutName := c.Params("layoutName")
	versionNumber, err := c.ParamsInt("n")
	if err != nil || versionNumber <= 0 {
		return c.Status(400).JSON(dto.NewErrorResponse("VALIDATION_ERROR", "Version number must be a positive integer"))
	}

	version, err := h.service.GetVersion(c.UserContext(), formKey, layoutName, versionNumber)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return c.Status(appErr.Status).JSON(appErr)
		}
		return c.Status(500).JSON(dto.NewErrorResponse("INTERNAL", err.Error()))
	}

	return c.JSON(dto.NewSuccessResponse(version))
}

// AssignRole handles PUT /api/formbuilder/forms/:formKey/assignments/:roleName
func (h *FormHandler) AssignRole(c *fiber.Ctx) error {
	if !requirePermission(c, "assign") {
		return nil
	}

	formKey := c.Params("formKey")
	roleName := c.Params("roleName")

	var req formbuilder.AssignRoleRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(errors.ErrBadRequest)
	}

	if req.LayoutName == "" {
		return c.Status(400).JSON(dto.NewErrorResponse("VALIDATION_ERROR", "layoutName is required"))
	}

	actor := getActor(c)
	assignment, err := h.service.AssignRole(c.UserContext(), formKey, roleName, req.LayoutName, actor)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return c.Status(appErr.Status).JSON(appErr)
		}
		return c.Status(500).JSON(dto.NewErrorResponse("INTERNAL", err.Error()))
	}

	return c.Status(201).JSON(dto.NewSuccessResponse(assignment))
}

// RevokeAssignment handles DELETE /api/formbuilder/forms/:formKey/assignments/:roleName
func (h *FormHandler) RevokeAssignment(c *fiber.Ctx) error {
	if !requirePermission(c, "assign") {
		return nil
	}

	formKey := c.Params("formKey")
	roleName := c.Params("roleName")

	actor := getActor(c)
	if err := h.service.RevokeAssignment(c.UserContext(), formKey, roleName, actor); err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return c.Status(appErr.Status).JSON(appErr)
		}
		return c.Status(500).JSON(dto.NewErrorResponse("INTERNAL", err.Error()))
	}

	return c.JSON(dto.NewMessageResponse("Assignment revoked"))
}

// requirePermission checks if the active role has the given formbuilder permission.
// Returns true if allowed, false if a 403 response was already written.
func requirePermission(c *fiber.Ctx, action string) bool {
	role := middleware.GetRole(c)
	if role == nil {
		_ = c.Status(403).JSON(dto.NewErrorResponse("FORBIDDEN", "No active role in session"))
		return false
	}
	if !role.HasPermission("formbuilder", action) {
		_ = c.Status(403).JSON(dto.NewErrorResponse("FORBIDDEN", "Missing permission: formbuilder."+action))
		return false
	}
	return true
}

// getActor extracts the user code from the request context for audit attribution.
func getActor(c *fiber.Ctx) string {
	claims := middleware.GetUserClaims(c)
	if claims != nil {
		return claims.UserCode
	}
	return "unknown"
}
