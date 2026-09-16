package helper

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/GanisSayogyo/tugas1-go-fiber/config"
	"github.com/GanisSayogyo/tugas1-go-fiber/app/model"
)

func GenerateAccessToken(user model.AuthUser) (string, error) {
	secret := config.GetEnv("JWT_SECRET", "")

	if len(secret) < 32 {
		return "", errors.New("JWT_SECRET harus minimal 32 karakter")
	}

	issuer := config.GetEnv("JWT_ISSUER", "praktikum-backend")
	ttlMinutes := config.GetEnvInt("JWT_ACCESS_TTL_MINUTES", 15)

	now := time.Now()
	expiresAt := now.Add(time.Duration(ttlMinutes) * time.Minute)

	claims := jwt.MapClaims{
		"sub":      fmt.Sprintf("%d", user.UserID),
		"username": user.Username,
		"role":     user.Role,
		"iss":      issuer,
		"iat":      now.Unix(),
		"exp":      expiresAt.Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(secret))
}

func ParseAccessToken(tokenString string) (model.AuthUser, error) {
	secret := config.GetEnv("JWT_SECRET", "")

	if len(secret) < 32 {
		return model.AuthUser{}, errors.New("JWT_SECRET harus minimal 32 karakter")
	}

	issuer := config.GetEnv("JWT_ISSUER", "praktikum-backend")

	token, err := jwt.Parse(
		tokenString,
		func(token *jwt.Token) (interface{}, error) {
			if token.Method != jwt.SigningMethodHS256 {
				return nil, errors.New("algoritma JWT tidak valid")
			}

			return []byte(secret), nil
		},
		jwt.WithIssuer(issuer),
	)

	if err != nil {
		return model.AuthUser{}, err
	}

	if !token.Valid {
		return model.AuthUser{}, errors.New("token tidak valid")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return model.AuthUser{}, errors.New("claims tidak valid")
	}

	sub, ok := claims["sub"].(string)
	if !ok {
		return model.AuthUser{}, errors.New("subject tidak valid")
	}

	var userID int
	if _, err := fmt.Sscanf(sub, "%d", &userID); err != nil {
		return model.AuthUser{}, errors.New("subject tidak valid")
	}

	username, ok := claims["username"].(string)
	if !ok {
		return model.AuthUser{}, errors.New("username tidak valid")
	}

	role, ok := claims["role"].(string)
	if !ok {
		return model.AuthUser{}, errors.New("role tidak valid")
	}

	return model.AuthUser{
		UserID:   userID,
		Username: username,
		Role:     role,
	}, nil
}