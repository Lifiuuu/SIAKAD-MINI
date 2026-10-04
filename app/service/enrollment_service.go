package service

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"siakad-mini/app/model"
	"siakad-mini/app/repository"
	"siakad-mini/helper"
)

type EnrollmentService struct {
	enrollments repository.EnrollmentRepository
	courses     repository.CourseRepository
	students    repository.StudentRepository
}

func NewEnrollmentService(e repository.EnrollmentRepository, c repository.CourseRepository, s repository.StudentRepository) *EnrollmentService {
	return &EnrollmentService{enrollments: e, courses: c, students: s}
}

func (s *EnrollmentService) Create(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	authUser, ok := helper.CurrentUser(c)
	if !ok || authUser.Role != "mahasiswa" {
		return helper.ErrorResponsePayload(c, fiber.StatusForbidden, "Forbidden", nil)
	}

	var req model.CreateEnrollmentRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.ErrorResponsePayload(c, fiber.StatusUnprocessableEntity, "Validasi gagal", map[string][]string{"body": {"format tidak valid"}})
	}

	// Validate req
	if req.CourseID == 0 || req.TahunAkademik == "" {
		return helper.ErrorResponsePayload(c, fiber.StatusUnprocessableEntity, "Validasi gagal", map[string][]string{"course_id_atau_tahun_akademik": {"Wajib diisi"}})
	}

	// Find student
	student, err := s.students.FindByUserID(ctx, authUser.UserID)
	if err != nil {
		return helper.Internal(err)
	}

	// Course exists? (locking will be done in Repo)
	course, err := s.courses.FindByID(ctx, req.CourseID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return helper.ErrorResponsePayload(c, fiber.StatusNotFound, "Mata kuliah tidak ditemukan", nil)
		}
		return helper.Internal(err)
	}

	en, err := s.enrollments.Create(ctx, student, course, req.TahunAkademik)
	if err != nil {
		if errors.Is(err, repository.ErrConflict) {
			return helper.ErrorResponsePayload(c, fiber.StatusUnprocessableEntity, "Batas SKS atau Kuota terlampaui", map[string][]string{"sisa_sks": {err.Error()}})
		}
		if errors.Is(err, repository.ErrDuplicate) {
			return helper.ErrorResponsePayload(c, fiber.StatusConflict, "Mata kuliah sudah diambil", nil)
		}
		return helper.Internal(err)
	}

	return helper.SuccessResponse(c, fiber.StatusCreated, "Berhasil mengambil mata kuliah", en, nil)
}

func (s *EnrollmentService) Delete(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	id, err := c.ParamsInt("id")
	if err != nil {
		return helper.ErrorResponsePayload(c, fiber.StatusNotFound, "Tidak ditemukan", nil)
	}

	authUser, ok := helper.CurrentUser(c)
	if !ok || authUser.Role != "mahasiswa" {
		return helper.ErrorResponsePayload(c, fiber.StatusForbidden, "Forbidden", nil)
	}

	student, err := s.students.FindByUserID(ctx, authUser.UserID)
	if err != nil {
		return helper.Internal(err)
	}

	err = s.enrollments.Delete(ctx, id, student.ID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return helper.ErrorResponsePayload(c, fiber.StatusNotFound, "KRS tidak ditemukan atau milik orang lain", nil)
		}
		return helper.Internal(err)
	}

	return c.SendStatus(fiber.StatusNoContent)
}
