package service

import (
	"errors"
	"strconv"

	"github.com/GanisSayogyo/tugas1-go-fiber/app/model"
	"github.com/gofiber/fiber/v2"

	"github.com/GanisSayogyo/tugas1-go-fiber/app/repository"
	"github.com/GanisSayogyo/tugas1-go-fiber/helper"
)

type CourseService struct {
	courses repository.CourseRepository
	tx      *repository.TxManager
}

var errCreditLimit = errors.New("batas SKS terlampaui")

func NewCourseService(
	courses repository.CourseRepository,
	tx *repository.TxManager,
) *CourseService {
	return &CourseService{
		courses: courses,
		tx:      tx,
	}
}

// translateCourseError menerjemahkan error repository ke HTTP error.
func translateCourseError(err error) error {
	switch {
	case errors.Is(err, repository.ErrCourseNotFound):
		return helper.NotFound("kelas tidak ditemukan")

	case errors.Is(err, repository.ErrCourseFull):
		return helper.NewError(
			fiber.StatusConflict,
			"COURSE_FULL",
			"kuota kelas sudah penuh",
		)

	case errors.Is(err, repository.ErrCourseDuplicate):
		return helper.NewError(
			fiber.StatusConflict,
			"ALREADY_ENROLLED",
			"Anda sudah terdaftar di kelas ini",
		)

	case errors.Is(err, repository.ErrNotEnrolled):
		return helper.NewError(
			fiber.StatusConflict,
			"NOT_ENROLLED",
			"Anda belum terdaftar di kelas ini",
		)

	case errors.Is(err, errCreditLimit):
		return helper.NewError(
			fiber.StatusConflict,
			"CREDIT_LIMIT_EXCEEDED",
			"batas "+strconv.Itoa(MaxCreditsPerSemester)+" SKS terlampaui",
		)

	default:
		return helper.Internal(err)
	}
}

// GET /courses
func (s *CourseService) List(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	courses, err := s.courses.FindAll(ctx)
	if err != nil {
		return helper.Internal(err)
	}

	return helper.Success(
		c,
		fiber.StatusOK,
		"daftar kelas berhasil diambil",
		courses,
	)
}

// GET /courses/:id
func (s *CourseService) Get(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id harus berupa angka positif")
	}

	course, err := s.courses.FindByID(ctx, id)
	if err != nil {
		return translateCourseError(err)
	}

	return helper.Success(
		c,
		fiber.StatusOK,
		"kelas berhasil ditemukan",
		course,
	)
}

// GET /courses/mine
func (s *CourseService) Mine(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	current, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Unauthorized("belum terautentikasi")
	}

	courses, err := s.courses.FindByStudent(ctx, current.UserID)
	if err != nil {
		return helper.Internal(err)
	}

	totalCredits := 0

	for _, course := range courses {
		totalCredits += course.Credits
	}

	return helper.Success(
		c,
		fiber.StatusOK,
		"KRS berhasil diambil",
		fiber.Map{
			"courses":       courses,
			"total_credits": totalCredits,
			"max_credits":   MaxCreditsPerSemester,
		},
	)
}

// POST /courses/:id/enroll
func (s *CourseService) Enroll(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	current, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Unauthorized("belum terautentikasi")
	}

	courseID, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id harus berupa angka positif")
	}

	var enrolled model.Course

	err := s.tx.WithTx(ctx, func(tx repository.DBTX) error {
		// Semua query harus menggunakan repository yang terikat
		// pada transaction ini.
		repo := s.courses.WithTx(tx)

		// 1. Kunci mahasiswa agar request KRS milik mahasiswa
		// yang sama berjalan secara bergantian.
		if err := repo.LockStudent(ctx, current.UserID); err != nil {
			return err
		}

		// 2. Periksa kelas dan batas SKS.
		course, err := repo.FindByID(ctx, courseID)
		if err != nil {
			return err
		}

		taken, err := repo.SumCredits(ctx, current.UserID)
		if err != nil {
			return err
		}

		if !WithinCreditLimit(taken, course.Credits) {
			return errCreditLimit
		}

		// 3. Ambil kursi secara atomik.
		enrolled, err = repo.IncrementEnrolled(ctx, courseID)
		if err != nil {
			return err
		}

		// 4. Catat enrollment.
		// Jika duplikat, seluruh transaction akan rollback,
		// termasuk penambahan enrolled_count.
		return repo.InsertEnrollment(
			ctx,
			courseID,
			current.UserID,
		)
	})

	if err != nil {
		return translateCourseError(err)
	}

	return helper.Created(
		c,
		"berhasil mendaftar kelas",
		enrolled,
		"/api/v1/courses/"+strconv.Itoa(courseID),
	)
}

// DELETE /courses/:id/enroll
func (s *CourseService) Cancel(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	current, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Unauthorized("belum terautentikasi")
	}

	courseID, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id harus berupa angka positif")
	}

	var updatedCourse model.Course

	err := s.tx.WithTx(ctx, func(tx repository.DBTX) error {
		repo := s.courses.WithTx(tx)

		// 1. Kunci mahasiswa.
		if err := repo.LockStudent(ctx, current.UserID); err != nil {
			return err
		}

		// 2. Kunci kelas.
		if err := repo.LockCourses(ctx, courseID); err != nil {
			return err
		}

		// 3. Hapus enrollment.
		if err := repo.DeleteEnrollment(
			ctx,
			courseID,
			current.UserID,
		); err != nil {
			return err
		}

		// 4. Kurangi jumlah peserta.
		var err error
		updatedCourse, err = repo.DecrementEnrolled(ctx, courseID)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return translateCourseError(err)
	}

	return helper.Success(
		c,
		fiber.StatusOK,
		"pendaftaran kelas berhasil dibatalkan",
		updatedCourse,
	)
}
