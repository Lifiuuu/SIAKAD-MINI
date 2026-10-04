package service

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"
	"siakad-mini/app/repository"
	"siakad-mini/helper"
)

type AuthService struct {
	users    repository.UserRepository
	students repository.StudentRepository
	jwt      *helper.JWTManager
}

func NewAuthService(users repository.UserRepository, students repository.StudentRepository, jwtManager *helper.JWTManager) *AuthService {
	return &AuthService{
		users:    users,
		students: students,
		jwt:      jwtManager,
	}
}

func (s *AuthService) Login(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	var req struct {
		Email    string `json:"email" validate:"required,email"`
		Password string `json:"password" validate:"required,min=8"`
	}
	if err := c.BodyParser(&req); err != nil {
		return helper.ErrorResponsePayload(c, fiber.StatusUnprocessableEntity, "Validasi gagal", map[string][]string{"body": {"format tidak valid"}})
	}

	if errs := helper.ValidateStruct(req); len(errs) > 0 {
		errMap := make(map[string][]string)
		for k, v := range errs {
			errMap[k] = append(errMap[k], v)
		}
		return helper.ErrorResponsePayload(c, fiber.StatusUnprocessableEntity, "Validasi gagal", errMap)
	}

	user, err := s.users.FindByEmail(ctx, strings.TrimSpace(req.Email))
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			helper.VerifyDummyPassword(req.Password)
			return helper.ErrorResponsePayload(c, fiber.StatusUnauthorized, "Kredensial salah", nil)
		}
		return helper.Internal(err)
	}

	if !helper.VerifyPassword(user.Password, req.Password) {
		return helper.ErrorResponsePayload(c, fiber.StatusUnauthorized, "Kredensial salah", nil)
	}

	// For student role, check soft delete
	if user.Role == "mahasiswa" {
		_, err := s.students.FindByUserID(ctx, user.ID)
		if err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				return helper.ErrorResponsePayload(c, fiber.StatusUnauthorized, "Kredensial salah", nil)
			}
			return helper.Internal(err)
		}
	}

	accessToken, err := s.jwt.GenerateAccess(user)
	if err != nil {
		return helper.Internal(err)
	}

	data := fiber.Map{
		"access_token": accessToken,
		"token_type":   "Bearer",
		"expires_in":   int(s.jwt.AccessTTL().Seconds()),
		"user": fiber.Map{
			"id":    user.ID,
			"email": user.Email,
			"role":  user.Role,
		},
	}

	return helper.SuccessResponse(c, fiber.StatusOK, "Login berhasil", data, nil)
}

func (s *AuthService) Me(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	authUser, ok := helper.CurrentUser(c)
	if !ok {
		return helper.ErrorResponsePayload(c, fiber.StatusUnauthorized, "Unauthorized", nil)
	}

	user, err := s.users.FindByID(ctx, authUser.UserID)
	if err != nil {
		return helper.ErrorResponsePayload(c, fiber.StatusUnauthorized, "User tidak ditemukan", nil)
	}

	data := fiber.Map{
		"id":    user.ID,
		"email": user.Email,
		"role":  user.Role,
	}

	if user.Role == "mahasiswa" {
		student, err := s.students.FindByUserID(ctx, user.ID)
		if err == nil {
			data["students"] = fiber.Map{
				"nim":      student.NIM,
				"nama":     student.Nama,
				"prodi":    student.Prodi,
				"angkatan": student.Angkatan,
			}
		}
	}

	return helper.SuccessResponse(c, fiber.StatusOK, "Profil berhasil diambil", data, nil)
}
