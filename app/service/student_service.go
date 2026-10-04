package service

import (
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"siakad-mini/app/model"
	"siakad-mini/app/repository"
	"siakad-mini/helper"
)

type StudentService struct {
	repo        repository.StudentRepository
	enrollments repository.EnrollmentRepository
}

func NewStudentService(repo repository.StudentRepository, enrollments repository.EnrollmentRepository) *StudentService {
	return &StudentService{repo: repo, enrollments: enrollments}
}

func (s *StudentService) List(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	page, _ := strconv.Atoi(c.Query("page", "1"))
	if page < 1 {
		page = 1
	}
	perPage, _ := strconv.Atoi(c.Query("per_page", "10"))
	if perPage < 1 {
		perPage = 10
	}
	if perPage > 50 {
		perPage = 50
	}

	q := model.ListQuery{
		Page:     page,
		Limit:    perPage,
		Search:   c.Query("search", ""),
		Sort:     c.Query("sort", "id"),
		Prodi:    c.Query("prodi", ""),
	}

	// Parse angkatan
	if angkatanStr := c.Query("angkatan", ""); angkatanStr != "" {
		if angkatan, err := strconv.Atoi(angkatanStr); err == nil {
			q.Angkatan = angkatan
		}
	}

	if q.Sort == "-ipk_terakhir" {
		q.Sort = "ipk_terakhir"
		q.Order = "desc"
	}

	students, total, err := s.repo.FindAll(ctx, q)
	if err != nil {
		return helper.Internal(err)
	}

	lastPage := total / perPage
	if total%perPage > 0 {
		lastPage++
	}
	if lastPage == 0 {
		lastPage = 1
	}

	meta := model.Meta{
		CurrentPage: page,
		PerPage:     perPage,
		Total:       total,
		LastPage:    lastPage,
	}

	return helper.SuccessResponse(c, fiber.StatusOK, "Daftar mahasiswa", students, meta)
}

func (s *StudentService) Create(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	var req model.CreateStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.ErrorResponsePayload(c, fiber.StatusUnprocessableEntity, "Validasi gagal", map[string][]string{"body": {"format JSON salah"}})
	}

	if errs := helper.ValidateStruct(req); len(errs) > 0 {
		errMap := make(map[string][]string)
		for k, v := range errs {
			errMap[k] = append(errMap[k], v)
		}
		return helper.ErrorResponsePayload(c, fiber.StatusUnprocessableEntity, "Validasi gagal", errMap)
	}

	hashed, err := helper.HashPassword(req.NIM)
	if err != nil {
		return helper.Internal(err)
	}

	u := model.User{
		Email:    req.Email,
		Password: hashed,
		Role:     "mahasiswa",
	}

	ipk := 0.0
	if req.IPKTerakhir != nil {
		ipk = *req.IPKTerakhir
	}

	student := model.Student{
		NIM:         req.NIM,
		Nama:        req.Nama,
		Prodi:       req.Prodi,
		Angkatan:    req.Angkatan,
		IPKTerakhir: ipk,
	}

	created, err := s.repo.CreateWithUser(ctx, student, u)
	if err != nil {
		if errors.Is(err, repository.ErrDuplicate) {
			return helper.ErrorResponsePayload(c, fiber.StatusUnprocessableEntity, "Validasi gagal", map[string][]string{"nim_atau_email": {"Duplikat"}})
		}
		return helper.Internal(err)
	}

	return helper.SuccessResponse(c, fiber.StatusCreated, "Mahasiswa berhasil dibuat", created, nil)
}

func (s *StudentService) Get(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	id, err := c.ParamsInt("id")
	if err != nil {
		return helper.ErrorResponsePayload(c, fiber.StatusNotFound, "Tidak ditemukan", nil)
	}

	// Authorization check
	authUser, ok := helper.CurrentUser(c)
	if !ok {
		return helper.ErrorResponsePayload(c, fiber.StatusUnauthorized, "Unauthorized", nil)
	}
	if authUser.Role == "mahasiswa" {
		userStudent, err := s.repo.FindByUserID(ctx, authUser.UserID)
		if err != nil || userStudent.ID != id {
			return helper.ErrorResponsePayload(c, fiber.StatusForbidden, "Forbidden", nil)
		}
	}

	student, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return helper.ErrorResponsePayload(c, fiber.StatusNotFound, "Mahasiswa tidak ditemukan", nil)
		}
		return helper.Internal(err)
	}

	// Get all enrolled courses
	courses, err := s.enrollments.GetAllEnrolledCourses(ctx, student.ID)
	if err != nil {
		return helper.Internal(err)
	}

	// Calculate total SKS and batas SKS
	totalSKS := 0
	for _, c := range courses {
		totalSKS += c.SKS
	}

	var batasSKS int
	if student.IPKTerakhir >= 3.00 {
		batasSKS = 24
	} else if student.IPKTerakhir >= 2.50 {
		batasSKS = 21
	} else {
		batasSKS = 18
	}

	detail := model.StudentDetail{
		Student:  student,
		TotalSKS: totalSKS,
		BatasSKS: batasSKS,
		Courses:  courses,
	}

	return helper.SuccessResponse(c, fiber.StatusOK, "Detail mahasiswa", detail, nil)
}

func (s *StudentService) Replace(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	id, err := c.ParamsInt("id")
	if err != nil {
		return helper.ErrorResponsePayload(c, fiber.StatusNotFound, "Tidak ditemukan", nil)
	}

	var req model.ReplaceStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.ErrorResponsePayload(c, fiber.StatusUnprocessableEntity, "Validasi gagal", map[string][]string{"body": {"format JSON salah"}})
	}

	if errs := helper.ValidateStruct(req); len(errs) > 0 {
		errMap := make(map[string][]string)
		for k, v := range errs {
			errMap[k] = append(errMap[k], v)
		}
		return helper.ErrorResponsePayload(c, fiber.StatusUnprocessableEntity, "Validasi gagal", errMap)
	}

	student, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return helper.ErrorResponsePayload(c, fiber.StatusNotFound, "Mahasiswa tidak ditemukan", nil)
		}
		return helper.Internal(err)
	}

	student.Nama = req.Nama
	student.Prodi = req.Prodi
	student.Angkatan = req.Angkatan
	if req.IPKTerakhir != nil {
		student.IPKTerakhir = *req.IPKTerakhir
	}

	updated, err := s.repo.Update(ctx, student)
	if err != nil {
		return helper.Internal(err)
	}

	return helper.SuccessResponse(c, fiber.StatusOK, "Mahasiswa diperbarui", updated, nil)
}

func (s *StudentService) Delete(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	id, err := c.ParamsInt("id")
	if err != nil {
		return helper.ErrorResponsePayload(c, fiber.StatusNotFound, "Tidak ditemukan", nil)
	}

	err = s.repo.SoftDelete(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return helper.ErrorResponsePayload(c, fiber.StatusNotFound, "Mahasiswa tidak ditemukan", nil)
		}
		return helper.Internal(err)
	}

	return c.SendStatus(fiber.StatusNoContent)
}
