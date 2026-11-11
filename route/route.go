package route

import (
	"github.com/gofiber/fiber/v2"

	"github.com/GanisSayogyo/tugas1-go-fiber/app/handler"
	"github.com/GanisSayogyo/tugas1-go-fiber/app/service"
	"github.com/GanisSayogyo/tugas1-go-fiber/helper"
	"github.com/GanisSayogyo/tugas1-go-fiber/middleware"
)

func Setup(
	app *fiber.App,
	authHandler *handler.AuthHandler,
	studentHandler *handler.StudentHandler,
	courseService *service.CourseService,
	supervisorService *service.SupervisorService,
	permissions *helper.PermissionSet,
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

	students.Get(
		"/",
		middleware.RequirePermission(permissions, "student:list"),
		studentHandler.GetAll,
	)

	students.Get("/:id", studentHandler.GetByID)

	students.Post(
		"/",
		middleware.RequirePermission(permissions, "student:create"),
		studentHandler.Create,
	)

	students.Put("/:id", studentHandler.Update)
	students.Patch("/:id", studentHandler.Patch)

	students.Delete(
		"/:id",
		middleware.RequirePermission(permissions, "student:delete"),
		studentHandler.Delete,
	)

	// Courses dan KRS
	courses := api.Group("/courses", middleware.RequireAuth)

	courses.Get("/", courseService.List)
	courses.Get("/mine", courseService.Mine)
	courses.Get("/:id", courseService.Get)

	courses.Post("/:id/enroll", courseService.Enroll)
	courses.Delete("/:id/enroll", courseService.Cancel)

	// Supervisors dan Supervisions
	supervisors := api.Group(
		"/supervisors",
		middleware.RequireAuth,
		middleware.RequirePermission(permissions, "supervision:manage"),
	)

	supervisors.Post("/:id/assign", supervisorService.Assign)
	supervisors.Patch("/:id", supervisorService.Update)

	supervisions := api.Group(
		"/supervisions",
		middleware.RequireAuth,
		middleware.RequirePermission(permissions, "supervision:manage"),
	)

	supervisions.Post("/switch", supervisorService.Switch)
}
