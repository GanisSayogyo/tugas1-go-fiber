package service

import (
	"errors"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"

	"github.com/GanisSayogyo/tugas1-go-fiber/app/model"
	"github.com/GanisSayogyo/tugas1-go-fiber/app/repository"
	"github.com/GanisSayogyo/tugas1-go-fiber/helper"
)

type SupervisorService struct {
	supervisors repository.SupervisorRepository
	tx          *repository.TxManager
}

var errSameSupervisor = errors.New("mahasiswa sudah memiliki supervisor tersebut")

func NewSupervisorService(
	supervisors repository.SupervisorRepository,
	tx *repository.TxManager,
) *SupervisorService {
	return &SupervisorService{
		supervisors: supervisors,
		tx:          tx,
	}
}

func translateSupervisorError(err error) error {
	switch {
	case errors.Is(err, repository.ErrSupervisorNotFound):
		return helper.NotFound("supervisor tidak ditemukan")

	case errors.Is(err, repository.ErrStudentNotFound):
		return helper.NotFound("mahasiswa tidak ditemukan")

	case errors.Is(err, repository.ErrSupervisorFull):
		return helper.NewError(
			fiber.StatusConflict,
			"SUPERVISOR_FULL",
			"kuota supervisor sudah penuh",
		)

	case errors.Is(err, repository.ErrAlreadySupervised):
		return helper.NewError(
			fiber.StatusConflict,
			"ALREADY_SUPERVISED",
			"mahasiswa sudah memiliki supervisor",
		)

	case errors.Is(err, repository.ErrSupervisionNotFound):
		return helper.NewError(
			fiber.StatusConflict,
			"SUPERVISION_NOT_FOUND",
			"mahasiswa belum memiliki supervisor",
		)

	case errors.Is(err, repository.ErrSupervisorVersion):
		return helper.NewError(
			fiber.StatusConflict,
			"VERSION_CONFLICT",
			"data supervisor telah berubah, silakan muat ulang",
		)

	case errors.Is(err, repository.ErrSupervisorQuota):
		return helper.NewError(
			fiber.StatusConflict,
			"INVALID_QUOTA",
			"kuota tidak boleh lebih kecil dari jumlah mahasiswa yang dibimbing",
		)

	case errors.Is(err, errSameSupervisor):
		return helper.BadRequest("mahasiswa sudah dibimbing supervisor tersebut")

	default:
		return helper.Internal(err)
	}
}

// POST /supervisors/:id/assign
func (s *SupervisorService) Assign(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	supervisorID, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id supervisor harus berupa angka positif")
	}

	var request model.AssignSupervisorRequest
	if err := c.BodyParser(&request); err != nil || request.StudentID <= 0 {
		return helper.BadRequest("student_id harus berupa angka positif")
	}

	var assigned model.Supervisor

	err := s.tx.WithTx(ctx, func(tx repository.DBTX) error {
		repo := s.supervisors.WithTx(tx)

		// Kunci baris mahasiswa agar tidak bisa diberi
		// dua supervisor oleh transaksi bersamaan.
		if err := repo.LockStudent(ctx, request.StudentID); err != nil {
			return err
		}

		// Kunci supervisor yang akan menerima mahasiswa.
		if err := repo.LockSupervisors(ctx, supervisorID); err != nil {
			return err
		}

		// Tambah jumlah bimbingan secara kondisional.
		var err error
		assigned, err = repo.Assign(ctx, supervisorID)
		if err != nil {
			return err
		}

		// Primary key student_id juga melindungi dari
		// penetapan ganda pada level database.
		return repo.InsertSupervision(
			ctx,
			request.StudentID,
			supervisorID,
		)
	})

	if err != nil {
		return translateSupervisorError(err)
	}

	return helper.Created(
		c,
		"supervisor berhasil ditetapkan",
		assigned,
		"/api/v1/supervisors/"+strconv.Itoa(supervisorID),
	)
}

// POST /supervisions/switch
func (s *SupervisorService) Switch(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	var request model.SwitchSupervisorRequest
	if err := c.BodyParser(&request); err != nil ||
		request.StudentID <= 0 ||
		request.NewSupervisorID <= 0 {
		return helper.BadRequest(
			"student_id dan new_supervisor_id harus berupa angka positif",
		)
	}

	err := s.tx.WithTx(ctx, func(tx repository.DBTX) error {
		repo := s.supervisors.WithTx(tx)

		// 1. Kunci mahasiswa.
		if err := repo.LockStudent(ctx, request.StudentID); err != nil {
			return err
		}

		// 2. Cari supervisor lama.
		oldSupervisorID, err := repo.FindSupervision(
			ctx,
			request.StudentID,
		)
		if err != nil {
			return err
		}

		if oldSupervisorID == request.NewSupervisorID {
			return errSameSupervisor
		}

		// 3. Kunci supervisor berdasarkan urutan ID.
		// Ini mengurangi risiko deadlock ketika dua transaksi
		// melakukan switch dalam waktu bersamaan.
		if err := repo.LockSupervisors(
			ctx,
			oldSupervisorID,
			request.NewSupervisorID,
		); err != nil {
			return err
		}

		// 4. Hapus penetapan lama.
		if err := repo.DeleteSupervision(
			ctx,
			request.StudentID,
		); err != nil {
			return err
		}

		// 5. Kurangi jumlah bimbingan supervisor lama.
		if err := repo.Decrement(ctx, oldSupervisorID); err != nil {
			return err
		}

		// 6. Tambahkan mahasiswa ke supervisor baru.
		if _, err := repo.Assign(ctx, request.NewSupervisorID); err != nil {
			return err
		}

		// 7. Simpan penetapan baru.
		return repo.InsertSupervision(
			ctx,
			request.StudentID,
			request.NewSupervisorID,
		)
	})

	if err != nil {
		return translateSupervisorError(err)
	}

	return helper.Success(
		c,
		fiber.StatusOK,
		"supervisor mahasiswa berhasil dipindahkan",
		fiber.Map{
			"student_id":        request.StudentID,
			"new_supervisor_id": request.NewSupervisorID,
		},
	)
}

// PATCH /supervisors/:id
func (s *SupervisorService) Update(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	supervisorID, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id supervisor harus berupa angka positif")
	}

	var request model.UpdateSupervisorRequest
	if err := c.BodyParser(&request); err != nil {
		return helper.BadRequest("format request tidak valid")
	}

	request.Name = strings.TrimSpace(request.Name)

	if request.Name == "" || request.Quota <= 0 || request.Version <= 0 {
		return helper.BadRequest(
			"name, quota, dan version wajib valid",
		)
	}

	var updated model.Supervisor

	err := s.tx.WithTx(ctx, func(tx repository.DBTX) error {
		repo := s.supervisors.WithTx(tx)

		var err error
		updated, err = repo.Update(
			ctx,
			supervisorID,
			request.Name,
			request.Quota,
			request.Version,
		)
		return err
	})

	if err != nil {
		return translateSupervisorError(err)
	}

	return helper.Success(
		c,
		fiber.StatusOK,
		"data supervisor berhasil diperbarui",
		updated,
	)
}
