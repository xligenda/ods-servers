package apierrors

import "github.com/gofiber/fiber/v2"

type APIError struct {
	StatusCode int    `json:"status"`
	Message    string `json:"message"`
}

func ErrorHandler(ctx *fiber.Ctx, err error) error {
	if apiErr, ok := err.(*APIError); ok {
		return ctx.Status(apiErr.StatusCode).JSON(apiErr)
	}
	return ctx.Status(ErrInternal.StatusCode).JSON(ErrInternal.Message)
}

func (e *APIError) Error() string {
	return e.Message
}

// With returns a new APIError with the same StatusCode but a custom message.
func (e *APIError) With(message string) *APIError {
	return &APIError{
		StatusCode: e.StatusCode,
		Message:    message,
	}
}

func New(statusCode int, message string) *APIError {
	return &APIError{StatusCode: statusCode, Message: message}
}

var (
	// Client Errors (4xx)
	ErrBadRequest       = &APIError{StatusCode: fiber.StatusBadRequest, Message: "Bad request"}
	ErrUnauthorized     = &APIError{StatusCode: fiber.StatusUnauthorized, Message: "Unauthorized"}
	ErrForbidden        = &APIError{StatusCode: fiber.StatusForbidden, Message: "Forbidden"}
	ErrNotFound         = &APIError{StatusCode: fiber.StatusNotFound, Message: "Resource not found"}
	ErrMethodNotAllowed = &APIError{StatusCode: fiber.StatusMethodNotAllowed, Message: "Method not allowed"}
	ErrRequestTimeout   = &APIError{StatusCode: fiber.StatusRequestTimeout, Message: "Request timeout"}
	ErrConflict         = &APIError{StatusCode: fiber.StatusConflict, Message: "Conflict detected"}
	ErrPayloadTooLarge  = &APIError{StatusCode: fiber.StatusRequestEntityTooLarge, Message: "Payload too large"}
	ErrUnsupportedMedia = &APIError{StatusCode: fiber.StatusUnsupportedMediaType, Message: "Unsupported media type"}
	ErrTooManyRequests  = &APIError{StatusCode: fiber.StatusTooManyRequests, Message: "Too many requests"}

	// Server Errors (5xx)
	ErrInternal           = &APIError{StatusCode: fiber.StatusInternalServerError, Message: "Internal server error"}
	ErrNotImplemented     = &APIError{StatusCode: fiber.StatusNotImplemented, Message: "Not implemented"}
	ErrBadGateway         = &APIError{StatusCode: fiber.StatusBadGateway, Message: "Bad gateway"}
	ErrServiceUnavailable = &APIError{StatusCode: fiber.StatusServiceUnavailable, Message: "Service unavailable"}
	ErrGatewayTimeout     = &APIError{StatusCode: fiber.StatusGatewayTimeout, Message: "Gateway timeout"}

	// Auth Errors
	ErrInvalidToken       = &APIError{StatusCode: fiber.StatusUnauthorized, Message: "Invalid token"}
	ErrExpiredToken       = &APIError{StatusCode: fiber.StatusUnauthorized, Message: "Token has expired"}
	ErrInsufficientRights = &APIError{StatusCode: fiber.StatusForbidden, Message: "Insufficient permissions"}

	// Validation Errors
	ErrValidationFailed     = &APIError{StatusCode: fiber.StatusUnprocessableEntity, Message: "Validation failed"}
	ErrMissingRequiredField = &APIError{StatusCode: fiber.StatusBadRequest, Message: "Missing required field"}

	// DB Errors
	ErrDatabaseError  = &APIError{StatusCode: fiber.StatusInternalServerError, Message: "Database error"}
	ErrDuplicateEntry = &APIError{StatusCode: fiber.StatusConflict, Message: "Duplicate entry"}
	ErrRecordNotFound = &APIError{StatusCode: fiber.StatusNotFound, Message: "Record not found"}

	// FS Errors
	ErrFileTooLarge         = &APIError{StatusCode: fiber.StatusRequestEntityTooLarge, Message: "File size too large"}
	ErrFileFormatNotAllowed = &APIError{StatusCode: fiber.StatusUnsupportedMediaType, Message: "File format not allowed"}
)
