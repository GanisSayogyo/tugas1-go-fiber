package model

import "time"

// Student merepresentasikan data mahasiswa dari tabel students.
type Student struct {
	ID        int       `json:"id"`
	NIM       string    `json:"nim"`
	Name      string    `json:"name"`
	Grade     float64   `json:"grade"`
	IsActive  bool      `json:"is_active"`
	OwnerID   int       `json:"owner_id"`
	CreatedAt time.Time `json:"created_at"`
}

// ListQuery berisi parameter untuk pengambilan daftar student.
type ListQuery struct {
	Page     int
	Limit    int
	Search   string
	IsActive *bool
	MinGrade *float64
	MaxGrade *float64
	Sort     string
	Order    string
}

// Meta berisi informasi pagination.
type Meta struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

// WebResponse adalah format response API yang seragam.
type WebResponse struct {
	Success bool              `json:"success"`
	Message string            `json:"message"`
	Data    any               `json:"data,omitempty"`
	Meta    *Meta             `json:"meta,omitempty"`
	Errors  map[string]string `json:"errors,omitempty"`
}

// CreateStudentRequest digunakan untuk membuat student baru.
type CreateStudentRequest struct {
	NIM      string  `json:"nim"`
	Name     string  `json:"name"`
	Grade    float64 `json:"grade"`
	IsActive bool    `json:"is_active"`
}

// UpdateStudentRequest digunakan untuk mengganti seluruh data student.
type UpdateStudentRequest struct {
	NIM      string  `json:"nim"`
	Name     string  `json:"name"`
	Grade    float64 `json:"grade"`
	IsActive bool    `json:"is_active"`
}

// PatchStudentRequest digunakan untuk memperbarui sebagian data student.
type PatchStudentRequest struct {
	NIM      *string  `json:"nim,omitempty"`
	Name     *string  `json:"name,omitempty"`
	Grade    *float64 `json:"grade,omitempty"`
	IsActive *bool    `json:"is_active,omitempty"`
}
