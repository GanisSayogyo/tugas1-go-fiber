package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"

	"github.com/GanisSayogyo/tugas1-go-fiber/helper"
)

func RequireAuth(c *fiber.Ctx) error {
	authHeader := c.Get("Authorization")

	if authHeader == "" {
		c.Set("WWW-Authenticate", `Bearer realm="api"`)

		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "authorization token required",
		})
	}

	parts := strings.SplitN(authHeader, " ", 2)

	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		c.Set("WWW-Authenticate", `Bearer realm="api"`)

		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "invalid authorization header",
		})
	}

	token := strings.TrimSpace(parts[1])

	if token == "" {
		c.Set("WWW-Authenticate", `Bearer realm="api"`)

		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "access token required",
		})
	}

	authUser, err := helper.ParseAccessToken(token)
	if err != nil {
		c.Set("WWW-Authenticate", `Bearer realm="api"`)

		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "invalid or expired access token",
		})
	}

	c.Locals("user_id", authUser.UserID)
	c.Locals("username", authUser.Username)
	c.Locals("role", authUser.Role)
	c.Locals("auth_user", authUser)

	return c.Next()
}
