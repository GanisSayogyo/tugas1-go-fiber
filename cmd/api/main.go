package main

import (
	"context"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"

	"github.com/GanisSayogyo/tugas1-go-fiber/app/handler"
	"github.com/GanisSayogyo/tugas1-go-fiber/app/repository"
	"github.com/GanisSayogyo/tugas1-go-fiber/app/service"
	"github.com/GanisSayogyo/tugas1-go-fiber/config"
	"github.com/GanisSayogyo/tugas1-go-fiber/database"
	"github.com/GanisSayogyo/tugas1-go-fiber/route"
)

func main() {
	config.LoadEnv()

	jwtSecret := os.Getenv("JWT_SECRET")

	if len(jwtSecret) < 32 {
		log.Fatal("JWT_SECRET harus minimal 32 karakter")
	}

	ctx := context.Background()

	db, err := database.NewPool(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Repository
	studentRepository := repository.NewStudentRepository(db)
	userRepository := repository.NewUserRepository(db)
	tokenRepository := repository.NewTokenRepository(db)

	// Service
	authService := service.NewAuthService(
		userRepository,
		tokenRepository,
	)

	// Handler
	studentHandler := handler.NewStudentHandler(studentRepository)
	authHandler := handler.NewAuthHandler(authService)

	// Fiber
	app := fiber.New()

	// Routes
	route.Setup(
		app,
		authHandler,
		studentHandler,
	)

	port := config.GetEnv("APP_PORT", "3000")

	log.Println("API running on http://localhost:" + port)

	log.Fatal(app.Listen(":" + port))
}
