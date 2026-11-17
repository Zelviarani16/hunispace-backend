package helper

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
)

const (
	CodeValidation         = "VALIDATION_ERROR"
	CodeBadRequest         = "BAD_REQUEST"
	CodeUnauthorized       = "UNAUTHORIZED"
	CodeForbidden          = "FORBIDDEN"
	CodeNotFound           = "NOT_FOUND"
	CodeConflict           = "CONFLICT"
	CodeServiceUnavailable = "SERVICE_UNAVAILABLE"
	CodeInternal           = "INTERNAL_ERROR"
)

type AppError struct {
	Status  int
	Code    string
	Message string
	Fields  map[string]string
	cause   error
}

func (e *AppError) Error() string {
	if e.cause != nil {
		return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.cause)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *AppError) Unwrap() error { return e.cause }

func NewError(status int, code, message string) *AppError {
	return &AppError{Status: status, Code: code, Message: message}
}

func BadRequest(message string) *AppError {
	return NewError(fiber.StatusBadRequest, CodeBadRequest, message)
}

func Unauthorized(message string) *AppError {
	return NewError(fiber.StatusUnauthorized, CodeUnauthorized, message)
}

func Forbidden(message string) *AppError {
	return NewError(fiber.StatusForbidden, CodeForbidden, message)
}

func NotFound(message string) *AppError {
	return NewError(fiber.StatusNotFound, CodeNotFound, message)
}

func Conflict(code, message string) *AppError {
	return NewError(fiber.StatusConflict, code, message)
}

func Validation(fields map[string]string) *AppError {
	e := NewError(fiber.StatusUnprocessableEntity, CodeValidation, "data yang dikirim tidak valid")
	e.Fields = fields
	return e
}

func Internal(cause error) *AppError {
	e := NewError(fiber.StatusInternalServerError, CodeInternal, "terjadi kesalahan pada server")
	e.cause = cause
	return e
}
