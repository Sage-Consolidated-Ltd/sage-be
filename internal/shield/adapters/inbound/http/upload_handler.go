package http

import (
	"strings"

	"sage-backend/internal/shared/logger"
	"sage-backend/internal/shared/middlewares"
	"sage-backend/internal/shared/response"
	"sage-backend/internal/shared/utils"
	"sage-backend/internal/shield/ports/dto"
	"sage-backend/internal/shield/ports/inbound"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

type UploadHandler struct {
	uploadService inbound.UploadUseCase
}

func NewUploadHandler(uploadService inbound.UploadUseCase) *UploadHandler {
	return &UploadHandler{
		uploadService: uploadService,
	}
}

func (u *UploadHandler) UploadLog(c *fiber.Ctx) error {
	orgID, _, _, ok := getSessionInfo(c)
	if !ok {
		return response.Error(c, fiber.StatusUnauthorized, "UNAUTHORIZED", nil)
	}
	rc := middlewares.GetRequestContext(c)

	var req dto.UploadLogRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "BAD_REQUEST", nil)
	}

	if err := utils.Validate.Struct(req); err != nil {
		errs := utils.ValidationErrors(err)
		return response.Error(c, fiber.StatusUnprocessableEntity, err.Error(), errs)
	}

	resp, err := u.uploadService.GetUploadURL(
		c.Context(),
		rc,
		orgID,
		req.Filename,
		req.ContentType,
		req.Size,
	)
	if err != nil {
		logger.Error("Error creating presignURL at UploadHandler.UploadLog", zap.Error(err))
		return response.Error(c, fiber.StatusInternalServerError, "FAILED_TO_CREATE_PRESIGN_URL", err.Error())
	}

	return response.JSON(c, fiber.StatusOK, "PresignURL created", resp)
}

func (u *UploadHandler) UploadComplete(c *fiber.Ctx) error {
	var req dto.UploadCompleteRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "BAD_REQUEST", nil)
	}
	if err := utils.Validate.Struct(req); err != nil {
		errs := utils.ValidationErrors(err)
		return response.Error(c, fiber.StatusUnprocessableEntity, err.Error(), errs)
	}

	rc := middlewares.GetRequestContext(c)
	if !strings.Contains(req.Key, rc.UserID) {
		return response.Error(c, fiber.StatusForbidden, "you do not own this file", nil)
	}

	confirmed, err := u.uploadService.ValidateUploadComplete(
		c.Context(),
		&rc,
		&req,
	)
	if err != nil {
		return response.Error(c, fiber.StatusUnprocessableEntity, "UPLOAD_VALIDATION_FAILED", err.Error())
	}

	return response.JSON(c, fiber.StatusOK, "UPLOAD_COMPLETE", confirmed)
}
