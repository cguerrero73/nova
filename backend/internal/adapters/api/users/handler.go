package api

import (
	"github.com/gofiber/fiber/v2"

	"github.com/nova/backend/internal/domain/users"
	"github.com/nova/backend/internal/handler/dto"
	"github.com/nova/backend/pkg/errors"
)

type UserHandler struct {
	userService *users.UserService
}

func NewUserHandler(service *users.UserService) *UserHandler {
	return &UserHandler{userService: service}
}

func (h *UserHandler) List(c *fiber.Ctx) error {
	tenant := c.Locals("tenant").(string)
	if tenant == "" {
		return c.Status(400).JSON(errors.ErrTenantRequired)
	}

	pagination := dto.PaginationQuery{
		Page:     c.QueryInt("page", 1),
		PageSize: c.QueryInt("page_size", 20),
	}

	users, total, err := h.userService.FindAll(c.Context(), tenant, pagination.GetLimit(), pagination.GetOffset())
	if err != nil {
		return c.Status(500).JSON(dto.NewErrorResponse("INTERNAL", err.Error()))
	}

	return c.JSON(dto.NewListResponse(users, dto.NewPaginationMeta(pagination.Page, pagination.PageSize, total)))
}

func (h *UserHandler) Get(c *fiber.Ctx) error {
	id := c.Params("id")
	user, err := h.userService.FindByID(c.Context(), id)
	if err != nil {
		return c.Status(500).JSON(dto.NewErrorResponse("INTERNAL", err.Error()))
	}
	if user == nil {
		return c.Status(404).JSON(dto.NewErrorResponse("NOT_FOUND", "User not found"))
	}

	return c.JSON(dto.NewSuccessResponse(user))
}

func (h *UserHandler) Create(c *fiber.Ctx) error {
	tenant := c.Locals("tenant").(string)
	if tenant == "" {
		return c.Status(400).JSON(errors.ErrTenantRequired)
	}

	var req users.CreateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(errors.ErrBadRequest)
	}

	user, err := h.userService.Create(c.Context(), tenant, &req)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return c.Status(appErr.Status).JSON(appErr)
		}
		return c.Status(500).JSON(dto.NewErrorResponse("INTERNAL", err.Error()))
	}

	return c.Status(201).JSON(dto.NewSuccessResponse(user))
}

func (h *UserHandler) Update(c *fiber.Ctx) error {
	id := c.Params("id")

	var req users.UpdateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(errors.ErrBadRequest)
	}

	user, err := h.userService.Update(c.Context(), id, &req)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return c.Status(appErr.Status).JSON(appErr)
		}
		return c.Status(500).JSON(dto.NewErrorResponse("INTERNAL", err.Error()))
	}

	return c.JSON(dto.NewSuccessResponse(user))
}

func (h *UserHandler) Delete(c *fiber.Ctx) error {
	id := c.Params("id")

	if err := h.userService.Delete(c.Context(), id); err != nil {
		return c.Status(500).JSON(dto.NewErrorResponse("INTERNAL", err.Error()))
	}

	return c.JSON(dto.NewMessageResponse("User deleted successfully"))
}
