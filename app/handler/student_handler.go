package handler

import (
	"errors"
	"math"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"

	"github.com/GanisSayogyo/tugas1-go-fiber/app/model"
	"github.com/GanisSayogyo/tugas1-go-fiber/app/repository"
)

type StudentHandler struct {
	repo *repository.StudentRepository
}

func NewStudentHandler(repo *repository.StudentRepository) *StudentHandler {
	return &StudentHandler{
		repo: repo,
	}
}

// GET /api/v1/students
func (h *StudentHandler) GetAll(c *fiber.Ctx) error {
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 10)

	if page < 1 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "page harus lebih besar dari 0",
		})
	}

	if limit < 1 || limit > 100 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "limit harus berada di antara 1 dan 100",
		})
	}

	search := strings.TrimSpace(c.Query("search"))

	var isActive *bool

	if value := c.Context().QueryArgs().Peek("is_active"); len(value) > 0 {
		parsed, err := strconv.ParseBool(string(value))
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": "is_active harus berupa true atau false",
			})
		}

		isActive = &parsed
	}

	var minGrade *float64
	if value := c.Query("min_grade"); value != "" {
		parsed, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": "min_grade harus berupa angka",
			})
		}

		minGrade = &parsed
	}

	var maxGrade *float64
	if value := c.Query("max_grade"); value != "" {
		parsed, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": "max_grade harus berupa angka",
			})
		}

		maxGrade = &parsed
	}

	// Whitelist sorting untuk mencegah SQL Injection.
	sortField := c.Query("sort", "id")

	allowedSort := map[string]bool{
		"id":         true,
		"nim":        true,
		"name":       true,
		"grade":      true,
		"is_active":  true,
		"created_at": true,
	}

	if !allowedSort[sortField] {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "sort field tidak valid",
		})
	}

	sortOrder := strings.ToLower(c.Query("order", "asc"))

	if sortOrder != "asc" && sortOrder != "desc" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "order harus asc atau desc",
		})
	}

	offset := (page - 1) * limit

	students, total, err := h.repo.FindAll(
		c.Context(),
		search,
		isActive,
		minGrade,
		maxGrade,
		sortField,
		sortOrder,
		limit,
		offset,
	)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "gagal mengambil data student",
		})
	}

	totalPages := 0
	if total > 0 {
		totalPages = int(math.Ceil(float64(total) / float64(limit)))
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "student berhasil ditemukan",
		"data":    students,
		"meta": fiber.Map{
			"page":        page,
			"limit":       limit,
			"total":       total,
			"total_pages": totalPages,
		},
	})
}

// GET /api/v1/students/:id
func (h *StudentHandler) GetByID(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))

	if err != nil || id < 1 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "id harus berupa angka positif",
		})
	}

	student, err := h.repo.FindByID(c.Context(), id)

	if errors.Is(err, repository.ErrNotFound) {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "student tidak ditemukan",
		})
	}

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "gagal mengambil student",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "student berhasil ditemukan",
		"data":    student,
	})
}

// POST /api/v1/students
func (h *StudentHandler) Create(c *fiber.Ctx) error {
	var input struct {
		NIM      string  `json:"nim"`
		Name     string  `json:"name"`
		Grade    float64 `json:"grade"`
		IsActive bool    `json:"is_active"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "body JSON tidak valid",
		})
	}

	input.NIM = strings.TrimSpace(input.NIM)
	input.Name = strings.TrimSpace(input.Name)

	if input.NIM == "" {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
			"success": false,
			"message": "validasi gagal",
			"errors": fiber.Map{
				"nim": "NIM wajib diisi",
			},
		})
	}

	if input.Name == "" {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
			"success": false,
			"message": "validasi gagal",
			"errors": fiber.Map{
				"name": "nama wajib diisi",
			},
		})
	}

	if input.Grade < 0 || input.Grade > 100 {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
			"success": false,
			"message": "validasi gagal",
			"errors": fiber.Map{
				"grade": "grade harus berada di antara 0 dan 100",
			},
		})
	}

	student := &model.Student{
		NIM:      input.NIM,
		Name:     input.Name,
		Grade:    input.Grade,
		IsActive: input.IsActive,
	}

	if err := h.repo.Create(c.Context(), student); err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "unique") {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"success": false,
				"message": "NIM sudah digunakan",
			})
		}

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "gagal membuat student",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "student berhasil dibuat",
		"data":    student,
	})
}
