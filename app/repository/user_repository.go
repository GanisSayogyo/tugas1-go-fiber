package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/GanisSayogyo/tugas1-go-fiber/app/model"
)

var (
	ErrUserNotFound = errors.New("user tidak ditemukan")
	ErrUserDuplicate = errors.New("username atau email sudah digunakan")
)

// UserRepository menangani akses data user ke PostgreSQL.
type UserRepository struct {
	db *pgxpool.Pool
}

// NewUserRepository membuat repository user baru.
func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{
		db: db,
	}
}

// FindByID mencari user berdasarkan ID.
func (r *UserRepository) FindByID(
	ctx context.Context,
	id int,
) (*model.User, error) {

	var user model.User

	err := r.db.QueryRow(
		ctx,
		`
		SELECT id, username, email, password, role, is_active, created_at
		FROM users
		WHERE id = $1
		`,
		id,
	).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Password,
		&user.Role,
		&user.IsActive,
		&user.CreatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrUserNotFound
	}

	if err != nil {
		return nil, err
	}

	return &user, nil
}

// FindByUsername mencari user berdasarkan username secara case-insensitive.
func (r *UserRepository) FindByUsername(
	ctx context.Context,
	username string,
) (*model.User, error) {

	var user model.User

	err := r.db.QueryRow(
		ctx,
		`
		SELECT id, username, email, password, role, is_active, created_at
		FROM users
		WHERE LOWER(username) = LOWER($1)
		`,
		username,
	).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Password,
		&user.Role,
		&user.IsActive,
		&user.CreatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrUserNotFound
	}

	if err != nil {
		return nil, err
	}

	return &user, nil
}

// Create menyimpan user baru ke PostgreSQL.
func (r *UserRepository) Create(
	ctx context.Context,
	user *model.User,
) error {

	err := r.db.QueryRow(
		ctx,
		`
		INSERT INTO users (username, email, password, role, is_active)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at
		`,
		user.Username,
		user.Email,
		user.Password,
		user.Role,
		user.IsActive,
	).Scan(
		&user.ID,
		&user.CreatedAt,
	)

	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return ErrUserDuplicate
		}

		return err
	}

	return nil
}