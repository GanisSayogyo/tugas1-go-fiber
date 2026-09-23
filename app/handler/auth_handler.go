package handler

import (
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v2"

	"github.com/GanisSayogyo/tugas1-go-fiber/app/model"
	"github.com/GanisSayogyo/tugas1-go-fiber/app/service"
	"github.com/GanisSayogyo/tugas1-go-fiber/helper"
)

type AuthHandler struct {
	authService *service.AuthService
	permissions *helper.PermissionSet
}

func NewAuthHandler(
	authService *service.AuthService,
	permissions *helper.PermissionSet,
) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		permissions: permissions,
	}
}

func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var req model.RegisterRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "request tidak valid",
		})
	}

	user, err := h.authService.Register(c.Context(), req)
	if err != nil {
		if errors.Is(err, service.ErrUserDuplicate) {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"success": false,
				"message": err.Error(),
			})
		}

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "register berhasil",
		"user": fiber.Map{
			"id":         user.ID,
			"username":   user.Username,
			"email":      user.Email,
			"role":       user.Role,
			"is_active":  user.IsActive,
			"created_at": user.CreatedAt,
		},
	})
}

func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req model.LoginRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "request tidak valid",
		})
	}

	tokens, err := h.authService.Login(c.Context(), req)
	if err != nil {
		if errors.Is(err, service.ErrUserInactive) {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"message": "user tidak aktif",
			})
		}

		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "username atau password salah",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "login berhasil",
		"data":    tokens,
	})
}

func (h *AuthHandler) Refresh(c *fiber.Ctx) error {
	var req model.RefreshRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "request tidak valid",
		})
	}

	if req.RefreshToken == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "refresh token wajib diisi",
		})
	}

	tokens, err := h.authService.Refresh(
		c.Context(),
		req.RefreshToken,
	)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "refresh token tidak valid",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "refresh berhasil",
		"data":    tokens,
	})
}

func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	var req model.RefreshRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "request tidak valid",
		})
	}

	if req.RefreshToken == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "refresh token wajib diisi",
		})
	}

	if err := h.authService.Logout(
		c.Context(),
		req.RefreshToken,
	); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "gagal logout",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "logout berhasil",
	})
}

func (h *AuthHandler) Me(c *fiber.Ctx) error {
	userIDValue := c.Locals("user_id")

	userID, ok := userIDValue.(int)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "identity tidak valid",
		})
	}

	user, err := h.authService.Me(c.Context(), userID)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "user tidak ditemukan",
		})
	}

	permissions := h.permissions.PermissionsOf(user.Role)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"id":          user.ID,
			"username":    user.Username,
			"email":       user.Email,
			"role":        user.Role,
			"permissions": permissions,
			"is_active":   user.IsActive,
			"created_at":  user.CreatedAt,
		},
	})
}

func GetUserIDFromContext(c *fiber.Ctx) (int, error) {
	value := c.Locals("user_id")

	switch v := value.(type) {
	case int:
		return v, nil
	case string:
		return strconv.Atoi(v)
	default:
		return 0, errors.New("user id tidak ditemukan")
	}
}
