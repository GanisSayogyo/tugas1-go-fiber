package handler

import (
	"errors"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"

	"encoding/csv"
	"fmt"
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

func logStudentRequest(c *fiber.Ctx) {
	user, ok := helper.CurrentUser(c)
	if !ok {
		log.Printf(
			"student request method=%s path=%s user_id=unknown role=unknown",
			c.Method(), c.Path(),
		)
		return
	}

	log.Printf(
		"student request method=%s path=%s user_id=%d role=%s",
		c.Method(), c.Path(), user.UserID, user.Role,
	)
}

func validationFailed(c *fiber.Ctx, validationErrors map[string]string) error {
	return helper.Validation(validationErrors)
}

func parseStudentID(c *fiber.Ctx) (int, error) {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil || id < 1 {
		return 0, fiber.NewError(fiber.StatusBadRequest, "id harus berupa angka positif")
	}
	return id, nil
}

func negotiateStudentFormat(c *fiber.Ctx) (string, error) {
	accept := strings.TrimSpace(c.Get("Accept"))

	if accept == "" || accept == "*/*" {
		return "json", nil
	}

	if strings.Contains(accept, "text/csv") {
		return "csv", nil
	}

	if strings.Contains(accept, "application/json") {
		return "json", nil
	}

	return "", fiber.NewError(
		fiber.StatusNotAcceptable,
		"format response yang diminta tidak tersedia",
	)
}

// GET /api/v1/students
func (h *StudentHandler) GetAll(c *fiber.Ctx) error {
	logStudentRequest(c)

	limit, err := strconv.Atoi(c.Query("limit", "10"))
	if err != nil || limit < 1 || limit > 100 {
		return helper.Fail(c, fiber.StatusBadRequest, "limit harus berada di antara 1 dan 100")
	}

	var cursorCreatedAt *time.Time
	var cursorID *int

	cursor := c.Query("cursor")
	if cursor != "" {
		createdAt, id, err := helper.DecodeCursor(cursor)
		if err != nil {
			return helper.Fail(c, fiber.StatusBadRequest, "cursor tidak valid")
		}

		cursorCreatedAt = &createdAt
		cursorID = &id
	}

	students, err := h.repo.FindAfterCursor(
		c.Context(),
		cursorCreatedAt,
		cursorID,
		limit+1,
	)
	if err != nil {
		log.Printf("ERROR: gagal mengambil student: %v", err)
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal mengambil data student")
	}

	hasMore := len(students) > limit
	if hasMore {
		students = students[:limit]
	}

	meta := fiber.Map{
		"limit":    limit,
		"has_more": hasMore,
	}

	if hasMore && len(students) > 0 {
		lastStudent := students[len(students)-1]
		meta["next_cursor"] = helper.EncodeCursor(
			lastStudent.CreatedAt,
			lastStudent.ID,
		)
	}

	format, err := negotiateStudentFormat(c)
	if err != nil {
		return err
	}

	if format == "csv" {
		csvRows := [][]string{
			{"id", "nim", "name", "grade", "is_active", "owner_id", "created_at"},
		}

		for _, student := range students {
			csvRows = append(csvRows, []string{
				strconv.Itoa(student.ID),
				student.NIM,
				student.Name,
				fmt.Sprintf("%g", student.Grade),
				strconv.FormatBool(student.IsActive),
				strconv.Itoa(student.OwnerID),
				student.CreatedAt.Format(time.RFC3339Nano),
			})
		}

		var output strings.Builder
		writer := csv.NewWriter(&output)

		if err := writer.WriteAll(csvRows); err != nil {
			return helper.Fail(c, fiber.StatusInternalServerError, "gagal membuat CSV")
		}

		c.Set(fiber.HeaderContentType, "text/csv; charset=utf-8")
		c.Set(fiber.HeaderContentDisposition, `attachment; filename="students.csv"`)
		return c.Status(fiber.StatusOK).SendString(output.String())
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "student berhasil ditemukan",
		"data":    students,
		"meta":    meta,
	})
}

// GET /api/v1/students/:id
func (h *StudentHandler) GetByID(c *fiber.Ctx) error {
	logStudentRequest(c)

	id, err := parseStudentID(c)
	if err != nil {
		return helper.Fail(c, fiber.StatusBadRequest, err.Error())
	}

	student, err := h.repo.FindByID(c.Context(), id)
	if errors.Is(err, repository.ErrNotFound) {
		return helper.Fail(c, fiber.StatusNotFound, "student tidak ditemukan")
	}
	if err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal mengambil student")
	}

	currentUser, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Fail(c, fiber.StatusUnauthorized, "belum terautentikasi")
	}

	if !service.CanAccessStudent(
		currentUser, student.OwnerID, h.permissions, "student:read:any",
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
	logStudentRequest(c)

	var input model.CreateStudentRequest
	if err := c.BodyParser(&input); err != nil {
		return helper.Fail(c, fiber.StatusBadRequest, "body JSON tidak valid")
	}

	input.NIM = strings.TrimSpace(input.NIM)
	input.Name = strings.TrimSpace(input.Name)

	if validationErrors := helper.ValidateStruct(input); len(validationErrors) > 0 {
		return validationFailed(c, validationErrors)
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
			return helper.Fail(c, fiber.StatusConflict, "NIM sudah digunakan")
		}
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal membuat student")
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "student berhasil dibuat",
		"data":    student,
	})
}

// PUT /api/v1/students/:id
func (h *StudentHandler) Update(c *fiber.Ctx) error {
	logStudentRequest(c)

	id, err := parseStudentID(c)
	if err != nil {
		return helper.Fail(c, fiber.StatusBadRequest, err.Error())
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
		currentUser, existing.OwnerID, h.permissions, "student:update:any",
	) {
		return helper.Fail(c, fiber.StatusForbidden, "tidak memiliki akses ke student ini")
	}

	var input model.UpdateStudentRequest
	if err := c.BodyParser(&input); err != nil {
		return helper.Fail(c, fiber.StatusBadRequest, "body JSON tidak valid")
	}

	input.NIM = strings.TrimSpace(input.NIM)
	input.Name = strings.TrimSpace(input.Name)

	if validationErrors := helper.ValidateStruct(input); len(validationErrors) > 0 {
		return validationFailed(c, validationErrors)
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
		return helper.Fail(c, fiber.StatusNotFound, "student tidak ditemukan")
	}
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "unique") {
			return helper.Fail(c, fiber.StatusConflict, "NIM sudah digunakan")
		}
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal memperbarui student")
	}

	updated, err := h.repo.FindByID(c.Context(), id)
	if err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal mengambil student setelah diperbarui")
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "student berhasil diperbarui",
		"data":    updated,
	})
}

// PATCH /api/v1/students/:id
func (h *StudentHandler) Patch(c *fiber.Ctx) error {
	logStudentRequest(c)

	id, err := parseStudentID(c)
	if err != nil {
		return helper.Fail(c, fiber.StatusBadRequest, err.Error())
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
		currentUser, existing.OwnerID, h.permissions, "student:update:any",
	) {
		return helper.Fail(c, fiber.StatusForbidden, "tidak memiliki akses ke student ini")
	}

	var input model.PatchStudentRequest
	if err := c.BodyParser(&input); err != nil {
		return helper.Fail(c, fiber.StatusBadRequest, "body JSON tidak valid")
	}

	if input.NIM == nil &&
		input.Name == nil &&
		input.Grade == nil &&
		input.IsActive == nil {
		return helper.Fail(
			c,
			fiber.StatusBadRequest,
			"minimal satu field harus diisi",
		)
	}

	if input.NIM != nil {
		value := strings.TrimSpace(*input.NIM)
		input.NIM = &value
	}
	if input.Name != nil {
		value := strings.TrimSpace(*input.Name)
		input.Name = &value
	}

	if validationErrors := helper.ValidateStruct(input); len(validationErrors) > 0 {
		return validationFailed(c, validationErrors)
	}

	updated, err := h.repo.Patch(
		c.Context(), id, input.NIM, input.Name, input.Grade, input.IsActive,
	)
	if errors.Is(err, repository.ErrNotFound) {
		return helper.Fail(c, fiber.StatusNotFound, "student tidak ditemukan")
	}
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "unique") {
			return helper.Fail(c, fiber.StatusConflict, "NIM sudah digunakan")
		}
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal memperbarui student")
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "student berhasil diperbarui",
		"data":    updated,
	})
}

// DELETE /api/v1/students/:id
func (h *StudentHandler) Delete(c *fiber.Ctx) error {
	logStudentRequest(c)

	id, err := parseStudentID(c)
	if err != nil {
		return helper.Fail(c, fiber.StatusBadRequest, err.Error())
	}

	err = h.repo.Delete(c.Context(), id)
	if errors.Is(err, repository.ErrNotFound) {
		return helper.Fail(c, fiber.StatusNotFound, "student tidak ditemukan")
	}
	if err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal menghapus student")
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "student berhasil dihapus",
	})
}
