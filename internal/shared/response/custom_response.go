package response

import "github.com/gofiber/fiber/v2"

type Response struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

type ErrorResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Error   interface{} `json:"error,omitempty"`
}

func JSON(c *fiber.Ctx, statusCode int, message string, data interface{}) error {
	return c.Status(statusCode).JSON(Response{
		Success: checkSuccess(statusCode),
		Message: message,
		Data:    data,
	})
}

func Error(c *fiber.Ctx, statusCode int, message string, err interface{}) error {
	return c.Status(statusCode).JSON(ErrorResponse{
		Success: checkSuccess(statusCode),
		Message: message,
		Error:   err,
	})
}

func checkSuccess(statusCode int) bool {
	if statusCode >= 200 && statusCode <= 300 {
		return true
	}
	return false
}
