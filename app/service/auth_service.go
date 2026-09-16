package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/GanisSayogyo/tugas1-go-fiber/app/model"
	"github.com/GanisSayogyo/tugas1-go-fiber/app/repository"
	"github.com/GanisSayogyo/tugas1-go-fiber/helper"
)

var (
	ErrInvalidCredentials = errors.New("username atau password salah")
	ErrUserInactive       = errors.New("user tidak aktif")
	ErrUserDuplicate      = errors.New("username atau email sudah digunakan")
)

type AuthService struct {
	userRepo  *repository.UserRepository
	tokenRepo *repository.TokenRepository
}

func NewAuthService(
	userRepo *repository.UserRepository,
	tokenRepo *repository.TokenRepository,
) *AuthService {
	return &AuthService{
		userRepo:  userRepo,
		tokenRepo: tokenRepo,
	}
}

func (s *AuthService) Register(
	ctx context.Context,
	req model.RegisterRequest,
) (*model.User, error) {

	if err := ValidateRegister(req.Username, req.Email, req.Password); err != nil {
		return nil, err
	}

	hash, err := helper.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	user := &model.User{
		Username: strings.TrimSpace(req.Username),
		Email:    strings.TrimSpace(req.Email),
		Password: hash,
		Role:     "user",
		IsActive: true,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *AuthService) Login(
	ctx context.Context,
	req model.LoginRequest,
) (model.TokenPair, error) {

	user, err := s.userRepo.FindByUsername(
		ctx,
		strings.TrimSpace(req.Username),
	)

	if err != nil {
		return model.TokenPair{}, ErrInvalidCredentials
	}

	if !user.IsActive {
		return model.TokenPair{}, ErrUserInactive
	}

	if !helper.CheckPassword(req.Password, user.Password) {
		return model.TokenPair{}, ErrInvalidCredentials
	}

	return s.issueTokenPair(ctx, *user)
}

func (s *AuthService) Refresh(
	ctx context.Context,
	refreshToken string,
) (model.TokenPair, error) {

	hash := helper.HashToken(refreshToken)

	token, err := s.tokenRepo.FindActive(ctx, hash)
	if err != nil {
		return model.TokenPair{}, errors.New("refresh token tidak valid")
	}

	if err := s.tokenRepo.Revoke(ctx, hash); err != nil {
		return model.TokenPair{}, err
	}

	user, err := s.userRepo.FindByID(ctx, token.UserID)
	if err != nil {
		return model.TokenPair{}, err
	}

	if !user.IsActive {
		return model.TokenPair{}, ErrUserInactive
	}

	return s.issueTokenPair(ctx, *user)
}

func (s *AuthService) Logout(
	ctx context.Context,
	refreshToken string,
) error {

	hash := helper.HashToken(refreshToken)

	return s.tokenRepo.Revoke(ctx, hash)
}

func (s *AuthService) Me(
	ctx context.Context,
	userID int,
) (*model.User, error) {

	return s.userRepo.FindByID(ctx, userID)
}

func (s *AuthService) issueTokenPair(
	ctx context.Context,
	user model.User,
) (model.TokenPair, error) {

	authUser := model.AuthUser{
		UserID:   user.ID,
		Username: user.Username,
		Role:     user.Role,
	}

	accessToken, err := helper.GenerateAccessToken(authUser)
	if err != nil {
		return model.TokenPair{}, err
	}

	refreshToken, err := helper.GenerateRefreshToken()
	if err != nil {
		return model.TokenPair{}, err
	}

	refreshTTL := 7 * 24 * time.Hour

	token := model.RefreshToken{
		UserID:    user.ID,
		TokenHash: helper.HashToken(refreshToken),
		ExpiresAt: time.Now().Add(refreshTTL),
		CreatedAt: time.Now(),
	}

	if err := s.tokenRepo.Save(ctx, token); err != nil {
		return model.TokenPair{}, err
	}

	return model.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    900,
	}, nil
}
