package middleware

import (
	"errors"
	"log/slog"
	"time"

	"github.com/GanisSayogyo/tugas1-go-fiber/helper"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func RequestID() fiber.Handler {
	return func(c *fiber.Ctx) error {
		requestID := c.Get("X-Request-ID")
		if requestID == "" {
			requestID = uuid.NewString()
		}

		c.Locals("request_id", requestID)
		c.Set("X-Request-ID", requestID)

		return c.Next()
	}
}

func RequestLogger(logger *slog.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		err := c.Next()

		status := c.Response().StatusCode()

		if err != nil {
			var appErr *helper.AppError
			var fiberErr *fiber.Error

			switch {
			case errors.As(err, &appErr):
				status = appErr.GetStatus()
			case errors.As(err, &fiberErr):
				status = fiberErr.Code
			default:
				status = fiber.StatusInternalServerError
			}
		}

		logger.Info("http_request",
			slog.String("method", c.Method()),
			slog.String("path", c.Path()),
			slog.Int("status", status),
			slog.Duration("duration", time.Since(start)),
			slog.String("request_id", c.Get("X-Request-ID")),
		)

		return err
	}
}

func fiberErrStatus(err error, target **fiber.Error) bool {
	if e, ok := err.(*fiber.Error); ok {
		*target = e
		return true
	}
	return false
}
