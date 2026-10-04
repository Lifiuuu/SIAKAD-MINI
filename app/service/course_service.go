package service

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"siakad-mini/app/model"
	"siakad-mini/app/repository"
	"siakad-mini/helper"
)

type CourseService struct {
	repo repository.CourseRepository
}

func NewCourseService(repo repository.CourseRepository) *CourseService {
	return &CourseService{repo: repo}
}

func (s *CourseService) List(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	// Available=true filters only if requested? The spec says "available=true (hanya yang kuotanya belum penuh)".
	// In our repository, we can do this logic later, but for now we fetch all and let's just return all.

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
		Page:   page,
		Limit:  perPage,
		Search: c.Query("search", ""),
	}

	// Parse semester filter
	if semesterStr := c.Query("semester", ""); semesterStr != "" {
		if semester, err := strconv.Atoi(semesterStr); err == nil {
			q.Semester = semester
		}
	}

	courses, _, err := s.repo.FindAll(ctx, q)
	if err != nil {
		return helper.Internal(err)
	}
	
	// Apply "available" filter manually if requested
	if c.Query("available") == "true" {
		filtered := make([]model.Course, 0)
		for _, c := range courses {
			if c.SisaKuota > 0 {
				filtered = append(filtered, c)
			}
		}
		courses = filtered
	}

	return helper.SuccessResponse(c, fiber.StatusOK, "Daftar mata kuliah", courses, nil)
}
