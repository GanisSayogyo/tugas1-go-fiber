package main

import (
	"crypto/rand"
	"encoding/hex"
	"log"

	"github.com/gofiber/fiber/v2"
)

// requestIDMiddleware memberikan ID unik pada setiap request.
func requestIDMiddleware(c *fiber.Ctx) error {
	requestID := c.Get("X-Request-Id")

	// Jika client sudah mengirim X-Request-Id,
	// gunakan ID tersebut.
	if requestID == "" {
		bytes := make([]byte, 16)

		if _, err := rand.Read(bytes); err != nil {
			return err
		}

		requestID = hex.EncodeToString(bytes)
	}

	c.Set("X-Request-Id", requestID)

	return c.Next()
}

func main() {
	app := fiber.New()

	// Middleware untuk memberikan request ID.
	app.Use(requestIDMiddleware)

	// Routing API Student
	api := app.Group("/api/v1")

	api.Get("/students", listStudents)
	api.Get("/students/:id", getStudent)
	api.Post("/students", requireJSON, createStudent)
	api.Put("/students/:id", requireJSON, replaceStudent)
	api.Patch("/students/:id", requireJSON, patchStudent)
	api.Delete("/students/:id", deleteStudent)

	log.Println("Student API running on http://localhost:3000")

	log.Fatal(app.Listen(":3000"))
}
