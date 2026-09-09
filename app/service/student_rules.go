package service

import (
	"strings"

	"github.com/GanisSayogyo/tugas1-go-fiber/app/model"
)

// ValidateCreate memeriksa request POST student.
func ValidateCreate(req model.CreateStudentRequest) map[string]string {
	errs := map[string]string{}

	if strings.TrimSpace(req.NIM) == "" {
		errs["nim"] = "NIM wajib diisi"
	}

	if strings.TrimSpace(req.Name) == "" {
		errs["name"] = "nama wajib diisi"
	}

	if req.Grade < 0 || req.Grade > 100 {
		errs["grade"] = "grade harus berada di antara 0 dan 100"
	}

	return errs
}

// ValidateReplace memeriksa request PUT student.
// PUT wajib mengirim seluruh field yang diperlukan.
func ValidateReplace(req model.UpdateStudentRequest) map[string]string {
	errs := map[string]string{}

	if strings.TrimSpace(req.NIM) == "" {
		errs["nim"] = "NIM wajib diisi"
	}

	if strings.TrimSpace(req.Name) == "" {
		errs["name"] = "nama wajib diisi"
	}

	if req.Grade < 0 || req.Grade > 100 {
		errs["grade"] = "grade harus berada di antara 0 dan 100"
	}

	return errs
}

// ApplyPatch menerapkan hanya field yang dikirim pada PATCH.
//
// Field bernilai nil tidak diubah.
func ApplyPatch(
	current model.Student,
	req model.PatchStudentRequest,
) (model.Student, map[string]string) {

	errs := map[string]string{}

	if req.NIM != nil {
		value := strings.TrimSpace(*req.NIM)

		if value == "" {
			errs["nim"] = "NIM tidak boleh kosong"
		} else {
			current.NIM = value
		}
	}

	if req.Name != nil {
		value := strings.TrimSpace(*req.Name)

		if value == "" {
			errs["name"] = "nama tidak boleh kosong"
		} else {
			current.Name = value
		}
	}

	if req.Grade != nil {
		if *req.Grade < 0 || *req.Grade > 100 {
			errs["grade"] = "grade harus berada di antara 0 dan 100"
		} else {
			current.Grade = *req.Grade
		}
	}

	if req.IsActive != nil {
		current.IsActive = *req.IsActive
	}

	return current, errs
}

// CountTotalPages menghitung jumlah halaman pagination.
func CountTotalPages(total, limit int) int {
	if limit <= 0 {
		return 0
	}

	return (total + limit - 1) / limit
}
