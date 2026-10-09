package config

import (
	"errors"
	"log/slog"
	"os"

	"github.com/gofiber/fiber/v2"
)

type applicationError interface {
	error
	GetStatus() int
	GetCode() string
	GetMessage() string
	GetFields() map[string]string
}

func ErrorHandler(logger *slog.Logger) fiber.ErrorHandler {
	return func(c *fiber.Ctx, err error) error {
		status := fiber.StatusInternalServerError
		code := "INTERNAL_ERROR"
		message := "terjadi kesalahan internal"
		var fields map[string]string

		var appErr applicationError
		if errors.As(err, &appErr) {
			status = appErr.GetStatus()
			code = appErr.GetCode()
			message = appErr.GetMessage()
			fields = appErr.GetFields()
		} else {
			var fiberErr *fiber.Error
			if errors.As(err, &fiberErr) {
				status = fiberErr.Code
				message = fiberErr.Message

				switch status {
				case fiber.StatusBadRequest:
					code = "BAD_REQUEST"
				case fiber.StatusUnauthorized:
					code = "UNAUTHORIZED"
				case fiber.StatusForbidden:
					code = "FORBIDDEN"
				case fiber.StatusNotFound:
					code = "NOT_FOUND"
				case fiber.StatusConflict:
					code = "CONFLICT"
				case fiber.StatusUnsupportedMediaType:
					code = "UNSUPPORTED_MEDIA_TYPE"
				case fiber.StatusUnprocessableEntity:
					code = "VALIDATION_ERROR"
				case fiber.StatusNotAcceptable:
					code = "NOT_ACCEPTABLE"
				case fiber.StatusTooManyRequests:
					code = "TOO_MANY_REQUESTS"
				default:
					if status >= fiber.StatusInternalServerError {
						code = "INTERNAL_ERROR"
						message = "terjadi kesalahan internal"
					}
				}
			}
		}

		if status < fiber.StatusInternalServerError {
			logger.Warn("request_rejected",
				slog.Int("status", status),
				slog.String("code", code),
				slog.String("path", c.Path()),
			)
		} else {
			logger.Error("request_failed",
				slog.Int("status", status),
				slog.String("code", code),
				slog.String("path", c.Path()),
				slog.String("error", err.Error()),
			)
		}

		response := fiber.Map{
			"success":    false,
			"message":    message,
			"code":       code,
			"request_id": c.GetRespHeader("X-Request-Id"),
		}

		if len(fields) > 0 {
			response["errors"] = fields
		}

		if sendErr := c.Status(status).JSON(response); sendErr != nil {
			logger.Error("error_response_failed",
				slog.String("error", sendErr.Error()),
			)
			return c.Status(fiber.StatusInternalServerError).
				SendString("Internal Server Error")
		}

		return nil
	}
}

func NewLogger() *slog.Logger {
	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
}
