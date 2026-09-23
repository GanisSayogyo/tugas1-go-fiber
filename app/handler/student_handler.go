package handler

import (
	"errors"
	"math"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"

	"github.com/GanisSayogyo/tugas1-go-fiber/app/model"
	"github.com/GanisSayogyo/tugas1-go-fiber/app/repository"
	"github.com/GanisSayogyo/tugas1-go-fiber/app/service"
	"github.com/GanisSayogyo/tugas1-go-fiber/helper"
)

type StudentHandler struct {
	repo        *repository.StudentRepository
	permissions *helper.PermissionSet
}

func NewStudentHandler(
	repo *repository.StudentRepository,
	permissions *helper.PermissionSet,
) *StudentHandler {
	return &StudentHandler{
		repo:        repo,
		permissions: permissions,
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

	currentUser, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Fail(c, fiber.StatusUnauthorized, "belum terautentikasi")
	}

	if !service.CanAccessStudent(
		currentUser,
		student.OwnerID,
		h.permissions,
		"student:read:any",
	) {
		return helper.Fail(c, fiber.StatusForbidden, "tidak memiliki akses ke student ini")
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

	currentUser, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Fail(c, fiber.StatusUnauthorized, "belum terautentikasi")
	}

	student := &model.Student{
		NIM:      input.NIM,
		Name:     input.Name,
		Grade:    input.Grade,
		IsActive: input.IsActive,
		OwnerID:  currentUser.UserID,
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

// PUT /api/v1/students/:id
func (h *StudentHandler) Update(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))

	if err != nil || id < 1 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "id harus berupa angka positif",
		})
	}

	currentUser, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Fail(c, fiber.StatusUnauthorized, "belum terautentikasi")
	}

	existing, err := h.repo.FindByID(c.Context(), id)

	if errors.Is(err, repository.ErrNotFound) {
		return helper.Fail(c, fiber.StatusNotFound, "student tidak ditemukan")
	}

	if err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal mengambil student")
	}

	if !service.CanAccessStudent(
		currentUser,
		existing.OwnerID,
		h.permissions,
		"student:update:any",
	) {
		return helper.Fail(c, fiber.StatusForbidden, "tidak memiliki akses ke student ini")
	}

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

	// PUT wajib mengirim seluruh field.
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
		ID:       id,
		NIM:      input.NIM,
		Name:     input.Name,
		Grade:    input.Grade,
		IsActive: input.IsActive,
	}

	err = h.repo.Update(c.Context(), student)

	if errors.Is(err, repository.ErrNotFound) {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "student tidak ditemukan",
		})
	}

	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "unique") {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"success": false,
				"message": "NIM sudah digunakan",
			})
		}

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "gagal memperbarui student",
		})
	}

	updated, err := h.repo.FindByID(c.Context(), id)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "gagal mengambil student setelah diperbarui",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "student berhasil diperbarui",
		"data":    updated,
	})
}

// PATCH /api/v1/students/:id
func (h *StudentHandler) Patch(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))

	if err != nil || id < 1 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "id harus berupa angka positif",
		})
	}

	var input struct {
		NIM      *string  `json:"nim,omitempty"`
		Name     *string  `json:"name,omitempty"`
		Grade    *float64 `json:"grade,omitempty"`
		IsActive *bool    `json:"is_active,omitempty"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "body JSON tidak valid",
		})
	}

	if input.NIM != nil {
		value := strings.TrimSpace(*input.NIM)

		if value == "" {
			return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
				"success": false,
				"message": "validasi gagal",
				"errors": fiber.Map{
					"nim": "NIM tidak boleh kosong",
				},
			})
		}

		input.NIM = &value
	}

	if input.Name != nil {
		value := strings.TrimSpace(*input.Name)

		if value == "" {
			return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
				"success": false,
				"message": "validasi gagal",
				"errors": fiber.Map{
					"name": "nama tidak boleh kosong",
				},
			})
		}

		input.Name = &value
	}

	if input.Grade != nil {
		if *input.Grade < 0 || *input.Grade > 100 {
			return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
				"success": false,
				"message": "validasi gagal",
				"errors": fiber.Map{
					"grade": "grade harus berada di antara 0 dan 100",
				},
			})
		}
	}

	updated, err := h.repo.Patch(
		c.Context(),
		id,
		input.NIM,
		input.Name,
		input.Grade,
		input.IsActive,
	)

	if errors.Is(err, repository.ErrNotFound) {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "student tidak ditemukan",
		})
	}

	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "unique") {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"success": false,
				"message": "NIM sudah digunakan",
			})
		}

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "gagal memperbarui student",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "student berhasil diperbarui",
		"data":    updated,
	})
}

// DELETE /api/v1/students/:id
func (h *StudentHandler) Delete(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))

	if err != nil || id < 1 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "id harus berupa angka positif",
		})
	}

	err = h.repo.Delete(c.Context(), id)

	if errors.Is(err, repository.ErrNotFound) {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "student tidak ditemukan",
		})
	}

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "gagal menghapus student",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "student berhasil dihapus",
	})
}
