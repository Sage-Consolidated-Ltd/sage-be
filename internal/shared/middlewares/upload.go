package middlewares

import (
	"fmt"
	"sage-backend/internal/shared/response"

	"github.com/gofiber/fiber/v2"
)

const (
	MaxAvatarSize = 5 * 1024 * 1024
	MaxBannerSize = 10 * 1024 * 1024
)

var AvatarMimeTypes = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/webp": true,
}

func ValidateFileUpload(fieldName string, maxSize int64, allowedTypes map[string]bool) fiber.Handler {
	return func(c *fiber.Ctx) error {
		fileHeader, err := c.FormFile(fieldName)
		if err != nil {
			return response.Error(c, fiber.StatusBadRequest, fmt.Sprintf("%s file required", fieldName), nil)
		}

		// Size check
		if fileHeader.Size > maxSize {
			return response.Error(c, fiber.StatusBadRequest,
				fmt.Sprintf("file too large (max %dMB)", maxSize/1024/1024), nil)
		}

		mimeType := fileHeader.Header.Get("Content-Type")
		if !allowedTypes[mimeType] {
			return response.Error(c, fiber.StatusBadRequest, "invalid file type", nil)
		}

		c.Locals("file_mime_type", mimeType)
		return c.Next()
	}
}
