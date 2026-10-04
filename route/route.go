package route

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"siakad-mini/app/service"
	"siakad-mini/helper"
	"siakad-mini/middleware"
)

type Dependencies struct {
	Pool              *pgxpool.Pool
	JWT               *helper.JWTManager
	AuthService       *service.AuthService
	StudentService    *service.StudentService
	CourseService     *service.CourseService
	EnrollmentService *service.EnrollmentService
}

func Register(app *fiber.App, deps Dependencies) {
	api := app.Group("/api/v1")

	// 1. POST /api/v1/auth/login
	api.Post("/auth/login", middleware.RequireJSON, middleware.LoginRateLimiter(), deps.AuthService.Login)

	// All other endpoints require authentication
	auth := api.Group("", middleware.RequireAuth(deps.JWT))

	// 2. GET /api/v1/auth/me
	auth.Get("/auth/me", deps.AuthService.Me)

	// Admin endpoints for students
	adminStudents := auth.Group("/students", middleware.RequireRole("admin"))
	// 3. GET /api/v1/students
	adminStudents.Get("/", deps.StudentService.List)
	// 4. POST /api/v1/students
	adminStudents.Post("/", middleware.RequireJSON, deps.StudentService.Create)
	// 6. PUT /api/v1/students/{id}
	adminStudents.Put("/:id", middleware.RequireJSON, deps.StudentService.Replace)
	// 7. DELETE /api/v1/students/{id}
	adminStudents.Delete("/:id", deps.StudentService.Delete)

	// 5. GET /api/v1/students/{id} (Admin or self)
	auth.Get("/students/:id", deps.StudentService.Get)

	// 8. GET /api/v1/courses (All roles)
	auth.Get("/courses", deps.CourseService.List)

	// Enrollments
	// 9. POST /api/v1/enrollments (Mahasiswa)
	auth.Post("/enrollments", middleware.RequireJSON, deps.EnrollmentService.Create)
	// 10. DELETE /api/v1/enrollments/{id} (Mahasiswa)
	auth.Delete("/enrollments/:id", deps.EnrollmentService.Delete)

	api.Get("/health", healthCheck(deps.Pool))
}

func healthCheck(pool *pgxpool.Pool) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(c.UserContext(), 2*time.Second)
		defer cancel()

		if err := pool.Ping(ctx); err != nil {
			return helper.ServiceUnavailable("database tidak dapat dihubungi")
		}
		return helper.SuccessResponse(c, fiber.StatusOK, "server dan database berjalan", nil, nil)
	}
}
