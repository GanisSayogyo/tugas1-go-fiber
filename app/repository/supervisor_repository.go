package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/GanisSayogyo/tugas1-go-fiber/app/model"
)

var (
	ErrSupervisorNotFound  = errors.New("pembimbing tidak ditemukan")
	ErrSupervisorFull      = errors.New("kuota pembimbing penuh")
	ErrStudentNotFound     = errors.New("mahasiswa tidak ditemukan")
	ErrAlreadySupervised   = errors.New("mahasiswa sudah memiliki pembimbing")
	ErrSupervisionNotFound = errors.New("data pembimbingan tidak ditemukan")
	ErrSupervisorVersion   = errors.New("versi pembimbing sudah berubah")
	ErrSupervisorQuota     = errors.New("kuota tidak boleh lebih kecil dari jumlah mahasiswa")
)

type SupervisorRepository interface {
	WithTx(DBTX) SupervisorRepository
	LockStudent(context.Context, int) error
	LockSupervisors(context.Context, ...int) error
	FindByID(context.Context, int) (model.Supervisor, error)
	FindSupervision(context.Context, int) (int, error)
	Assign(context.Context, int) (model.Supervisor, error)
	InsertSupervision(context.Context, int, int) error
	DeleteSupervision(context.Context, int) error
	Decrement(context.Context, int) error
	Update(context.Context, int, string, int, int) (model.Supervisor, error)
}

type supervisorPostgresRepository struct {
	db DBTX
}

func NewSupervisorRepository(db DBTX) SupervisorRepository {
	return &supervisorPostgresRepository{db: db}
}

func (r *supervisorPostgresRepository) WithTx(tx DBTX) SupervisorRepository {
	return &supervisorPostgresRepository{db: tx}
}

const supervisorColumns = `
	id, name, quota, supervised_count, version, created_at
`

func scanSupervisor(row pgx.Row) (model.Supervisor, error) {
	var s model.Supervisor
	err := row.Scan(
		&s.ID,
		&s.Name,
		&s.Quota,
		&s.SupervisedCount,
		&s.Version,
		&s.CreatedAt,
	)
	return s, err
}

func (r *supervisorPostgresRepository) LockStudent(
	ctx context.Context, id int,
) error {
	var lockedID int
	err := r.db.QueryRow(
		ctx,
		`SELECT id FROM students WHERE id = $1 FOR UPDATE`,
		id,
	).Scan(&lockedID)

	if errors.Is(err, pgx.ErrNoRows) {
		return ErrStudentNotFound
	}
	if err != nil {
		return fmt.Errorf("mengunci mahasiswa: %w", err)
	}
	return nil
}

func (r *supervisorPostgresRepository) LockSupervisors(
	ctx context.Context, ids ...int,
) error {
	rows, err := r.db.Query(
		ctx,
		`SELECT id FROM supervisors
		 WHERE id = ANY($1)
		 ORDER BY id
		 FOR UPDATE`,
		ids,
	)
	if err != nil {
		return fmt.Errorf("mengunci pembimbing: %w", err)
	}
	defer rows.Close()

	found := 0
	for rows.Next() {
		found++
	}
	if err := rows.Err(); err != nil {
		return err
	}
	if found != len(ids) {
		return ErrSupervisorNotFound
	}
	return nil
}

func (r *supervisorPostgresRepository) FindByID(
	ctx context.Context, id int,
) (model.Supervisor, error) {
	s, err := scanSupervisor(r.db.QueryRow(
		ctx,
		`SELECT `+supervisorColumns+` FROM supervisors WHERE id = $1`,
		id,
	))
	if errors.Is(err, pgx.ErrNoRows) {
		return model.Supervisor{}, ErrSupervisorNotFound
	}
	if err != nil {
		return model.Supervisor{}, fmt.Errorf("mengambil pembimbing: %w", err)
	}
	return s, nil
}

func (r *supervisorPostgresRepository) FindSupervision(
	ctx context.Context, studentID int,
) (int, error) {
	var supervisorID int
	err := r.db.QueryRow(
		ctx,
		`SELECT supervisor_id FROM supervisions WHERE student_id = $1`,
		studentID,
	).Scan(&supervisorID)

	if errors.Is(err, pgx.ErrNoRows) {
		return 0, ErrSupervisionNotFound
	}
	if err != nil {
		return 0, err
	}
	return supervisorID, nil
}

func (r *supervisorPostgresRepository) Assign(
	ctx context.Context, id int,
) (model.Supervisor, error) {
	s, err := scanSupervisor(r.db.QueryRow(
		ctx,
		`UPDATE supervisors
		 SET supervised_count = supervised_count + 1
		 WHERE id = $1 AND supervised_count < quota
		 RETURNING `+supervisorColumns,
		id,
	))
	if err == nil {
		return s, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return model.Supervisor{}, fmt.Errorf("menambah bimbingan: %w", err)
	}
	if _, findErr := r.FindByID(ctx, id); findErr != nil {
		return model.Supervisor{}, findErr
	}
	return model.Supervisor{}, ErrSupervisorFull
}

func (r *supervisorPostgresRepository) InsertSupervision(
	ctx context.Context, studentID, supervisorID int,
) error {
	_, err := r.db.Exec(
		ctx,
		`INSERT INTO supervisions (student_id, supervisor_id)
		 VALUES ($1, $2)`,
		studentID,
		supervisorID,
	)
	if pgErrorCode(err) == codeUniqueViolation {
		return ErrAlreadySupervised
	}
	if err != nil {
		return fmt.Errorf("menyimpan pembimbingan: %w", err)
	}
	return nil
}

func (r *supervisorPostgresRepository) DeleteSupervision(
	ctx context.Context, studentID int,
) error {
	tag, err := r.db.Exec(
		ctx,
		`DELETE FROM supervisions WHERE student_id = $1`,
		studentID,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrSupervisionNotFound
	}
	return nil
}

func (r *supervisorPostgresRepository) Decrement(
	ctx context.Context, id int,
) error {
	tag, err := r.db.Exec(
		ctx,
		`UPDATE supervisors
		 SET supervised_count = supervised_count - 1
		 WHERE id = $1 AND supervised_count > 0`,
		id,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrSupervisorNotFound
	}
	return nil
}

func (r *supervisorPostgresRepository) Update(
	ctx context.Context,
	id int,
	name string,
	quota int,
	version int,
) (model.Supervisor, error) {
	s, err := scanSupervisor(r.db.QueryRow(
		ctx,
		`UPDATE supervisors
		 SET name = $1,
		     quota = $2,
		     version = version + 1
		 WHERE id = $3
		   AND version = $4
		   AND quota >= supervised_count
		   AND $2 >= supervised_count
		 RETURNING `+supervisorColumns,
		name, quota, id, version,
	))
	if err == nil {
		return s, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return model.Supervisor{}, fmt.Errorf("mengubah pembimbing: %w", err)
	}

	current, findErr := r.FindByID(ctx, id)
	if findErr != nil {
		return model.Supervisor{}, findErr
	}
	if current.Version != version {
		return model.Supervisor{}, ErrSupervisorVersion
	}
	return model.Supervisor{}, ErrSupervisorQuota
}
