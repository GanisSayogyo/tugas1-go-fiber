package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
)

func main() {
	app := fiber.New()

	// Routing API Student
	api := app.Group("/api/v1")

	api.Get("/students", listStudents)
	api.Get("/students/:id", getStudent)
	api.Post("/students", createStudent)
	api.Put("/students/:id", replaceStudent)

	log.Println("Student API running on http://localhost:3000")

	log.Fatal(app.Listen(":3000"))
}