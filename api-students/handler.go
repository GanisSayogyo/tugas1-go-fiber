package main

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
)

// students menyimpan seluruh data student di memory.
var students = []Student{
	{
		ID:        1,
		NIM:       "20250001",
		Name:      "Andi Pratama",
		Grade:     85,
		IsActive:  true,
		CreatedAt: time.Now(),
	},
	{
		ID:        2,
		NIM:       "20250002",
		Name:      "Budi Santoso",
		Grade:     78,
		IsActive:  true,
		CreatedAt: time.Now(),
	},
	{
		ID:        3,
		NIM:       "20250003",
		Name:      "Citra Lestari",
		Grade:     92,
		IsActive:  true,
		CreatedAt: time.Now(),
	},
	{
		ID:        4,
		NIM:       "20250004",
		Name:      "Dimas Saputra",
		Grade:     65,
		IsActive:  false,
		CreatedAt: time.Now(),
	},
	{
		ID:        5,
		NIM:       "20250005",
		Name:      "Eka Putri",
		Grade:     88,
		IsActive: true,
		CreatedAt:  time.Now(),
	},
}

// mu menjaga agar akses ke data students aman ketika
// API menerima beberapa request secara bersamaan.
var mu sync.RWMutex

// nextStudentID menyimpan ID berikutnya yang akan digunakan.
var nextStudentID = 6

// listStudents menangani GET /api/v1/students.
func listStudents(c *fiber.Ctx) error {
	mu.RLock()
	defer mu.RUnlock()

	q := parseListQuery(c)

	// Salin data agar proses filtering dan sorting
	// tidak mengubah data asli.
	result := make([]Student, len(students))
	copy(result, students)

	// Search berdasarkan NIM atau nama.
	if q.Search != "" {
		search := strings.ToLower(strings.TrimSpace(q.Search))

		filtered := make([]Student, 0)

		for _, student := range result {
			if strings.Contains(strings.ToLower(student.Name), search) ||
				strings.Contains(strings.ToLower(student.NIM), search) {
				filtered = append(filtered, student)
			}
		}

		result = filtered
	}

	// Filter is_active.
	if q.IsActive != nil {
		filtered := make([]Student, 0)

		for _, student := range result {
			if student.IsActive == *q.IsActive {
				filtered = append(filtered, student)
			}
		}

		result = filtered
	}

	// Filter minimum grade.
	if q.MinGrade != nil {
		filtered := make([]Student, 0)

		for _, student := range result {
			if student.Grade >= *q.MinGrade {
				filtered = append(filtered, student)
			}
		}

		result = filtered
	}

	// Filter maximum grade.
	if q.MaxGrade != nil {
		filtered := make([]Student, 0)

		for _, student := range result {
			if student.Grade <= *q.MaxGrade {
				filtered = append(filtered, student)
			}
		}

		result = filtered
	}

	// Sorting berdasarkan field yang sudah melalui whitelist.
	sort.SliceStable(result, func(i, j int) bool {
		less := false

		switch q.Sort {
		case "id":
			less = result[i].ID < result[j].ID
		case "nim":
			less = result[i].NIM < result[j].NIM
		case "name":
			less = strings.ToLower(result[i].Name) <
				strings.ToLower(result[j].Name)
		case "grade":
			less = result[i].Grade < result[j].Grade
		case "is_active":
			less = !result[i].IsActive && result[j].IsActive
		case "created_at":
			less = result[i].CreatedAt.Before(result[j].CreatedAt)
		}

		if q.Order == "desc" {
			return !less
		}

		return less
	})

	// Total data setelah filter/search.
	total := len(result)

	// Pagination.
	start := (q.Page - 1) * q.Limit

	if start >= total {
		result = []Student{}
	} else {
		end := start + q.Limit

		if end > total {
			end = total
		}

		result = result[start:end]
	}

	totalPages := 0

	if total > 0 {
		totalPages = (total + q.Limit - 1) / q.Limit
	}

	meta := &Meta{
		Page:       q.Page,
		Limit:      q.Limit,
		Total:      total,
		TotalPages: totalPages,
	}

	return okList(
		c,
		"student berhasil ditemukan",
		result,
		meta,
	)
}

// getStudent menangani GET /api/v1/students/:id.
func getStudent(c *fiber.Ctx) error {
	id, valid := parsePositiveInt(c, "id")

	if !valid {
		return fail(
			c,
			fiber.StatusBadRequest,
			"id harus berupa angka positif",
		)
	}

	mu.RLock()
	defer mu.RUnlock()

	for _, student := range students {
		if student.ID == id {
			return ok(
				c,
				"student berhasil ditemukan",
				student,
			)
		}
	}

	return fail(
		c,
		fiber.StatusNotFound,
		"student tidak ditemukan",
	)
}

// studentLocation membuat URL resource student.
func studentLocation(c *fiber.Ctx, id int) string {
	return fmt.Sprintf(
		"%s/api/v1/students/%d",
		c.BaseURL(),
		id,
	)
}

// parseStudentID mengambil ID dari parameter URL
// dan mengubahnya menjadi integer.
func parseStudentID(c *fiber.Ctx) (int, error) {
	rawID := c.Params("id")

	id, err := strconv.Atoi(rawID)

	if err != nil || id < 1 {
		return 0, fmt.Errorf("invalid student id")
	}

	return id, nil
}

// createStudent menangani POST /api/v1/students.
func createStudent(c *fiber.Ctx) error {
	var req CreateStudentRequest

	// Membaca JSON request body.
	if err := c.BodyParser(&req); err != nil {
		return fail(
			c,
			fiber.StatusBadRequest,
			"format JSON tidak valid",
		)
	}

	// Rapikan input string.
	req.NIM = strings.TrimSpace(req.NIM)
	req.Name = strings.TrimSpace(req.Name)

	// Validasi field wajib.
	errors := make(map[string]string)

	if req.NIM == "" {
		errors["nim"] = "NIM wajib diisi"
	}

	if req.Name == "" {
		errors["name"] = "nama wajib diisi"
	}

	// Nilai grade harus berada pada rentang 0-100.
	if req.Grade < 0 || req.Grade > 100 {
		errors["grade"] = "grade harus berada di antara 0 dan 100"
	}

	if len(errors) > 0 {
		return failValidation(c, errors)
	}

	mu.Lock()
	defer mu.Unlock()

	// NIM harus unik.
	for _, student := range students {
		if strings.EqualFold(student.NIM, req.NIM) {
			return fail(
				c,
				fiber.StatusConflict,
				"NIM sudah digunakan",
			)
		}
	}

	// Buat student baru.
	student := Student{
		ID:        nextStudentID,
		NIM:       req.NIM,
		Name:      req.Name,
		Grade:     req.Grade,
		IsActive:  req.IsActive,
		CreatedAt: time.Now(),
	}

	// Simpan ke memory.
	students = append(students, student)
	nextStudentID++

	return created(
		c,
		"student berhasil dibuat",
		student,
		studentLocation(c, student.ID),
	)
}

// replaceStudent menangani PUT /api/v1/students/:id.
// PUT mengganti seluruh data student.
func replaceStudent(c *fiber.Ctx) error {
	id, valid := parsePositiveInt(c, "id")

	if !valid {
		return fail(
			c,
			fiber.StatusBadRequest,
			"id harus berupa angka positif",
		)
	}

	var req ReplaceStudentRequest

	// Membaca JSON request body.
	if err := c.BodyParser(&req); err != nil {
		return fail(
			c,
			fiber.StatusBadRequest,
			"format JSON tidak valid",
		)
	}

	// Rapikan input string.
	req.NIM = strings.TrimSpace(req.NIM)
	req.Name = strings.TrimSpace(req.Name)

	// Validasi seluruh field.
	errors := make(map[string]string)

	if req.NIM == "" {
		errors["nim"] = "NIM wajib diisi"
	}

	if req.Name == "" {
		errors["name"] = "nama wajib diisi"
	}

	if req.Grade < 0 || req.Grade > 100 {
		errors["grade"] = "grade harus berada di antara 0 dan 100"
	}

	if len(errors) > 0 {
		return failValidation(c, errors)
	}

	mu.Lock()
	defer mu.Unlock()

	// Cari student berdasarkan ID.
	index := -1

	for i, student := range students {
		if student.ID == id {
			index = i
			break
		}
	}

	if index == -1 {
		return fail(
			c,
			fiber.StatusNotFound,
			"student tidak ditemukan",
		)
	}

	// Pastikan NIM tidak digunakan student lain.
	for i, student := range students {
		if i != index && strings.EqualFold(student.NIM, req.NIM) {
			return fail(
				c,
				fiber.StatusConflict,
				"NIM sudah digunakan",
			)
		}
	}

	// Pertahankan ID dan CreatedAt,
	// lalu ganti seluruh field yang dimiliki resource.
	students[index].NIM = req.NIM
	students[index].Name = req.Name
	students[index].Grade = req.Grade
	students[index].IsActive = req.IsActive

	return ok(
		c,
		"student berhasil diperbarui",
		students[index],
	)
}