package helper

import (
	"github.com/gofiber/fiber/v2"

	"github.com/GanisSayogyo/tugas1-go-fiber/app/model"
)

func CurrentUser(c *fiber.Ctx) (model.AuthUser, bool) {
	value := c.Locals("auth_user")

	user, ok := value.(model.AuthUser)
	return user, ok
}
