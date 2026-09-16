package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/GanisSayogyo/tugas1-go-fiber/app/model"
)

// TokenRepository menangani refresh token ke PostgreSQL.
type TokenRepository struct {
	db *pgxpool.Pool
}

// NewTokenRepository membuat repository refresh token baru.
func NewTokenRepository(db *pgxpool.Pool) *TokenRepository {
	return &TokenRepository{
		db: db,
	}
}

// Save menyimpan refresh token ke database.
func (r *TokenRepository) Save(
	ctx context.Context,
	token model.RefreshToken,
) error {

	_, err := r.db.Exec(
		ctx,
		`
		INSERT INTO refresh_tokens
			(user_id, token_hash, expires_at, revoked_at, created_at)
		VALUES
			($1, $2, $3, $4, $5)
		`,
		token.UserID,
		token.TokenHash,
		token.ExpiresAt,
		token.RevokedAt,
		token.CreatedAt,
	)

	return err
}

// FindActive mencari refresh token yang masih aktif.
func (r *TokenRepository) FindActive(
	ctx context.Context,
	tokenHash string,
) (model.RefreshToken, error) {

	var token model.RefreshToken

	err := r.db.QueryRow(
		ctx,
		`
		SELECT id, user_id, token_hash, expires_at, revoked_at, created_at
		FROM refresh_tokens
		WHERE token_hash = $1
		  AND revoked_at IS NULL
		  AND expires_at > NOW()
		`,
		tokenHash,
	).Scan(
		&token.ID,
		&token.UserID,
		&token.TokenHash,
		&token.ExpiresAt,
		&token.RevokedAt,
		&token.CreatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return model.RefreshToken{}, errors.New("refresh token tidak valid")
	}

	if err != nil {
		return model.RefreshToken{}, err
	}

	return token, nil
}

// Revoke mencabut satu refresh token.
func (r *TokenRepository) Revoke(
	ctx context.Context,
	tokenHash string,
) error {

	_, err := r.db.Exec(
		ctx,
		`
		UPDATE refresh_tokens
		SET revoked_at = NOW()
		WHERE token_hash = $1
		  AND revoked_at IS NULL
		`,
		tokenHash,
	)

	return err
}

// RevokeAllForUser mencabut semua refresh token milik user.
func (r *TokenRepository) RevokeAllForUser(
	ctx context.Context,
	userID int,
) error {

	_, err := r.db.Exec(
		ctx,
		`
		UPDATE refresh_tokens
		SET revoked_at = NOW()
		WHERE user_id = $1
		  AND revoked_at IS NULL
		`,
		userID,
	)

	return err
}