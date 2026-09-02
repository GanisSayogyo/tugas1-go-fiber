package main

import (
	"context"
	"log"

	"github.com/gofiber/fiber/v2"

	"github.com/GanisSayogyo/tugas1-go-fiber/app/handler"
	"github.com/GanisSayogyo/tugas1-go-fiber/app/repository"
	"github.com/GanisSayogyo/tugas1-go-fiber/config"
	"github.com/GanisSayogyo/tugas1-go-fiber/database"
)

func main() {
	config.LoadEnv()

	ctx := context.Background()

	db, err := database.NewPool(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Repository
	studentRepository := repository.NewStudentRepository(db)

	// Handler
	studentHandler := handler.NewStudentHandler(studentRepository)

	// Fiber
	app := fiber.New()

	// API v1
	api := app.Group("/api/v1")

	api.Get("/students", studentHandler.GetAll)
	api.Get("/students/:id", studentHandler.GetByID)
	api.Post("/students", studentHandler.Create)
	api.Put("/students/:id", studentHandler.Update)
	api.Patch("/students/:id", studentHandler.Patch)
	api.Delete("/students/:id", studentHandler.Delete)

	log.Println("Student API running on http://localhost:3000")

	log.Fatal(app.Listen(":3000"))
}
