package main

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
)

// ok mengirim response 200 OK.
func ok(c *fiber.Ctx, message string, data any) error {
	return c.Status(fiber.StatusOK).JSON(WebResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// okList mengirim response 200 OK untuk endpoint berbentuk daftar.
func okList(c *fiber.Ctx, message string, data any, meta *Meta) error {
	return c.Status(fiber.StatusOK).JSON(WebResponse{
		Success: true,
		Message: message,
		Data:    data,
		Meta:    meta,
	})
}

// created mengirim response 201 Created.
// Location menunjukkan URL resource yang baru dibuat.
func created(c *fiber.Ctx, message string, data any, location string) error {
	c.Set("Location", location)

	return c.Status(fiber.StatusCreated).JSON(WebResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// noContent mengirim response 204 No Content.
// Response ini tidak mempunyai body.
func noContent(c *fiber.Ctx) error {
	return c.SendStatus(fiber.StatusNoContent)
}

// fail mengirim response error dengan status HTTP tertentu.
func fail(c *fiber.Ctx, status int, message string) error {
	return c.Status(status).JSON(WebResponse{
		Success: false,
		Message: message,
	})
}

// failValidation mengirim response 422 dengan detail error per field.
func failValidation(c *fiber.Ctx, errors map[string]string) error {
	return c.Status(fiber.StatusUnprocessableEntity).JSON(WebResponse{
		Success: false,
		Message: "validasi gagal",
		Errors:  errors,
	})
}

// parsePositiveInt membaca parameter integer positif dari URL.
func parsePositiveInt(c *fiber.Ctx, name string) (int, bool) {
	value, err := strconv.Atoi(c.Params(name))

	if err != nil || value < 1 {
		return 0, false
	}

	return value, true
}

// allowedSort adalah daftar field yang boleh digunakan untuk sorting.
// Client tidak boleh bebas memasukkan nama field ke query.
var allowedSort = map[string]bool{
	"id":         true,
	"nim":        true,
	"name":       true,
	"grade":      true,
	"is_active":  true,
	"created_at": true,
}

// parseListQuery membaca query string endpoint daftar student
// dan memberikan nilai bawaan yang aman.
func parseListQuery(c *fiber.Ctx) ListQuery {
	q := ListQuery{
		Page:   c.QueryInt("page", 1),
		Limit:  c.QueryInt("limit", 10),
		Search: c.Query("search"),
		Sort:   c.Query("sort", "id"),
		Order:  c.Query("order", "asc"),
	}

	// Page minimal 1.
	if q.Page < 1 {
		q.Page = 1
	}

	// Limit minimal 1.
	if q.Limit < 1 {
		q.Limit = 10
	}

	// Batas maksimum limit untuk mencegah request
	// mengambil data terlalu banyak sekaligus.
	if q.Limit > 100 {
		q.Limit = 100
	}

	// Sort hanya boleh menggunakan field yang ada
	// di dalam whitelist.
	if !allowedSort[q.Sort] {
		q.Sort = "id"
	}

	// Order hanya boleh asc atau desc.
	if q.Order != "desc" {
		q.Order = "asc"
	}

	// Filter is_active.
	if raw := c.Query("is_active"); raw != "" {
		if value, err := strconv.ParseBool(raw); err == nil {
			q.IsActive = &value
		}
	}

	// Filter nilai grade minimum.
	if raw := c.Query("min_grade"); raw != "" {
		if value, err := strconv.ParseFloat(raw, 64); err == nil {
			q.MinGrade = &value
		}
	}

	// Filter nilai grade maksimum.
	if raw := c.Query("max_grade"); raw != "" {
		if value, err := strconv.ParseFloat(raw, 64); err == nil {
			q.MaxGrade = &value
		}
	}

	return q
}