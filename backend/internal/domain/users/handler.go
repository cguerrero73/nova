package users

import (
	"github.com/gofiber/fiber/v2"

	"github.com/nova/backend/internal/handler/dto"
	"github.com/nova/backend/pkg/errors"
)

type UserHandler struct {
	service *UserService
}

func NewUserHandler(service *UserService) *UserHandler {
	return &UserHandler{service: service}
}

type CreateUserRequest struct {
	Code       string `json:"usr_code"`
	Name       string `json:"usr_name" validate:"required"`
	Email      string `json:"usr_email" validate:"required,email"`
	Password   string `json:"usr_password" validate:"required,min=8"`
	Phone      string `json:"usr_phone"`
	DefaultOrg string `json:"usr_default_org"`
}

type UpdateUserRequest struct {
	Name       string `json:"usr_name"`
	Email      string `json:"usr_email" validate:"email"`
	Phone      string `json:"usr_phone"`
	Status     string `json:"usr_status"`
	DefaultOrg string `json:"usr_default_org"`
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

	users, total, err := h.service.FindAll(c.Context(), tenant, pagination.GetLimit(), pagination.GetOffset())
	if err != nil {
		return c.Status(500).JSON(dto.NewErrorResponse("INTERNAL", err.Error()))
	}

	return c.JSON(dto.NewListResponse(users, dto.NewPaginationMeta(pagination.Page, pagination.PageSize, total)))
}

func (h *UserHandler) Get(c *fiber.Ctx) error {
	id := c.Params("id")
	user, err := h.service.FindByID(c.Context(), id)
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

	var req CreateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(errors.ErrBadRequest)
	}

	user, err := h.service.Create(c.Context(), tenant, &req)
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

	var req UpdateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(errors.ErrBadRequest)
	}

	user, err := h.service.Update(c.Context(), id, &req)
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

	if err := h.service.Delete(c.Context(), id); err != nil {
		return c.Status(500).JSON(dto.NewErrorResponse("INTERNAL", err.Error()))
	}

	return c.JSON(dto.NewMessageResponse("User deleted successfully"))
}
