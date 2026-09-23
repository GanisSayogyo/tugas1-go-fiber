package middleware

import (
	"github.com/gofiber/fiber/v2"

	"github.com/GanisSayogyo/tugas1-go-fiber/helper"
)

func RequirePermission(perms *helper.PermissionSet, permission string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		user, ok := helper.CurrentUser(c)

		if !ok {
			return helper.Fail(
				c,
				fiber.StatusUnauthorized,
				"belum terautentikasi",
			)
		}

		if !perms.Can(user.Role, permission) {
			return helper.Fail(
				c,
				fiber.StatusForbidden,
				"role "+user.Role+" tidak memiliki hak "+permission,
			)
		}

		return c.Next()
	}
}
