package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/GanisSayogyo/tugas1-go-fiber/app/model"
)

// ErrNotFound digunakan ketika data student tidak ditemukan.
var ErrNotFound = errors.New("student tidak ditemukan")

// ErrDuplicate digunakan ketika NIM sudah digunakan.
var ErrDuplicate = errors.New("NIM sudah digunakan")

// StudentRepository menangani akses data student ke PostgreSQL.
type StudentRepository struct {
	db *pgxpool.Pool
}

// NewStudentRepository membuat repository student baru.
func NewStudentRepository(db *pgxpool.Pool) *StudentRepository {
	return &StudentRepository{
		db: db,
	}
}

// FindByID mencari satu student berdasarkan ID.
func (r *StudentRepository) FindByID(
	ctx context.Context,
	id int,
) (*model.Student, error) {
	var student model.Student

	err := r.db.QueryRow(
		ctx,
		`
		SELECT id, nim, name, grade, is_active, created_at
		FROM students
		WHERE id = $1
		`,
		id,
	).Scan(
		&student.ID,
		&student.NIM,
		&student.Name,
		&student.Grade,
		&student.IsActive,
		&student.CreatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}

	if err != nil {
		return nil, err
	}

	return &student, nil
}

// FindAll mengambil daftar student dengan filter,
// sorting, dan pagination dari PostgreSQL.
func (r *StudentRepository) FindAll(
	ctx context.Context,
	search string,
	isActive *bool,
	minGrade *float64,
	maxGrade *float64,
	sortField string,
	sortOrder string,
	limit int,
	offset int,
) ([]model.Student, int, error) {

	var total int

	err := r.db.QueryRow(
		ctx,
		`
		SELECT COUNT(*)
		FROM students
		WHERE
			($1 = '' OR nim ILIKE '%' || $1 || '%' OR name ILIKE '%' || $1 || '%')
			AND ($2::boolean IS NULL OR is_active = $2)
			AND ($3::numeric IS NULL OR grade >= $3)
			AND ($4::numeric IS NULL OR grade <= $4)
		`,
		search,
		isActive,
		minGrade,
		maxGrade,
	).Scan(&total)

	if err != nil {
		return nil, 0, err
	}

	query := `
		SELECT id, nim, name, grade, is_active, created_at
		FROM students
		WHERE
			($1 = '' OR nim ILIKE '%' || $1 || '%' OR name ILIKE '%' || $1 || '%')
			AND ($2::boolean IS NULL OR is_active = $2)
			AND ($3::numeric IS NULL OR grade >= $3)
			AND ($4::numeric IS NULL OR grade <= $4)
		ORDER BY ` + sortField + ` ` + sortOrder + `
		LIMIT $5 OFFSET $6
	`

	rows, err := r.db.Query(
		ctx,
		query,
		search,
		isActive,
		minGrade,
		maxGrade,
		limit,
		offset,
	)

	if err != nil {
		return nil, 0, err
	}

	defer rows.Close()

	students := make([]model.Student, 0)

	for rows.Next() {
		var student model.Student

		err := rows.Scan(
			&student.ID,
			&student.NIM,
			&student.Name,
			&student.Grade,
			&student.IsActive,
			&student.CreatedAt,
		)

		if err != nil {
			return nil, 0, err
		}

		students = append(students, student)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return students, total, nil
}

// Create menyimpan student baru ke PostgreSQL.
func (r *StudentRepository) Create(
	ctx context.Context,
	student *model.Student,
) error {

	err := r.db.QueryRow(
		ctx,
		`
		INSERT INTO students (nim, name, grade, is_active)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at
		`,
		student.NIM,
		student.Name,
		student.Grade,
		student.IsActive,
	).Scan(
		&student.ID,
		&student.CreatedAt,
	)

	if err != nil {
		return err
	}

	return nil
}

// Update mengganti seluruh data student berdasarkan ID.
func (r *StudentRepository) Update(
	ctx context.Context,
	student *model.Student,
) error {

	result, err := r.db.Exec(
		ctx,
		`
		UPDATE students
		SET
			nim = $1,
			name = $2,
			grade = $3,
			is_active = $4
		WHERE id = $5
		`,
		student.NIM,
		student.Name,
		student.Grade,
		student.IsActive,
		student.ID,
	)

	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

// Patch memperbarui sebagian field student berdasarkan ID.
func (r *StudentRepository) Patch(
	ctx context.Context,
	id int,
	nim *string,
	name *string,
	grade *float64,
	isActive *bool,
) (*model.Student, error) {

	student, err := r.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if nim != nil {
		student.NIM = *nim
	}

	if name != nil {
		student.Name = *name
	}

	if grade != nil {
		student.Grade = *grade
	}

	if isActive != nil {
		student.IsActive = *isActive
	}

	err = r.Update(ctx, student)
	if err != nil {
		return nil, err
	}

	return r.FindByID(ctx, id)
}

// Delete menghapus student berdasarkan ID.
func (r *StudentRepository) Delete(
	ctx context.Context,
	id int,
) error {

	result, err := r.db.Exec(
		ctx,
		`
		DELETE FROM students
		WHERE id = $1
		`,
		id,
	)

	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}