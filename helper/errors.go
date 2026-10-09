package helper

import (
	"errors"
	"fmt"

	"github.com/gofiber/fiber/v2"
)

const (
	CodeValidation       = "VALIDATION_ERROR"
	CodeBadRequest       = "BAD_REQUEST"
	CodeUnauthorized     = "UNAUTHORIZED"
	CodeForbidden        = "FORBIDDEN"
	CodeNotFound         = "NOT_FOUND"
	CodeConflict         = "CONFLICT"
	CodeUnsupportedMedia = "UNSUPPORTED_MEDIA_TYPE"
	CodeNotAcceptable    = "NOT_ACCEPTABLE"
	CodeTooManyRequests  = "TOO_MANY_REQUESTS"
	CodeInternal         = "INTERNAL_ERROR"
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
		return e.cause.Error()
	}
	return e.Message
}

func (e *AppError) Unwrap() error {
	return e.cause
}

func BadRequest(message string) *AppError {
	return &AppError{
		Status:  fiber.StatusBadRequest,
		Code:    CodeBadRequest,
		Message: message,
	}
}

func Validation(fields map[string]string) *AppError {
	return &AppError{
		Status:  fiber.StatusUnprocessableEntity,
		Code:    CodeValidation,
		Message: "validasi gagal",
		Fields:  fields,
	}
}

func Unauthorized(message string) *AppError {
	return &AppError{
		Status:  fiber.StatusUnauthorized,
		Code:    CodeUnauthorized,
		Message: message,
	}
}

func Forbidden(message string) *AppError {
	return &AppError{
		Status:  fiber.StatusForbidden,
		Code:    CodeForbidden,
		Message: message,
	}
}

func NotFound(message string) *AppError {
	return &AppError{
		Status:  fiber.StatusNotFound,
		Code:    CodeNotFound,
		Message: message,
	}
}

func Conflict(message string) *AppError {
	return &AppError{
		Status:  fiber.StatusConflict,
		Code:    CodeConflict,
		Message: message,
	}
}

func UnsupportedMedia() *AppError {
	return &AppError{
		Status:  fiber.StatusUnsupportedMediaType,
		Code:    CodeUnsupportedMedia,
		Message: "tipe media request tidak didukung",
	}
}

func NotAcceptable() *AppError {
	return &AppError{
		Status:  fiber.StatusNotAcceptable,
		Code:    CodeNotAcceptable,
		Message: "format response yang diminta tidak tersedia",
	}
}

func Internal(err error) *AppError {
	if err == nil {
		err = errors.New("unknown internal error")
	}

	return &AppError{
		Status:  fiber.StatusInternalServerError,
		Code:    CodeInternal,
		Message: "terjadi kesalahan internal",
		cause:   fmt.Errorf("internal error: %w", err),
	}
}

func (e *AppError) GetStatus() int {
	return e.Status
}

func (e *AppError) GetCode() string {
	return e.Code
}

func (e *AppError) GetMessage() string {
	return e.Message
}

func (e *AppError) GetFields() map[string]string {
	return e.Fields
}
