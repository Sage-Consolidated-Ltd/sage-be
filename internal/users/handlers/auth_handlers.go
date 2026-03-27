package handlers

import (
	"sage-backend/internal/shared/errors/apperrors"
	"sage-backend/internal/shared/response"
	"sage-backend/internal/shared/utils"
	"sage-backend/internal/users/requests"
	"sage-backend/internal/users/services"

	"github.com/gofiber/fiber/v2"
)

type AuthHandler struct {
	authServ services.AuthServiceInt
}

func NewAuthHandler(authServ services.AuthServiceInt) *AuthHandler {
	return &AuthHandler{
		authServ: authServ,
	}
}

func (h *AuthHandler) CreateUser(c *fiber.Ctx) error {
	var req requests.CreateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, err.Error(), nil)
	}
	if err := utils.Validate.Struct(req); err != nil {
		errs := utils.ValidationErrors(err)
		return response.Error(c, fiber.StatusUnprocessableEntity, err.Error(), errs)
	}

	// create user
	if err := h.authServ.CreateUser(c.Context(), &req); err != nil {
		if err, ok := err.(*apperrors.ErrorResponse); ok {
			return response.Error(c, err.StatusCode, err.Error(), nil)
		}
		return response.Error(c, fiber.StatusInternalServerError, err.Error(), nil)
	}

	return response.JSON(c, fiber.StatusOK, "User created successfully. A mail has been forwarded to verify account", nil)
}
