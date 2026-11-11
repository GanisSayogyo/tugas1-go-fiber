package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/GanisSayogyo/tugas1-go-fiber/app/model"
)

var (
	ErrCourseNotFound   = errors.New("kelas tidak ditemukan")
	ErrCourseFull       = errors.New("kuota kelas sudah penuh")
	ErrCourseDuplicate  = errors.New("sudah terdaftar di kelas ini")
	ErrNotEnrolled      = errors.New("tidak terdaftar di kelas ini")
	ErrVersionConflict  = errors.New("data sudah diubah pihak lain")
	ErrQuotaBelowFilled = errors.New("kuota lebih kecil dari jumlah peserta")
)

type CourseRepository interface {
	WithTx(tx DBTX) CourseRepository

	FindAll(ctx context.Context) ([]model.Course, error)
	FindByID(ctx context.Context, id int) (model.Course, error)
	FindByStudent(ctx context.Context, userID int) ([]model.Course, error)

	LockStudent(ctx context.Context, userID int) error
	LockCourses(ctx context.Context, ids ...int) error
	SumCredits(ctx context.Context, userID int) (int, error)

	InsertEnrollment(ctx context.Context, courseID, userID int) error
	DeleteEnrollment(ctx context.Context, courseID, userID int) error

	IncrementEnrolled(ctx context.Context, courseID int) (model.Course, error)
	DecrementEnrolled(ctx context.Context, courseID int) (model.Course, error)

	UpdateQuota(ctx context.Context, id, quota, version int) (model.Course, error)
}

const courseColumns = `
	id, code, name, credits, quota,
	enrolled_count, version, created_at
`

type coursePostgresRepository struct {
	db DBTX
}

// NewCourseRepository menerima pool atau DBTX lainnya.
func NewCourseRepository(db DBTX) CourseRepository {
	return &coursePostgresRepository{db: db}
}

// WithTx membuat repository baru yang menggunakan transaction.
func (r *coursePostgresRepository) WithTx(tx DBTX) CourseRepository {
	return &coursePostgresRepository{db: tx}
}

// scanCourse membaca satu row menjadi model.Course.
func scanCourse(row pgx.Row) (model.Course, error) {
	var course model.Course

	err := row.Scan(
		&course.ID,
		&course.Code,
		&course.Name,
		&course.Credits,
		&course.Quota,
		&course.EnrolledCount,
		&course.Version,
		&course.CreatedAt,
	)

	return course, err
}

// collect membaca sekumpulan row menjadi slice Course.
func (r *coursePostgresRepository) collect(
	rows pgx.Rows,
	err error,
) ([]model.Course, error) {
	if err != nil {
		return nil, fmt.Errorf("mengambil daftar kelas: %w", err)
	}

	defer rows.Close()

	result := make([]model.Course, 0)

	for rows.Next() {
		course, err := scanCourse(rows)
		if err != nil {
			return nil, fmt.Errorf("membaca data kelas: %w", err)
		}

		result = append(result, course)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("membaca hasil query kelas: %w", err)
	}

	return result, nil
}

// FindAll mengambil seluruh kelas.
func (r *coursePostgresRepository) FindAll(
	ctx context.Context,
) ([]model.Course, error) {
	return r.collect(
		r.db.Query(
			ctx,
			"SELECT "+courseColumns+" FROM courses ORDER BY code",
		),
	)
}

// FindByID mengambil kelas berdasarkan ID.
func (r *coursePostgresRepository) FindByID(
	ctx context.Context,
	id int,
) (model.Course, error) {
	course, err := scanCourse(
		r.db.QueryRow(
			ctx,
			"SELECT "+courseColumns+" FROM courses WHERE id = $1",
			id,
		),
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Course{}, ErrCourseNotFound
		}

		return model.Course{}, fmt.Errorf("mengambil kelas: %w", err)
	}

	return course, nil
}

// FindByStudent mengambil kelas yang diikuti mahasiswa.
func (r *coursePostgresRepository) FindByStudent(
	ctx context.Context,
	userID int,
) ([]model.Course, error) {
	query := `
		SELECT
			c.id,
			c.code,
			c.name,
			c.credits,
			c.quota,
			c.enrolled_count,
			c.version,
			c.created_at
		FROM courses c
		JOIN enrollments e ON e.course_id = c.id
		WHERE e.user_id = $1
		ORDER BY c.code
	`

	return r.collect(r.db.Query(ctx, query, userID))
}

func (r *coursePostgresRepository) LockStudent(
	ctx context.Context,
	userID int,
) error {
	var id int

	err := r.db.QueryRow(
		ctx,
		"SELECT id FROM users WHERE id = $1 FOR UPDATE",
		userID,
	).Scan(&id)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrCourseNotFound
		}
		return fmt.Errorf("mengunci mahasiswa: %w", err)
	}

	return nil
}

func (r *coursePostgresRepository) LockCourses(
	ctx context.Context,
	ids ...int,
) error {
	rows, err := r.db.Query(
		ctx,
		"SELECT id FROM courses WHERE id = ANY($1) ORDER BY id FOR UPDATE",
		ids,
	)
	if err != nil {
		return fmt.Errorf("mengunci kelas: %w", err)
	}
	defer rows.Close()

	found := 0

	for rows.Next() {
		found++
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("membaca kunci kelas: %w", err)
	}

	if found != len(ids) {
		return ErrCourseNotFound
	}

	return nil
}

func (r *coursePostgresRepository) SumCredits(
	ctx context.Context,
	userID int,
) (int, error) {
	var total int

	err := r.db.QueryRow(
		ctx,
		`SELECT COALESCE(SUM(c.credits), 0)
		 FROM enrollments e
		 JOIN courses c ON c.id = e.course_id
		 WHERE e.user_id = $1`,
		userID,
	).Scan(&total)

	if err != nil {
		return 0, fmt.Errorf("menghitung SKS: %w", err)
	}

	return total, nil
}

func (r *coursePostgresRepository) InsertEnrollment(
	ctx context.Context,
	courseID, userID int,
) error {
	_, err := r.db.Exec(
		ctx,
		"INSERT INTO enrollments (course_id, user_id) VALUES ($1, $2)",
		courseID,
		userID,
	)

	if err != nil {
		if pgErrorCode(err) == codeUniqueViolation {
			return ErrCourseDuplicate
		}
		return fmt.Errorf("mencatat pendaftaran: %w", err)
	}

	return nil
}

func (r *coursePostgresRepository) DeleteEnrollment(
	ctx context.Context,
	courseID, userID int,
) error {
	tag, err := r.db.Exec(
		ctx,
		"DELETE FROM enrollments WHERE course_id = $1 AND user_id = $2",
		courseID,
		userID,
	)

	if err != nil {
		return fmt.Errorf("menghapus pendaftaran: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return ErrNotEnrolled
	}

	return nil
}

func (r *coursePostgresRepository) IncrementEnrolled(
	ctx context.Context,
	courseID int,
) (model.Course, error) {
	course, err := scanCourse(
		r.db.QueryRow(
			ctx,
			`UPDATE courses
			 SET enrolled_count = enrolled_count + 1
			 WHERE id = $1 AND enrolled_count < quota
			 RETURNING `+courseColumns,
			courseID,
		),
	)

	if err == nil {
		return course, nil
	}

	if !errors.Is(err, pgx.ErrNoRows) {
		return model.Course{}, fmt.Errorf("menambah peserta: %w", err)
	}

	if _, findErr := r.FindByID(ctx, courseID); findErr != nil {
		return model.Course{}, findErr
	}

	return model.Course{}, ErrCourseFull
}

func (r *coursePostgresRepository) DecrementEnrolled(
	ctx context.Context,
	courseID int,
) (model.Course, error) {
	course, err := scanCourse(
		r.db.QueryRow(
			ctx,
			`UPDATE courses
			 SET enrolled_count = enrolled_count - 1
			 WHERE id = $1 AND enrolled_count > 0
			 RETURNING `+courseColumns,
			courseID,
		),
	)

	if err == nil {
		return course, nil
	}

	if errors.Is(err, pgx.ErrNoRows) {
		if _, findErr := r.FindByID(ctx, courseID); findErr != nil {
			return model.Course{}, findErr
		}
		return model.Course{}, ErrNotEnrolled
	}

	return model.Course{}, fmt.Errorf("mengurangi peserta: %w", err)
}

func (r *coursePostgresRepository) UpdateQuota(
	ctx context.Context,
	id, quota, version int,
) (model.Course, error) {
	course, err := scanCourse(
		r.db.QueryRow(
			ctx,
			`UPDATE courses
			 SET quota = $1, version = version + 1
			 WHERE id = $2
			   AND version = $3
			   AND $1 >= enrolled_count
			 RETURNING `+courseColumns,
			quota,
			id,
			version,
		),
	)

	if err == nil {
		return course, nil
	}

	if !errors.Is(err, pgx.ErrNoRows) {
		return model.Course{}, fmt.Errorf("mengubah kuota: %w", err)
	}

	current, findErr := r.FindByID(ctx, id)
	if findErr != nil {
		return model.Course{}, findErr
	}

	if current.Version != version {
		return model.Course{}, ErrVersionConflict
	}

	return model.Course{}, ErrQuotaBelowFilled
}
