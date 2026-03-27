package apperrors

import "github.com/gofiber/fiber/v2"

type ErrorCode string

const (
	ErrNotFound     ErrorCode = "NOT_FOUND"
	ErrConflict     ErrorCode = "CONFLICT"
	ErrUnauthorized ErrorCode = "UNAUTHORIZED"
)

type ErrorResponse struct {
	Message    string    `json:"message"`
	StatusCode int       `json:"status_code"`
	Code       ErrorCode `json:"code"`
}

func (e *ErrorResponse) Error() string {
	return e.Message
}

func NewCustomError(message string, statuscode int, errorCode ErrorCode) *ErrorResponse {
	return &ErrorResponse{
		Message:    message,
		StatusCode: statuscode,
		Code:       errorCode,
	}
}

func NotFoundError(message string) *ErrorResponse {
	return &ErrorResponse{
		Message:    message,
		StatusCode: fiber.StatusNotFound,
		Code:       ErrNotFound,
	}
}

func BadException(message string) *ErrorResponse {
	return &ErrorResponse{
		Message:    message,
		StatusCode: fiber.StatusBadRequest,
	}
}

func InternalServerError(message string) *ErrorResponse {
	return &ErrorResponse{
		Message:    message,
		StatusCode: fiber.StatusInternalServerError,
	}
}

func UnprocessableEntityError(message string) *ErrorResponse {
	return &ErrorResponse{
		Message:    message,
		StatusCode: fiber.StatusUnprocessableEntity,
	}
}

func ConflictError(message string) *ErrorResponse {
	return &ErrorResponse{
		Message:    message,
		StatusCode: fiber.StatusConflict,
	}
}

func UnauthorizedException(message string) *ErrorResponse {
	return &ErrorResponse{
		Message:    message,
		StatusCode: fiber.StatusUnauthorized,
	}
}
