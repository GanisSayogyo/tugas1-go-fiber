package route

import (
	"github.com/gofiber/fiber/v2"

	"github.com/GanisSayogyo/tugas1-go-fiber/app/handler"
	"github.com/GanisSayogyo/tugas1-go-fiber/middleware"
)

func Setup(
	app *fiber.App,
	authHandler *handler.AuthHandler,
	studentHandler *handler.StudentHandler,
) {
	api := app.Group("/api/v1")

	// Public
	api.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"success": true,
			"message": "API is running",
		})
	})

	auth := api.Group("/auth")

	auth.Post("/register", authHandler.Register)
	auth.Post("/login", authHandler.Login)
	auth.Post("/refresh", authHandler.Refresh)
	auth.Post("/logout", authHandler.Logout)

	// Protected
	auth.Get("/me", middleware.RequireAuth, authHandler.Me)

	students := api.Group("/students", middleware.RequireAuth)

	students.Get("/", studentHandler.GetAll)
	students.Get("/:id", studentHandler.GetByID)
	students.Post("/", studentHandler.Create)
	students.Put("/:id", studentHandler.Update)
	students.Patch("/:id", studentHandler.Patch)
	students.Delete("/:id", studentHandler.Delete)
}
