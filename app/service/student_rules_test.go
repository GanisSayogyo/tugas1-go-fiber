package service

import (
	"testing"

	"github.com/GanisSayogyo/tugas1-go-fiber/app/model"
)

func TestValidateCreate(t *testing.T) {
	tests := []struct {
		name      string
		request   model.CreateStudentRequest
		wantError bool
	}{
		{
			name: "valid",
			request: model.CreateStudentRequest{
				NIM:      "20250001",
				Name:     "Andi Pratama",
				Grade:    85,
				IsActive: true,
			},
			wantError: false,
		},
		{
			name: "NIM kosong",
			request: model.CreateStudentRequest{
				NIM:      "",
				Name:     "Andi",
				Grade:    85,
				IsActive: true,
			},
			wantError: true,
		},
		{
			name: "nama kosong",
			request: model.CreateStudentRequest{
				NIM:      "20250001",
				Name:     "",
				Grade:    85,
				IsActive: true,
			},
			wantError: true,
		},
		{
			name: "grade tidak valid",
			request: model.CreateStudentRequest{
				NIM:      "20250001",
				Name:     "Andi",
				Grade:    150,
				IsActive: true,
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := ValidateCreate(tt.request)

			hasError := len(errs) > 0

			if hasError != tt.wantError {
				t.Fatalf(
					"wantError=%v, got=%v, errors=%v",
					tt.wantError,
					hasError,
					errs,
				)
			}
		})
	}
}

func TestValidateReplace(t *testing.T) {
	valid := model.UpdateStudentRequest{
		NIM:      "20250001",
		Name:     "Andi",
		Grade:    85,
		IsActive: true,
	}

	if errs := ValidateReplace(valid); len(errs) != 0 {
		t.Fatalf("request valid menghasilkan error: %v", errs)
	}

	invalid := model.UpdateStudentRequest{
		NIM:      "",
		Name:     "",
		Grade:    150,
		IsActive: true,
	}

	errs := ValidateReplace(invalid)

	if len(errs) != 3 {
		t.Fatalf("expected 3 errors, got %d: %v", len(errs), errs)
	}
}

func TestApplyPatch(t *testing.T) {
	current := model.Student{
		ID:       1,
		NIM:      "20250001",
		Name:     "Andi",
		Grade:    85,
		IsActive: true,
	}

	newName := "Andi PATCH"
	newGrade := 99.0

	req := model.PatchStudentRequest{
		Name:  &newName,
		Grade: &newGrade,
	}

	result, errs := ApplyPatch(current, req)

	if len(errs) != 0 {
		t.Fatalf("tidak seharusnya ada error: %v", errs)
	}

	if result.Name != "Andi PATCH" {
		t.Errorf("name tidak berubah: %s", result.Name)
	}

	if result.Grade != 99 {
		t.Errorf("grade tidak berubah: %v", result.Grade)
	}

	if result.NIM != "20250001" {
		t.Errorf("NIM yang tidak dikirim ikut berubah: %s", result.NIM)
	}

	if result.IsActive != true {
		t.Error("is_active yang tidak dikirim ikut berubah")
	}
}

func TestCountTotalPages(t *testing.T) {
	tests := []struct {
		total int
		limit int
		want  int
	}{
		{0, 10, 0},
		{1, 10, 1},
		{10, 10, 1},
		{11, 10, 2},
		{137, 20, 7},
	}

	for _, tt := range tests {
		got := CountTotalPages(tt.total, tt.limit)

		if got != tt.want {
			t.Errorf(
				"total=%d limit=%d: expected %d, got %d",
				tt.total,
				tt.limit,
				tt.want,
				got,
			)
		}
	}
}
